package google

import (
	"context"
	"fmt"
	"time"

	"github.com/bruli/myCalendar/internal/domain/calendar"
	"golang.org/x/oauth2"
	gcalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type EventsRepository struct{}

func (e EventsRepository) GetEvents(ctx context.Context, from, to time.Time, accessToken, tokenType string, eventType calendar.SlotType) ([]calendar.Event, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
		TokenType:   tokenType,
	})
	client := oauth2.NewClient(ctx, ts)
	svc, err := gcalendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar client: %s", err.Error())
	}
	resp, err := svc.Events.List("primary").
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %s", err.Error())
	}
	events := make([]calendar.Event, len(resp.Items))
	for i, ev := range resp.Items {
		var (
			description *string
			location    *string
		)
		if ev.Description != "" {
			description = &ev.Description
		}
		if ev.Location != "" {
			location = &ev.Location
		}
		start, _ := time.Parse(time.RFC3339, ev.Start.DateTime)
		end, _ := time.Parse(time.RFC3339, ev.End.DateTime)
		events[i] = *calendar.NewEvent(ev.Summary, description, location, start, end, ev.HtmlLink, eventType)
	}
	return events, nil
}

func NewEventsRepository() *EventsRepository {
	return &EventsRepository{}
}
