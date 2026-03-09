package calendar

import (
	"fmt"
	"time"
)

const (
	WeeklySlotType SlotType = "weekly"
	DailySlotType  SlotType = "daily"
)

type SlotType string

type Event struct {
	summary     string
	description *string
	location    *string
	start       time.Time
	end         time.Time
	link        string
	slotType    SlotType
}

func (e Event) Summary() string {
	return fmt.Sprintf("*%s*", e.summary)
}

func (e Event) Description() *string {
	return e.description
}

func (e Event) Location() *string {
	return e.location
}

func (e Event) Start() string {
	switch e.slotType {
	case WeeklySlotType:
		return e.start.Weekday().String() + " " + e.start.Format("15:04")
	default:
		return e.start.Format("15:04")
	}
}

func (e Event) End() time.Time {
	return e.end
}

func (e Event) Link() string {
	return e.link
}

func NewEvent(
	summary string,
	description *string,
	location *string,
	start time.Time,
	end time.Time,
	link string,
	slotType SlotType,
) *Event {
	return &Event{
		summary:     summary,
		description: description,
		location:    location,
		start:       start,
		end:         end,
		link:        link,
		slotType:    slotType,
	}
}
