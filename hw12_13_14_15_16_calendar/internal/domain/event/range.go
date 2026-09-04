package event

import (
	"time"
)

// Range - интервал для выборки событий.
type Range struct {
	From time.Time
	To   time.Time
}

// DayRange возвращает диапазон суток для date.
func DayRange(date time.Time) *Range {
	from := startOfDay(date)

	return &Range{
		From: from,
		To:   from.AddDate(0, 0, 1),
	}
}

// WeekRange возвращает диапазон недели, если date понедельник.
func WeekRange(date time.Time) (*Range, error) {
	if err := IsFirstDayOfWeek(date); err != nil {
		return nil, err
	}

	from := startOfDay(date)

	return &Range{
		From: from,
		To:   from.AddDate(0, 0, 7),
	}, nil
}

// MonthRange возвращает диапазон месяца, если date первое число.
func MonthRange(date time.Time) (*Range, error) {
	if err := IsFirstDayOfMonth(date); err != nil {
		return nil, err
	}

	from := startOfDay(date)

	return &Range{
		From: from,
		To:   from.AddDate(0, 1, 0),
	}, nil
}

func startOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}
