package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/bruli/myCalendar/internal/domain/auth"
)

type GetTasks struct {
	tasksRepo TasksRepository
	publisher Publisher
	authRepo  auth.AuthenticationRepository
}

func (e GetTasks) Get(ctx context.Context, from, to time.Time, messageTitle string, eventType SlotType) error {
	tokenstr, err := e.authRepo.Read(ctx)
	if err != nil {
		return err
	}
	accessToken := tokenstr.AccessToken()
	tokenType := tokenstr.TokenType()
	tasks, err := e.tasksRepo.GetTasks(ctx, from, to, accessToken, tokenType, eventType)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		messageTitle = "✅ No tasks found"
	}
	if err = e.publisher.Publish(ctx, messageTitle); err != nil {
		return err
	}
	for _, evnt := range tasks {
		msg := fmt.Sprintf("%s\nDue at: %s\nLink -> %s", evnt.Title(), evnt.Due(), evnt.Link())
		if err = e.publisher.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func NewGetTasks(taskRepo TasksRepository, eventsPub Publisher, authRepo auth.AuthenticationRepository) *GetTasks {
	return &GetTasks{tasksRepo: taskRepo, publisher: eventsPub, authRepo: authRepo}
}
