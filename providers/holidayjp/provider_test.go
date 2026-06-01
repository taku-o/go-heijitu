package holidayjp_test

import (
	"context"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
	holidayjp "github.com/taku-o/go-heijitu/providers/holidayjp"
)

// コンパイル時のインターフェース充足チェック。
// このファイルがコンパイルされることで Provider が HolidayProvider を満たすことを保証する。
var _ heijitu.HolidayProvider = (*holidayjp.Provider)(nil)

// TestIsHoliday_Holiday は祝日の日付に対して IsHoliday が true を返すことを確認する。
func TestIsHoliday_Holiday(t *testing.T) {
	// Given: holidayjp プロバイダーと祝日（元日）
	p := holidayjp.New()
	ctx := context.Background()
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 祝日の日付に対して IsHoliday を呼ぶ
	got, err := p.IsHoliday(ctx, target)

	// Then: true と nil error が返る
	if err != nil {
		t.Fatalf("IsHoliday(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if !got {
		t.Errorf("IsHoliday(%s) = false, want true", target.Format(time.DateOnly))
	}
}

// TestIsHoliday_Weekday は平日の日付に対して IsHoliday が false を返すことを確認する。
func TestIsHoliday_Weekday(t *testing.T) {
	// Given: holidayjp プロバイダーと平日（2025-01-06 月曜）
	p := holidayjp.New()
	ctx := context.Background()
	target := time.Date(2025, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: 平日の日付に対して IsHoliday を呼ぶ
	got, err := p.IsHoliday(ctx, target)

	// Then: false と nil error が返る
	if err != nil {
		t.Fatalf("IsHoliday(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got {
		t.Errorf("IsHoliday(%s) = true, want false", target.Format(time.DateOnly))
	}
}

// TestHolidayName_Holiday は祝日の日付に対して HolidayName が正しい祝日名を返すことを確認する。
func TestHolidayName_Holiday(t *testing.T) {
	// Given: holidayjp プロバイダーと祝日（元日）
	p := holidayjp.New()
	ctx := context.Background()
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 祝日の日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: 祝日名 "元日" と nil error が返る
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got != "元日" {
		t.Errorf("HolidayName(%s) = %q, want %q", target.Format(time.DateOnly), got, "元日")
	}
}

// TestHolidayName_NonHoliday は平日の日付に対して HolidayName が空文字と nil error を返すことを確認する。
func TestHolidayName_NonHoliday(t *testing.T) {
	// Given: holidayjp プロバイダーと平日（2025-01-06）
	p := holidayjp.New()
	ctx := context.Background()
	target := time.Date(2025, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: 平日の日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: 空文字と nil error が返る（非祝日エラーは ("", nil) に変換される）
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got != "" {
		t.Errorf("HolidayName(%s) = %q, want empty string", target.Format(time.DateOnly), got)
	}
}

// TestHolidaysBetween_IncludesEndpoints は HolidaysBetween が両端の日付を含む祝日を昇順で返すことを確認する。
func TestHolidaysBetween_IncludesEndpoints(t *testing.T) {
	// Given: holidayjp プロバイダー、from=元日（2025-01-01）、to=成人の日（2025-01-13）
	p := holidayjp.New()
	ctx := context.Background()
	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.January, 13, 0, 0, 0, 0, time.UTC)

	// When: from〜to の範囲で HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, from, to)

	// Then: 元日と成人の日の2件が昇順で返る（両端を含む）
	if err != nil {
		t.Fatalf("HolidaysBetween returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("HolidaysBetween: got %d holidays, want 2", len(got))
	}
	wantFirst := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !got[0].Date.Equal(wantFirst) {
		t.Errorf("HolidaysBetween: got[0].Date = %s, want %s", got[0].Date.Format(time.DateOnly), wantFirst.Format(time.DateOnly))
	}
	wantSecond := time.Date(2025, time.January, 13, 0, 0, 0, 0, time.UTC)
	if !got[1].Date.Equal(wantSecond) {
		t.Errorf("HolidaysBetween: got[1].Date = %s, want %s", got[1].Date.Format(time.DateOnly), wantSecond.Format(time.DateOnly))
	}
}

// TestHolidaysBetween_FromAfterTo は from > to の場合に空スライスと nil error が返ることを確認する。
func TestHolidaysBetween_FromAfterTo(t *testing.T) {
	// Given: holidayjp プロバイダー、from > to（逆順）
	p := holidayjp.New()
	ctx := context.Background()
	from := time.Date(2025, time.January, 13, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC) // from > to

	// When: from が to より後の場合に HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, from, to)

	// Then: 空スライスと nil error が返る
	if err != nil {
		t.Fatalf("HolidaysBetween(from>to) returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("HolidaysBetween(from>to): got %d holidays, want 0", len(got))
	}
}
