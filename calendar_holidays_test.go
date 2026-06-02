package heijitu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
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
