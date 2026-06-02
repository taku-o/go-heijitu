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

// NextBusinessDay は from の翌日以降で最初の営業日を返す。
// from 自身は返却対象に含まない。
func (bc *BusinessCalendar) NextBusinessDay(ctx context.Context, from time.Time) (time.Time, error) {
	candidate := from.AddDate(0, 0, 1)
	for {
		ok, err := bc.IsBusinessDay(ctx, candidate)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			return candidate, nil
		}
		candidate = candidate.AddDate(0, 0, 1)
	}
}

// FirstBusinessDayOfMonth は指定年月の最初の営業日を返す。
func (bc *BusinessCalendar) FirstBusinessDayOfMonth(ctx context.Context, year int, month time.Month) (time.Time, error) {
	candidate := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	for candidate.Month() == month {
		ok, err := bc.IsBusinessDay(ctx, candidate)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			return candidate, nil
		}
		candidate = candidate.AddDate(0, 0, 1)
	}
	return time.Time{}, nil
}

// FirstBusinessDaysOfYear は指定年の各月の最初の営業日を12要素のスライスで返す。
// index 0 が1月、index 11 が12月に対応する。
func (bc *BusinessCalendar) FirstBusinessDaysOfYear(ctx context.Context, year int) ([]time.Time, error) {
	result := make([]time.Time, 12)
	for i := 0; i < 12; i++ {
		month := time.Month(i + 1)
		d, err := bc.FirstBusinessDayOfMonth(ctx, year, month)
		if err != nil {
			return nil, err
		}
		result[i] = d
	}
	return result, nil
}

// Holidays は指定期間の祝日リストをプロバイダーに委譲して返す。
// 除外日付によるフィルタリングは行わない。
func (bc *BusinessCalendar) Holidays(ctx context.Context, from, to time.Time) ([]Holiday, error) {
	return bc.provider.HolidaysBetween(ctx, from, to)
}

func isExcluded(t time.Time, dates []MonthDay) bool {
	for _, md := range dates {
		if md.Matches(t) {
			return true
		}
	}
	return false
}
