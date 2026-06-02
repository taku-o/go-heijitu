package heijitu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
	holidayjp "github.com/taku-o/go-heijitu/providers/holidayjp"
)

// --- Holidays のテスト ---

// TestHolidays_ReturnsProviderHolidays はプロバイダーの祝日リストがそのまま返ることを確認する。
func TestHolidays_ReturnsProviderHolidays(t *testing.T) {
	// Given: 3件の祝日が登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
			"2024-01-08": "成人の日",
			"2024-02-11": "建国記念の日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)

	// When: Holidays を呼ぶ
	got, err := bc.Holidays(ctx, from, to)

	// Then: 3件の祝日が返る（Requirement 5.1）
	if err != nil {
		t.Fatalf("Holidays returned error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("Holidays: got %d holidays, want 3", len(got))
	}
}

// TestHolidays_IncludesEndpoints は両端の日付を含むことを確認する。
func TestHolidays_IncludesEndpoints(t *testing.T) {
	// Given: from と to 当日が祝日として登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
			"2024-01-08": "成人の日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: from〜to の範囲で Holidays を呼ぶ
	got, err := bc.Holidays(ctx, from, to)

	// Then: 両端を含む2件が返る（Requirement 5.1）
	if err != nil {
		t.Fatalf("Holidays returned error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Holidays: got %d holidays, want 2", len(got))
	}
}

// TestHolidays_DoesNotIncludeExcludedDates は除外日付が Holidays の結果に含まれないことを確認する。
func TestHolidays_DoesNotIncludeExcludedDates(t *testing.T) {
	// Given: 1月1日が祝日で、1月1日を除外日付として登録した BusinessCalendar
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
			"2024-01-08": "成人の日",
		},
	}
	excluded := []heijitu.MonthDay{{Month: time.January, Day: 1}}
	bc := heijitu.New(p, heijitu.WithExcludedDates(excluded))
	ctx := context.Background()
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC)

	// When: Holidays を呼ぶ
	got, err := bc.Holidays(ctx, from, to)

	// Then: 除外日付は Holidays の結果に影響しない。プロバイダーの祝日がそのまま返る（Requirement 5.2）
	// Holidays はプロバイダーに委譲するだけなので、除外日付は含まれない
	if err != nil {
		t.Fatalf("Holidays returned error: %v", err)
	}
	// プロバイダーの祝日2件がそのまま返る（除外日付はフィルタされない）
	if len(got) != 2 {
		t.Errorf("Holidays: got %d holidays, want 2 (excluded dates should not filter provider holidays)", len(got))
	}
}

// TestHolidays_EmptyRange は祝日がない期間で空スライスが返ることを確認する。
func TestHolidays_EmptyRange(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC)

	// When: 祝日がない期間で Holidays を呼ぶ
	got, err := bc.Holidays(ctx, from, to)

	// Then: 空スライスが返る
	if err != nil {
		t.Fatalf("Holidays returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Holidays: got %d holidays, want 0", len(got))
	}
}

// TestHolidays_FromAfterTo は from > to の場合に空スライスが返ることを確認する。
func TestHolidays_FromAfterTo(t *testing.T) {
	// Given: 祝日が登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()
	from := time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: from > to で Holidays を呼ぶ
	got, err := bc.Holidays(ctx, from, to)

	// Then: 空スライスと nil error が返る
	if err != nil {
		t.Fatalf("Holidays returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Holidays(from>to): got %d holidays, want 0", len(got))
	}
}

// TestHolidays_ProviderError はプロバイダーがエラーを返したとき伝播されることを確認する。
func TestHolidays_ProviderError(t *testing.T) {
	// Given: HolidaysBetween でエラーを返すプロバイダー
	providerErr := errors.New("provider error")
	p := &testProvider{err: providerErr}
	bc := heijitu.New(p)
	ctx := context.Background()
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)

	// When: Holidays を呼ぶ
	_, err := bc.Holidays(ctx, from, to)

	// Then: プロバイダーのエラーが伝播する（Requirement 5.3）
	if err == nil {
		t.Fatal("Holidays: expected error to be propagated, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("Holidays: expected %v, got %v", providerErr, err)
	}
}

// --- holidayjp プロバイダーを使った統合テスト ---

// TestHolidays_Integration_ReturnsHolidaysInRange は holidayjp プロバイダーで指定期間の祝日が正しい件数で返ることを確認する。
func TestHolidays_Integration_ReturnsHolidaysInRange(t *testing.T) {
	// Given: holidayjp プロバイダーと2026年1月1日〜3月31日の範囲
	// 2026年1〜3月の祝日: 元日(1/1)、成人の日(1/12)、建国記念の日(2/11)、天皇誕生日(2/23)、春分の日(3/20) = 5件
	p := holidayjp.New()
	bc := heijitu.New(p)
	ctx := context.Background()
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)

	// When: 2026年1月1日〜3月31日の範囲で Holidays を呼ぶ
	got, err := bc.Holidays(ctx, from, to)

	// Then: 5件の祝日が日付昇順・正しい名称で返る
	if err != nil {
		t.Fatalf("Holidays returned error: %v", err)
	}
	want := []heijitu.Holiday{
		{Date: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), Name: "元日"},
		{Date: time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC), Name: "成人の日"},
		{Date: time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC), Name: "建国記念の日"},
		{Date: time.Date(2026, time.February, 23, 0, 0, 0, 0, time.UTC), Name: "天皇誕生日"},
		{Date: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC), Name: "春分の日"},
	}
	if len(got) != len(want) {
		t.Fatalf("Holidays(2026-01-01 to 2026-03-31): got %d holidays, want %d", len(got), len(want))
	}
	for i, w := range want {
		if !got[i].Date.Equal(w.Date) {
			t.Errorf("Holidays[%d].Date = %s, want %s", i, got[i].Date.Format(dateLayout), w.Date.Format(dateLayout))
		}
		if got[i].Name != w.Name {
			t.Errorf("Holidays[%d].Name = %q, want %q", i, got[i].Name, w.Name)
		}
	}
}
