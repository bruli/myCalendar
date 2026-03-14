package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/bruli/myCalendar/internal/domain/auth"
)

type GetEvents struct {
	eventsRepo EventsRepository
	eventsPub  Publisher
	authRepo   auth.AuthenticationRepository
}

func (e GetEvents) Get(ctx context.Context, from, to time.Time, messageTitle string, eventType SlotType) error {
	tokenstr, err := e.authRepo.Read(ctx)
	if err != nil {
		return err
	}
	accessToken := tokenstr.AccessToken()
	tokenType := tokenstr.TokenType()
	evnts, err := e.eventsRepo.GetEvents(ctx, from, to, accessToken, tokenType, eventType)
	if err != nil {
		return err
	}
	if len(evnts) == 0 {
		messageTitle = "📅 No events found"
	}
	if err := e.eventsPub.Publish(ctx, messageTitle); err != nil {
		return err
	}
	for _, evnt := range evnts {
		msg := fmt.Sprintf("%s\nStart at: %s\nLink -> %s", evnt.Summary(), evnt.Start(), evnt.Link())
		if err = e.eventsPub.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func NewGetEvents(eventsRepo EventsRepository, eventsPub Publisher, authRepo auth.AuthenticationRepository) *GetEvents {
	return &GetEvents{eventsRepo: eventsRepo, eventsPub: eventsPub, authRepo: authRepo}
}
