package calendar

import "time"

type Task struct {
	title    string
	due      *time.Time
	link     string
	slotType SlotType
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Due() string {
	var due string
	if t.due != nil {
		due = t.due.Format("15:04")
	}
	switch t.slotType {
	case WeeklySlotType:
		return t.due.Weekday().String() + " " + due
	default:
		return due
	}
}

func (t Task) Link() string {
	return t.link
}

func NewTask(title string, due *time.Time, link string, slotType SlotType) *Task {
	return &Task{title: title, due: due, link: link, slotType: slotType}
}
