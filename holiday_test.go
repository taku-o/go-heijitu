package heijitu_test

import (
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
)

func TestHoliday_Fields(t *testing.T) {
	date := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	name := "元日"

	h := heijitu.Holiday{
		Date: date,
		Name: name,
	}

	if !h.Date.Equal(date) {
		t.Errorf("Holiday.Date = %v, want %v", h.Date, date)
	}
	if h.Name != name {
		t.Errorf("Holiday.Name = %q, want %q", h.Name, name)
	}
}
