package heijitu_test

import (
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
)

func TestMonthDay_Matches(t *testing.T) {
	tests := []struct {
		name string
		md   heijitu.MonthDay
		t    time.Time
		want bool
	}{
		{
			name: "月と日が一致する場合はtrueを返す",
			md:   heijitu.MonthDay{Month: time.January, Day: 1},
			t:    time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "年が異なっても月日が一致する場合はtrueを返す",
			md:   heijitu.MonthDay{Month: time.March, Day: 15},
			t:    time.Date(2020, time.March, 15, 12, 30, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "時刻が異なっても月日が一致する場合はtrueを返す",
			md:   heijitu.MonthDay{Month: time.July, Day: 4},
			t:    time.Date(2023, time.July, 4, 23, 59, 59, 999999999, time.UTC),
			want: true,
		},
		{
			name: "12月31日の月日が一致する場合はtrueを返す",
			md:   heijitu.MonthDay{Month: time.December, Day: 31},
			t:    time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "月が異なる場合はfalseを返す",
			md:   heijitu.MonthDay{Month: time.January, Day: 1},
			t:    time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "日が異なる場合はfalseを返す",
			md:   heijitu.MonthDay{Month: time.January, Day: 1},
			t:    time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "月も日も異なる場合はfalseを返す",
			md:   heijitu.MonthDay{Month: time.January, Day: 1},
			t:    time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "2月29日を設定した場合に閏年の2月29日ではtrueを返す",
			md:   heijitu.MonthDay{Month: time.February, Day: 29},
			t:    time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC), // 2024年は閏年
			want: true,
		},
		{
			name: "2月29日を設定した場合に平年の2月28日ではfalseを返す",
			md:   heijitu.MonthDay{Month: time.February, Day: 29},
			t:    time.Date(2023, time.February, 28, 0, 0, 0, 0, time.UTC), // 2023年は平年
			want: false,
		},
		{
			name: "平年にtime.Date(year, February, 29)を渡すとMarch 1に正規化されMonthDay{Feb,29}にはマッチしない",
			md:   heijitu.MonthDay{Month: time.February, Day: 29},
			t:    time.Date(2023, time.February, 29, 0, 0, 0, 0, time.UTC), // 平年: Goが自動的にMarch 1へ正規化
			want: false,
		},
		{
			name: "存在しない月13を設定した場合はいかなる日付にも一致しない",
			md:   heijitu.MonthDay{Month: time.Month(13), Day: 1},
			t:    time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "存在しない日32を設定した場合はいかなる日付にも一致しない",
			md:   heijitu.MonthDay{Month: time.January, Day: 32},
			t:    time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.md.Matches(tt.t)
			if got != tt.want {
				t.Errorf(
					"MonthDay{Month: %v, Day: %d}.Matches(%v) = %v, want %v",
					tt.md.Month, tt.md.Day, tt.t.Format(dateLayout), got, tt.want,
				)
			}
		})
	}
}
