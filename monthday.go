package heijitu

import "time"

type MonthDay struct {
	Month time.Month `yaml:"month" json:"month"`
	Day   int        `yaml:"day" json:"day"`
}

func (md MonthDay) Matches(t time.Time) bool {
	return t.Month() == md.Month && t.Day() == md.Day
}
