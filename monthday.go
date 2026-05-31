package heijitu

import "time"

type MonthDay struct {
	Month time.Month
	Day   int
}

func (md MonthDay) Matches(t time.Time) bool {
	return t.Month() == md.Month && t.Day() == md.Day
}
