package calendar

import (
	"context"
	"time"
)

type EventsRepository interface {
	GetEvents(ctx context.Context, from, to time.Time, accessToken, tokenType string, eventType SlotType) ([]Event, error)
}
type TasksRepository interface {
	GetTasks(ctx context.Context, from, to time.Time, accessToken, tokenType string, eventType SlotType) ([]Task, error)
}

type Publisher interface {
	Publish(ctx context.Context, message string) error
}
