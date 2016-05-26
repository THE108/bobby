package utils

import (
	"fmt"
	"time"
)

func GetPreviousDateRange(date time.Time) (time.Time, time.Time) {
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)
	to := from.Add(24 * time.Hour).Add(-1 * time.Second)
	return from, to
}

func GetDateFromArgs(arg string, now time.Time) (time.Time, error) {
	switch arg {
	case "now", "today":
		return now, nil
	case "yesterday":
		return now.Add(-24 * time.Hour), nil
	case "tomorrow":
		return now.Add(24 * time.Hour), nil
	}

	if t, err := time.ParseInLocation("2006-01-02", arg, time.Local); err == nil {
		return t, nil
	}

	return now, fmt.Errorf("Unknown date format: %q\n", arg)
}
