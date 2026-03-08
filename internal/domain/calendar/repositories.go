package calendar

import (
	"context"
	"time"
)

type EventsRepository interface {
	GetEvents(ctx context.Context, from, to time.Time, accessToken, tokenType string, eventType EventType) ([]Event, error)
}

type EventsPublisher interface {
	Publish(ctx context.Context, message string) error
}
