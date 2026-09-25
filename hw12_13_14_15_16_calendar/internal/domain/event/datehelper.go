package event

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrFirstDayOfWeek  = errors.New("not first day of the week")
	ErrFirstDayOfMonth = errors.New("not first day of the month")
)

func IsFirstDayOfMonth(date time.Time) error {
	if date.Day() == 1 {
		return nil
	}

	return fmt.Errorf("%w: %s", ErrFirstDayOfMonth, date.Format("2006-01-02"))
}

func IsFirstDayOfWeek(date time.Time) error {
	if date.Weekday() == time.Monday {
		return nil
	}

	return fmt.Errorf("%w: %s", ErrFirstDayOfWeek, date.Format("2006-01-02"))
}
