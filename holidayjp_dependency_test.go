package heijitu_test

import (
	"testing"
	"time"

	holiday "github.com/holiday-jp/holiday_jp-go"
)

// ビルドに成功すること自体が go.mod / go.sum の正しい生成を示すスモークテスト。
func TestHolidayJP_IsHoliday(t *testing.T) {
	cases := []struct {
		date time.Time
		want bool
	}{
		{time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2024, time.January, 4, 0, 0, 0, 0, time.UTC), false},
	}
	for _, tc := range cases {
		if got := holiday.IsHoliday(tc.date); got != tc.want {
			t.Errorf("IsHoliday(%s) = %v, want %v", tc.date.Format(dateLayout), got, tc.want)
		}
	}
}

func TestHolidayJP_HolidayName(t *testing.T) {
	target := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	name, err := holiday.HolidayName(target)
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(dateLayout), err)
	}
	if name != "元日" {
		t.Errorf("HolidayName(%s) = %q, want %q", target.Format(dateLayout), name, "元日")
	}
}
