package main

import (
	"time"
)

type Date struct{}

func (Date) Draw(x int) {
	now := time.Now()
	text := now.Format("02.01")
	Draw(x, "calendar", 0, weekdayIcon(now), text)
}

func (Date) Width() int {
	return Width("calendar", Icon("calendar"), "02.01")
}

func weekdayIcon(t time.Time) Icon {
	switch t.Weekday() {
	case time.Sunday:
		return "sunday"
	case time.Monday:
		return "monday"
	case time.Tuesday:
		return "tuesday"
	case time.Wednesday:
		return "wednesday"
	case time.Thursday:
		return "thursday"
	case time.Friday:
		return "friday"
	case time.Saturday:
		return "saturday"
	}
	return "calendar"
}
