package heijitu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
)

// --- FirstBusinessDayOfMonth のテスト ---

// TestFirstBusinessDayOfMonth_FirstDayIsWeekday は月初が平日・非祝日の場合に1日が返ることを確認する。
func TestFirstBusinessDayOfMonth_FirstDayIsWeekday(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024年7月（7/1は月曜日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: 2024年7月の FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.July)

	// Then: 7月1日（月曜日）が返る（Requirement 3.1）
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	want := time.Date(2024, time.July, 1, 0, 0, 0, 0, time.Local)
	if got.Day() != want.Day() || got.Month() != want.Month() || got.Year() != want.Year() {
		t.Errorf("FirstBusinessDayOfMonth(2024, July) = %v, want %v", got.Format(dateLayout), want.Format(dateLayout))
	}
}

// TestFirstBusinessDayOfMonth_FirstDayIsSaturday は月初が土曜日の場合に翌月曜日が返ることを確認する。
func TestFirstBusinessDayOfMonth_FirstDayIsSaturday(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024年6月（6/1は土曜日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: 2024年6月の FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.June)

	// Then: 6月3日（月曜日）が返る（Requirement 3.2）
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	if got.Day() != 3 || got.Month() != time.June {
		t.Errorf("FirstBusinessDayOfMonth(2024, June) = %v, want 2024-06-03", got.Format(dateLayout))
	}
}

// TestFirstBusinessDayOfMonth_FirstDayIsSunday は月初が日曜日の場合に翌月曜日が返ることを確認する。
func TestFirstBusinessDayOfMonth_FirstDayIsSunday(t *testing.T) {
	// Given: 祝日なしのプロバイダーと2024年9月（9/1は日曜日）
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: 2024年9月の FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.September)

	// Then: 9月2日（月曜日）が返る（Requirement 3.2）
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	if got.Day() != 2 || got.Month() != time.September {
		t.Errorf("FirstBusinessDayOfMonth(2024, September) = %v, want 2024-09-02", got.Format(dateLayout))
	}
}

// TestFirstBusinessDayOfMonth_FirstDayIsHoliday は月初が祝日の場合に翌営業日が返ることを確認する。
func TestFirstBusinessDayOfMonth_FirstDayIsHoliday(t *testing.T) {
	// Given: 2024-01-01（月曜日・元日）を祝日として登録したプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: 2024年1月の FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.January)

	// Then: 1月2日（火曜日）が返る（Requirement 3.2）
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	if got.Day() != 2 || got.Month() != time.January {
		t.Errorf("FirstBusinessDayOfMonth(2024, January) = %v, want 2024-01-02", got.Format(dateLayout))
	}
}

