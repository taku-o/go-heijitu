package heijitu

import (
	"context"
	"time"
)

// HolidayProvider は祝日情報を提供するインターフェース。
// 各メソッドが返すエラーは握りつぶさず即座に呼び出し元へ伝播しなければならない。
type HolidayProvider interface {
	IsHoliday(ctx context.Context, t time.Time) (bool, error)
	HolidayName(ctx context.Context, t time.Time) (string, error)
	HolidaysBetween(ctx context.Context, from, to time.Time) ([]Holiday, error)
}
