// Package holidayjp は holiday-jp/holiday_jp-go の埋め込み祝日データを用いた
// heijitu.HolidayProvider 実装を提供する。外部ネットワーク接続は不要。
package holidayjp

import (
	"context"
	"slices"
	"time"

	heijitu "github.com/taku-o/go-heijitu"

	holiday "github.com/holiday-jp/holiday_jp-go"
)

// Provider は holiday_jp-go を HolidayProvider インターフェースにブリッジするゼロ値アダプター。
// 状態を持たない。
type Provider struct{}

// New は holidayjp プロバイダーを返す。引数不要・外部接続不要。
func New() *Provider {
	return &Provider{}
}

// IsHoliday は指定日が日本の国民祝日かどうかを返す。エラーは常に nil。
func (p *Provider) IsHoliday(_ context.Context, t time.Time) (bool, error) {
	return holiday.IsHoliday(t), nil
}

// HolidayName は指定日の祝日名を返す。非祝日の場合は ("", nil) を返す。
// holiday_jp-go は I/O なし埋め込みデータライブラリのため、エラーは「非祝日」の意味のみ。
func (p *Provider) HolidayName(_ context.Context, t time.Time) (string, error) {
	name, err := holiday.HolidayName(t)
	if err != nil {
		return "", nil
	}
	return name, nil
}

// HolidaysBetween は from〜to（両端含む）の祝日リストを日付昇順で返す。
func (p *Provider) HolidaysBetween(_ context.Context, from, to time.Time) ([]heijitu.Holiday, error) {
	holidays := holiday.Between(from, to)
	result := make([]heijitu.Holiday, 0, len(holidays))
	for dateStr, h := range holidays {
		t, err := time.ParseInLocation(time.DateOnly, dateStr, from.Location())
		if err != nil {
			return nil, err
		}
		result = append(result, heijitu.Holiday{Date: t, Name: h.Name()})
	}
	slices.SortFunc(result, func(a, b heijitu.Holiday) int {
		return a.Date.Compare(b.Date)
	})
	return result, nil
}
