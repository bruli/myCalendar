package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bruli/myCalendar/internal/config"
	"github.com/bruli/myCalendar/internal/domain/auth"
	"github.com/bruli/myCalendar/internal/domain/calendar"
	telegram "github.com/bruli/myCalendar/internal/infra/Telegram"
	"github.com/bruli/myCalendar/internal/infra/disk"
	googleinfra "github.com/bruli/myCalendar/internal/infra/google"
	httpinfra "github.com/bruli/myCalendar/internal/infra/http"
	"github.com/robfig/cron/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	log := buildLog()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf, err := config.New()
	if err != nil {
		log.ErrorContext(ctx, "Error loading config", "err", err)
		os.Exit(1)
	}
	serverListener, err := net.Listen("tcp", conf.ServerHost)
	log.InfoContext(ctx, "Starting server", "host", conf.ServerHost)
	if err != nil {
		log.ErrorContext(ctx, "Error starting server", "err", err)
		os.Exit(1)
	}
	defer func() {
		_ = serverListener.Close()
	}()

	srv := httpinfra.NewServer(conf.ServerHost)
	defer func() {
		log.InfoContext(ctx, "Closing server")
		_ = srv.Shutdown(ctx)
	}()

	go runHTTPServer(ctx, srv, log, serverListener)

	cfg := buildOauthConfig(conf)

	authRepo := disk.NewAuthenticationRepository(conf.TokensFile)
	tokenRepo := googleinfra.NewTokenRepository(cfg)
	refreshToken := auth.NewRefreshToken(authRepo, tokenRepo)
	messagePublisher, err := telegram.NewMessagePublisher(conf.TelegramToken, conf.TelegramChatID)
	if err != nil {
		log.ErrorContext(ctx, "Error creating telegram publisher", "err", err)
		os.Exit(1)
	}

	if err = refreshToken.Refresh(ctx); err != nil {
		switch {
		case errors.As(err, &auth.RefreshError{}):
			log.WarnContext(ctx, "Error authenticating", "err", err)
			initializeCallback(ctx, log, tokenRepo, authRepo, cfg, messagePublisher, conf.CallbackHost)
		default:
			log.ErrorContext(ctx, "Error refreshing token on main", "err", err)
			os.Exit(1)
		}
	}

	eventsRepo := googleinfra.NewEventsRepository()
	tasksRepo := googleinfra.NewTasksRepository(log)
	getEventsSVC := calendar.NewGetEvents(eventsRepo, messagePublisher, authRepo)
	getTasksSVC := calendar.NewGetTasks(tasksRepo, messagePublisher, authRepo)

	c := jobs(ctx, log, refreshToken, getEventsSVC, getTasksSVC)

	c.Start()
	<-ctx.Done()
	shutdown(ctx, srv, log)
}

func jobs(
	ctx context.Context,
	log *slog.Logger,
	refreshToken *auth.RefreshToken,
	getEventsSVC *calendar.GetEvents,
	getTasksSVC *calendar.GetTasks,
) *cron.Cron {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.ErrorContext(ctx, "Error loading location", "err", err)
		os.Exit(1)
	}
	c := cron.New(cron.WithLocation(loc))
	log.InfoContext(ctx, "Running daily job")
	_, _ = c.AddFunc("0 8 * * 2-7", func() {
		start := time.Now().Truncate(24 * time.Hour)
		end := start.Add(24*time.Hour - time.Second)
		runGetEvents(ctx, refreshToken, log, getEventsSVC, "📅 Today's events\n────────────────", calendar.DailySlotType, start, end)
		runGetTasks(ctx, refreshToken, log, getTasksSVC, "✅ Today's tasks\n────────────────", calendar.DailySlotType, start, end)
	})
	log.InfoContext(ctx, "Running weekly job")
	_, _ = c.AddFunc("0 8 * * 1", func() {
		start := time.Now().Truncate(24 * time.Hour)
		end := start.AddDate(0, 0, 7)
		runGetEvents(ctx, refreshToken, log, getEventsSVC, "📅 Events for the week\n────────────────────", calendar.WeeklySlotType, start, end)
		runGetTasks(ctx, refreshToken, log, getTasksSVC, "✅ Tasks for the week\n────────────────────", calendar.WeeklySlotType, start, end)
	})
	return c
}