// TestFirstBusinessDayOfMonth_ConsecutiveNonBusinessDays は月初から連続する非営業日をスキップすることを確認する。
func TestFirstBusinessDayOfMonth_ConsecutiveNonBusinessDays(t *testing.T) {
	// Given: 2024年6月（6/1土曜、6/2日曜）で6/3（月曜）を祝日として登録
	p := &testProvider{
		holidays: map[string]string{
			"2024-06-03": "テスト祝日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: 2024年6月の FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.June)

	// Then: 土日+祝日をスキップして6月4日（火曜日）が返る（Requirement 3.2）
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	if got.Day() != 4 || got.Month() != time.June {
		t.Errorf("FirstBusinessDayOfMonth(2024, June) = %v, want 2024-06-04", got.Format(dateLayout))
	}
}

// TestFirstBusinessDayOfMonth_WithExcludedDates は除外日付が考慮されることを確認する。
func TestFirstBusinessDayOfMonth_WithExcludedDates(t *testing.T) {
	// Given: 7月1日を除外日付として登録（2024-07-01は月曜日）
	p := emptyProvider()
	excluded := []heijitu.MonthDay{{Month: time.July, Day: 1}}
	bc := heijitu.New(p, heijitu.WithExcludedDates(excluded))
	ctx := context.Background()

	// When: 2024年7月の FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.July)

	// Then: 除外日付をスキップして7月2日（火曜日）が返る（Requirement 3.2）
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	if got.Day() != 2 || got.Month() != time.July {
		t.Errorf("FirstBusinessDayOfMonth(2024, July) with excluded = %v, want 2024-07-02", got.Format(dateLayout))
	}
}

// TestFirstBusinessDayOfMonth_ProviderError はプロバイダーがエラーを返したとき伝播されることを確認する。
func TestFirstBusinessDayOfMonth_ProviderError(t *testing.T) {
	// Given: IsHoliday でエラーを返すプロバイダー
	providerErr := errors.New("provider error")
	p := &testProvider{err: providerErr}
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: FirstBusinessDayOfMonth を呼ぶ（2024年7月、7/1は月曜日なのでIsHolidayが呼ばれる）
	_, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.July)

	// Then: プロバイダーのエラーが伝播する（Requirement 3.3）
	if err == nil {
		t.Fatal("FirstBusinessDayOfMonth: expected error to be propagated, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("FirstBusinessDayOfMonth: expected %v, got %v", providerErr, err)
	}
}

// TestFirstBusinessDayOfMonth_UsesTimeLocal は返り値が time.Local を使用していることを確認する。
func TestFirstBusinessDayOfMonth_UsesTimeLocal(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: FirstBusinessDayOfMonth を呼ぶ
	got, err := bc.FirstBusinessDayOfMonth(ctx, 2024, time.July)

	// Then: 返り値の Location が time.Local と同じ
	if err != nil {
		t.Fatalf("FirstBusinessDayOfMonth returned error: %v", err)
	}
	if got.Location() != time.Local {
		t.Errorf("FirstBusinessDayOfMonth: Location = %v, want %v", got.Location(), time.Local)
	}
}

// --- FirstBusinessDaysOfYear のテスト ---

// TestFirstBusinessDaysOfYear_Returns12Elements は12要素のスライスが返ることを確認する。
func TestFirstBusinessDaysOfYear_Returns12Elements(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: FirstBusinessDaysOfYear を呼ぶ
	got, err := bc.FirstBusinessDaysOfYear(ctx, 2024)

	// Then: 12要素のスライスが返る（Requirement 4.1）
	if err != nil {
		t.Fatalf("FirstBusinessDaysOfYear returned error: %v", err)
	}
	if len(got) != 12 {
		t.Fatalf("FirstBusinessDaysOfYear: got %d elements, want 12", len(got))
	}
}

// TestFirstBusinessDaysOfYear_IndexCorrespondsToMonth は各要素が対応する月の最初の営業日であることを確認する。
func TestFirstBusinessDaysOfYear_IndexCorrespondsToMonth(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: FirstBusinessDaysOfYear を呼ぶ
	got, err := bc.FirstBusinessDaysOfYear(ctx, 2024)

	// Then: index 0 = 1月、index 11 = 12月で、各要素の月が正しい（Requirement 4.1, 4.2）
	if err != nil {
		t.Fatalf("FirstBusinessDaysOfYear returned error: %v", err)
	}
	for i, d := range got {
		expectedMonth := time.Month(i + 1)
		if d.Month() != expectedMonth {
			t.Errorf("FirstBusinessDaysOfYear[%d]: month = %v, want %v", i, d.Month(), expectedMonth)
		}
		if d.Year() != 2024 {
			t.Errorf("FirstBusinessDaysOfYear[%d]: year = %d, want 2024", i, d.Year())
		}
	}
}

// TestFirstBusinessDaysOfYear_ConsistentWithFirstBusinessDayOfMonth は
// 各要素が FirstBusinessDayOfMonth の結果と一致することを確認する。
func TestFirstBusinessDaysOfYear_ConsistentWithFirstBusinessDayOfMonth(t *testing.T) {
	// Given: 1月1日を祝日として登録したプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: FirstBusinessDaysOfYear と各月の FirstBusinessDayOfMonth を呼ぶ
	yearResult, err := bc.FirstBusinessDaysOfYear(ctx, 2024)
	if err != nil {
		t.Fatalf("FirstBusinessDaysOfYear returned error: %v", err)
	}

	// Then: 各要素が FirstBusinessDayOfMonth の結果と一致する（Requirement 4.2）
	for i := 0; i < 12; i++ {
		month := time.Month(i + 1)
		monthResult, err := bc.FirstBusinessDayOfMonth(ctx, 2024, month)
		if err != nil {
			t.Fatalf("FirstBusinessDayOfMonth(%v) returned error: %v", month, err)
		}
		if !yearResult[i].Equal(monthResult) {
			t.Errorf("FirstBusinessDaysOfYear[%d] = %v, FirstBusinessDayOfMonth(%v) = %v",
				i, yearResult[i].Format(dateLayout), month, monthResult.Format(dateLayout))
		}
	}
}

// TestFirstBusinessDaysOfYear_ProviderError はプロバイダーがエラーを返したとき伝播されることを確認する。
func TestFirstBusinessDaysOfYear_ProviderError(t *testing.T) {
	// Given: IsHoliday でエラーを返すプロバイダー
	providerErr := errors.New("provider error")
	p := &testProvider{err: providerErr}
	bc := heijitu.New(p)
	ctx := context.Background()

	// When: FirstBusinessDaysOfYear を呼ぶ
	_, err := bc.FirstBusinessDaysOfYear(ctx, 2024)

	// Then: プロバイダーのエラーが伝播する（Requirement 4.3）
	if err == nil {
		t.Fatal("FirstBusinessDaysOfYear: expected error to be propagated, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("FirstBusinessDaysOfYear: expected %v, got %v", providerErr, err)
	}
}
