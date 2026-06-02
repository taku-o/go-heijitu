package heijitu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
	holidayjp "github.com/taku-o/go-heijitu/providers/holidayjp"
)

// --- NextBusinessDay のテスト ---

// TestNextBusinessDay_WeekdayToWeekday は平日から翌平日が返ることを確認する。
func TestNextBusinessDay_WeekdayToWeekday(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024-01-08（月曜日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	monday := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: 月曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, monday)

	// Then: 翌日の火曜日（2024-01-09）が返る（Requirement 2.1）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2024, time.January, 9, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", monday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_FridaySkipsWeekend は金曜日から翌月曜日が返ることを確認する。
func TestNextBusinessDay_FridaySkipsWeekend(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024-01-12（金曜日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	friday := time.Date(2024, time.January, 12, 0, 0, 0, 0, time.UTC)

	// When: 金曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, friday)

	// Then: 土日をスキップして月曜日（2024-01-15）が返る（Requirement 2.2）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", friday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_SaturdaySkipsToMonday は土曜日から翌月曜日が返ることを確認する。
func TestNextBusinessDay_SaturdaySkipsToMonday(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024-01-06（土曜日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	saturday := time.Date(2024, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: 土曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, saturday)

	// Then: 日曜をスキップして月曜日（2024-01-08）が返る（Requirement 2.2）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", saturday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_SkipsHoliday は祝日をスキップして翌営業日が返ることを確認する。
func TestNextBusinessDay_SkipsHoliday(t *testing.T) {
	// Given: 2024-01-09（火曜日）を祝日として登録したプロバイダーと2024-01-08（月曜日）
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-09": "テスト祝日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()
	monday := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: 翌日が祝日の月曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, monday)

	// Then: 祝日をスキップして水曜日（2024-01-10）が返る（Requirement 2.3）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", monday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_SkipsExcludedDate は除外日付をスキップして翌営業日が返ることを確認する。
func TestNextBusinessDay_SkipsExcludedDate(t *testing.T) {
	// Given: 1月9日を除外日付として登録した BusinessCalendar と2024-01-08（月曜日）
	p := emptyProvider()
	excluded := []heijitu.MonthDay{{Month: time.January, Day: 9}}
	bc := heijitu.New(p, heijitu.WithExcludedDates(excluded))
	ctx := context.Background()
	monday := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: 翌日が除外日付の月曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, monday)

	// Then: 除外日付をスキップして水曜日（2024-01-10）が返る（Requirement 2.4）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", monday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_SkipsConsecutiveNonBusinessDays は連続する非営業日（祝日+週末）をスキップすることを確認する。
func TestNextBusinessDay_SkipsConsecutiveNonBusinessDays(t *testing.T) {
	// Given: 金曜日の翌週月曜日が祝日のケース
	// 2024-01-12（金）→ 1/13（土）→ 1/14（日）→ 1/15（月・祝日）→ 1/16（火）が答え
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-15": "テスト祝日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()
	friday := time.Date(2024, time.January, 12, 0, 0, 0, 0, time.UTC)

	// When: 金曜日に NextBusinessDay を呼ぶ（翌月曜が祝日）
	got, err := bc.NextBusinessDay(ctx, friday)

	// Then: 土日+祝日をスキップして火曜日（2024-01-16）が返る（Requirement 2.2, 2.3）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2024, time.January, 16, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", friday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_FromDateNotReturned は from 自身が営業日でも from は返されないことを確認する。
func TestNextBusinessDay_FromDateNotReturned(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024-01-08（月曜日・営業日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	monday := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: 営業日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, monday)

	// Then: from 自身ではなく翌日が返る（Requirement 2.1 — strictly after）
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	if got.Equal(monday) {
		t.Errorf("NextBusinessDay(%v) returned from date itself, want strictly after", monday.Format(dateLayout))
	}
}

// TestNextBusinessDay_ProviderError はプロバイダーがエラーを返したとき伝播されることを確認する。
func TestNextBusinessDay_ProviderError(t *testing.T) {
	// Given: IsHoliday でエラーを返すプロバイダーと2024-01-08（月曜日）
	providerErr := errors.New("provider error")
	p := &testProvider{err: providerErr}
	bc := heijitu.New(p)
	ctx := context.Background()
	monday := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: NextBusinessDay を呼ぶ
	_, err := bc.NextBusinessDay(ctx, monday)

	// Then: プロバイダーのエラーが伝播する（Requirement 2.5）
	if err == nil {
		t.Fatal("NextBusinessDay: expected error to be propagated, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("NextBusinessDay: expected %v, got %v", providerErr, err)
	}
}

// --- holidayjp プロバイダーを使った統合テスト ---

// TestNextBusinessDay_Integration_FridaySkipsWeekendAndHoliday は holidayjp プロバイダーで
// 金曜日から土日および直後の祝日をスキップして翌営業日が返ることを確認する。
func TestNextBusinessDay_Integration_FridaySkipsWeekendAndHoliday(t *testing.T) {
	// Given: holidayjp プロバイダーと2025-01-10（金曜日）
	// 2025-01-11（土）、1/12（日）、1/13（月・成人の日）をスキップして1/14（火）が翌営業日
	p := holidayjp.New()
	bc := heijitu.New(p)
	ctx := context.Background()
	friday := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)

	// When: 金曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, friday)

	// Then: 土日+祝日をスキップして火曜日（2025-01-14）が返る
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2025, time.January, 14, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", friday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestNextBusinessDay_Integration_SkipsHoliday は holidayjp プロバイダーで祝日をスキップして翌営業日が返ることを確認する。
func TestNextBusinessDay_Integration_SkipsHoliday(t *testing.T) {
	// Given: holidayjp プロバイダーと2025-02-10（月曜日）
	// 2025-02-11（火曜日）は建国記念の日
	p := holidayjp.New()
	bc := heijitu.New(p)
	ctx := context.Background()
	monday := time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC)

	// When: 翌日が祝日の月曜日に NextBusinessDay を呼ぶ
	got, err := bc.NextBusinessDay(ctx, monday)

	// Then: 建国記念の日をスキップして水曜日（2025-02-12）が返る
	if err != nil {
		t.Fatalf("NextBusinessDay returned error: %v", err)
	}
	want := time.Date(2025, time.February, 12, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextBusinessDay(%v) = %v, want %v", monday.Format(dateLayout), got.Format(dateLayout), want.Format(dateLayout))
	}
}