func runGetEvents(
	ctx context.Context,
	refreshToken *auth.RefreshToken,
	log *slog.Logger,
	getEventsSVC *calendar.GetEvents,
	title string,
	slotType calendar.SlotType,
	start, end time.Time,
) {
	if err := refreshToken.Refresh(ctx); err != nil {
		log.ErrorContext(ctx, "Error refreshing token on main", "err", err)
	}

	if err := getEventsSVC.Get(ctx, start, end, title, slotType); err != nil {
		log.ErrorContext(ctx, "Error getting events", "err", err)
	}
}
func runGetTasks(
	ctx context.Context,
	refreshToken *auth.RefreshToken,
	log *slog.Logger,
	getTasksSVC *calendar.GetTasks,
	title string,
	slotType calendar.SlotType,
	start, end time.Time,
) {
	if err := refreshToken.Refresh(ctx); err != nil {
		log.ErrorContext(ctx, "Error refreshing token on main", "err", err)
	}

	if err := getTasksSVC.Get(ctx, start, end, title, slotType); err != nil {
		log.ErrorContext(ctx, "Error getting tasks", "err", err)
	}
}

func runHTTPServer(ctx context.Context, srv *http.Server, log *slog.Logger, serverListener net.Listener) {
	go shutdown(ctx, srv, log)

	if err := srv.Serve(serverListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.ErrorContext(ctx, "Error starting server", "err", err)
		os.Exit(1)
	}
}

func shutdown(ctx context.Context, srv *http.Server, log *slog.Logger) {
	<-ctx.Done()
	log.InfoContext(ctx, "Ctrl+C received, shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("error shutting down server", "err", err)
	}
}

func initializeCallback(
	ctx context.Context,
	log *slog.Logger,
	tokenRepo *googleinfra.TokenRepository,
	authRepo *disk.AuthenticationRepository,
	cfg *oauth2.Config,
	publisher *telegram.MessagePublisher,
	callbackHost string,
) {
	log.InfoContext(ctx, "Initializing callback")
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	listener, err := net.Listen("tcp", callbackHost)
	log.InfoContext(ctx, "Starting callback", "host", callbackHost)
	if err != nil {
		log.ErrorContext(ctx, "failed listening:", "address", callbackHost, "err", err)
		os.Exit(1)
	}
	defer func() {
		_ = listener.Close()
	}()

	srv := httpinfra.NewCallbackServer(callbackHost, codeCh, errCh)
	defer func() {
		log.InfoContext(ctx, "Closing callback")
		_ = srv.Shutdown(ctx)
	}()

	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	buildAuthURL(ctx, cfg, publisher)

	createTokenSVC := auth.NewCreateTokens(tokenRepo, authRepo)

	var code string
	select {
	case code = <-codeCh:
		log.InfoContext(ctx, "Authorization received.")
		if err = createTokenSVC.Create(ctx, code); err != nil {
			log.ErrorContext(ctx, "Error creating tokens", "err", err)
			os.Exit(1)
		}
		log.InfoContext(ctx, "Tokens created.")
	case err := <-errCh:
		log.ErrorContext(ctx, "Error en OAuth", "err", err)
		os.Exit(1)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("error shutting down server", "err", err)
		}
	}
}

func buildAuthURL(ctx context.Context, cfg *oauth2.Config, publisher *telegram.MessagePublisher) {
	state := fmt.Sprintf("state-%d", time.Now().UnixNano())
	authURL := cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)

	_ = publisher.Publish(ctx, fmt.Sprintf("Click -> %s", authURL))
}

func buildOauthConfig(conf *config.Config) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.CallbackURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
			"https://www.googleapis.com/auth/tasks.readonly",
		},
		Endpoint: google.Endpoint,
	}
	return cfg
}

func buildLog() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler)
}
