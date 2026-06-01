package heijitu

import (
	"context"
	"time"
)

// BusinessCalendar は祝日プロバイダーと除外日付を保持し、営業日判定を行う。
type BusinessCalendar struct {
	provider      HolidayProvider
	excludedDates []MonthDay
}

// New は HolidayProvider と任意の Option から BusinessCalendar を生成する。
// provider が nil の場合はパニックする。
func New(provider HolidayProvider, opts ...Option) *BusinessCalendar {
	if provider == nil {
		panic("heijitu: provider must not be nil")
	}
	bc := &BusinessCalendar{
		provider: provider,
	}
	for _, opt := range opts {
		opt(bc)
	}
	return bc
}

// IsBusinessDay は指定した日付が営業日かどうかを返す。
// 土曜日・日曜日・祝日・除外日付に一致する場合は false を返す。
// extraExcluded はこの呼び出し限りで有効な追加除外日付。
func (bc *BusinessCalendar) IsBusinessDay(ctx context.Context, t time.Time, extraExcluded ...MonthDay) (bool, error) {
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false, nil
	}

	isHoliday, err := bc.provider.IsHoliday(ctx, t)
	if err != nil {
		return false, err
	}
	if isHoliday {
		return false, nil
	}

	if isExcluded(t, bc.excludedDates) {
		return false, nil
	}

	if isExcluded(t, extraExcluded) {
		return false, nil
	}

	return true, nil
}

func isExcluded(t time.Time, dates []MonthDay) bool {
	for _, md := range dates {
		if md.Matches(t) {
			return true
		}
	}
	return false
}
