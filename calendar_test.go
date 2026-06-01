package heijitu_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
)

// writeTempCalendarFile はカレンダーテスト用の一時ファイルを dir に作成してパスを返す。
// config_test.go の writeTempFile は package heijitu 内部専用のため、
// package heijitu_test からはアクセスできないため別途定義する。
func writeTempCalendarFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempCalendarFile: %v", err)
	}
	return path
}

// emptyProvider は祝日を1件も持たない testProvider を返す。
func emptyProvider() *testProvider {
	return &testProvider{holidays: map[string]string{}}
}

// --- New / Option のテスト ---

// TestNew_CreatesCalendarWithNoOptions は、オプションなしで New を呼び出した場合に
// BusinessCalendar が正常に構築され、平日の非祝日に true を返すことを確認する。
func TestNew_CreatesCalendarWithNoOptions(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	// When: オプションなしで New を呼ぶ
	bc := heijitu.New(p)
	ctx := context.Background()
	// 2024-01-08 月曜日・非祝日
	target := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// Then: エラーなし・true が返る（Requirement 5.5）
	got, err := bc.IsBusinessDay(ctx, target)
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if !got {
		t.Errorf("IsBusinessDay(%v) = false, want true", target.Format(dateLayout))
	}
}

// --- IsBusinessDay のテスト ---

// TestIsBusinessDay_Saturday は土曜日に false を返すことを確認する。
func TestIsBusinessDay_Saturday(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	// 2024-01-06 土曜日
	saturday := time.Date(2024, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: 土曜日に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, saturday)

	// Then: false と nil error が返る（Requirement 5.1）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if got {
		t.Errorf("IsBusinessDay(%v) = true, want false for Saturday", saturday.Format(dateLayout))
	}
}

// TestIsBusinessDay_Sunday は日曜日に false を返すことを確認する。
func TestIsBusinessDay_Sunday(t *testing.T) {
	// Given: 祝日なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	// 2024-01-07 日曜日
	sunday := time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC)

	// When: 日曜日に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, sunday)

	// Then: false と nil error が返る（Requirement 5.1）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if got {
		t.Errorf("IsBusinessDay(%v) = true, want false for Sunday", sunday.Format(dateLayout))
	}
}

// TestIsBusinessDay_Holiday はプロバイダーが祝日と判定する日付に false を返すことを確認する。
func TestIsBusinessDay_Holiday(t *testing.T) {
	// Given: 2024-01-08（月曜日）を祝日として登録したプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-08": "成人の日",
		},
	}
	bc := heijitu.New(p)
	ctx := context.Background()
	target := time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC)

	// When: 祝日の月曜日に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, target)

	// Then: false と nil error が返る（Requirement 5.2）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if got {
		t.Errorf("IsBusinessDay(%v) = true, want false for holiday", target.Format(dateLayout))
	}
}

// TestIsBusinessDay_WeekdayNonHolidayNoExclusion は平日・非祝日・除外日付なしの場合に true を返すことを確認する。
func TestIsBusinessDay_WeekdayNonHolidayNoExclusion(t *testing.T) {
	// Given: 祝日なし・除外日付なしのプロバイダー
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	// 2024-01-09 火曜日
	target := time.Date(2024, time.January, 9, 0, 0, 0, 0, time.UTC)

	// When: 平日・非祝日・除外日付なしの日付に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, target)

	// Then: true と nil error が返る（Requirement 5.5）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if !got {
		t.Errorf("IsBusinessDay(%v) = false, want true for regular weekday", target.Format(dateLayout))
	}
}

// TestIsBusinessDay_ProviderError はプロバイダーがエラーを返したとき IsBusinessDay がそのエラーを伝播することを確認する。
func TestIsBusinessDay_ProviderError(t *testing.T) {
	// Given: IsHoliday でエラーを返すプロバイダー
	providerErr := errors.New("provider connection failed")
	p := &testProvider{err: providerErr}
	bc := heijitu.New(p)
	ctx := context.Background()
	// 2024-01-09 火曜日（土日ではない日付）
	target := time.Date(2024, time.January, 9, 0, 0, 0, 0, time.UTC)

	// When: IsHoliday がエラーを返す日付に IsBusinessDay を呼ぶ
	_, err := bc.IsBusinessDay(ctx, target)

	// Then: プロバイダーのエラーが伝播する（Requirement 3.7）
	if err == nil {
		t.Fatal("IsBusinessDay: expected error to be propagated, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("IsBusinessDay: expected %v, got %v", providerErr, err)
	}
}

// --- WithExcludedDates のテスト ---

// TestWithExcludedDates_ExcludedDateReturnsFalse は WithExcludedDates で登録した日付に false を返すことを確認する。
func TestWithExcludedDates_ExcludedDateReturnsFalse(t *testing.T) {
	// Given: 3月21日を除外日付として登録した BusinessCalendar（祝日なしのプロバイダー）
	// 2024-03-21 木曜日
	p := emptyProvider()
	excluded := []heijitu.MonthDay{
		{Month: time.March, Day: 21},
	}
	bc := heijitu.New(p, heijitu.WithExcludedDates(excluded))
	ctx := context.Background()
	target := time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC)

	// When: 除外日付に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, target)

	// Then: false と nil error が返る（Requirement 4.2, 5.3）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if got {
		t.Errorf("IsBusinessDay(%v) = true, want false for excluded date", target.Format(dateLayout))
	}
}

// TestWithExcludedDates_NonExcludedWeekdayReturnsTrue は WithExcludedDates に含まれない平日に true を返すことを確認する。
func TestWithExcludedDates_NonExcludedWeekdayReturnsTrue(t *testing.T) {
	// Given: 3月21日のみを除外日付として登録した BusinessCalendar
	p := emptyProvider()
	excluded := []heijitu.MonthDay{
		{Month: time.March, Day: 21},
	}
	bc := heijitu.New(p, heijitu.WithExcludedDates(excluded))
	ctx := context.Background()
	// 2024-03-22 金曜日（除外対象外）
	target := time.Date(2024, time.March, 22, 0, 0, 0, 0, time.UTC)

	// When: 除外対象外の平日に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, target)

	// Then: true と nil error が返る
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if !got {
		t.Errorf("IsBusinessDay(%v) = false, want true for non-excluded weekday", target.Format(dateLayout))
	}
}

// TestWithExcludedDates_MultipleOptions は複数の WithExcludedDates オプションが追記（マージ）されることを確認する。
func TestWithExcludedDates_MultipleOptions(t *testing.T) {
	// Given: 2つの WithExcludedDates オプションを渡した BusinessCalendar
	p := emptyProvider()
	opt1 := heijitu.WithExcludedDates([]heijitu.MonthDay{{Month: time.March, Day: 21}})
	opt2 := heijitu.WithExcludedDates([]heijitu.MonthDay{{Month: time.September, Day: 23}})
	bc := heijitu.New(p, opt1, opt2)
	ctx := context.Background()

	tests := []struct {
		name   string
		target time.Time
		want   bool
	}{
		{
			name:   "opt1 の除外日付（3月21日）は false",
			target: time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC), // 木曜日
			want:   false,
		},
		{
			name:   "opt2 の除外日付（9月23日）は false",
			target: time.Date(2024, time.September, 23, 0, 0, 0, 0, time.UTC), // 月曜日
			want:   false,
		},
		{
			name:   "いずれの除外日付にも含まれない平日は true",
			target: time.Date(2024, time.March, 22, 0, 0, 0, 0, time.UTC), // 金曜日
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bc.IsBusinessDay(ctx, tt.target)
			if err != nil {
				t.Fatalf("IsBusinessDay returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsBusinessDay(%v) = %v, want %v", tt.target.Format(dateLayout), got, tt.want)
			}
		})
	}
}

// --- extraExcluded のテスト ---

// TestIsBusinessDay_ExtraExcluded_AppliedOnlyForTheCall は extraExcluded が当該呼び出し限りで有効であり、
// 次の呼び出しには影響しないことを確認する。
func TestIsBusinessDay_ExtraExcluded_AppliedOnlyForTheCall(t *testing.T) {
	// Given: 除外日付なしの BusinessCalendar
	p := emptyProvider()
	bc := heijitu.New(p)
	ctx := context.Background()
	// 2024-01-15 月曜日
	target := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	extra := heijitu.MonthDay{Month: time.January, Day: 15}

	// When: extraExcluded を渡して IsBusinessDay を呼ぶ
	got1, err := bc.IsBusinessDay(ctx, target, extra)

	// Then: この呼び出しでは false（extraExcluded に一致する）（Requirement 5.4）
	if err != nil {
		t.Fatalf("IsBusinessDay (with extra) returned error: %v", err)
	}
	if got1 {
		t.Errorf("IsBusinessDay(%v) with extra = true, want false", target.Format(dateLayout))
	}

	// When: 同じ日付に extraExcluded なしで IsBusinessDay を呼ぶ
	got2, err := bc.IsBusinessDay(ctx, target)

	// Then: 次の呼び出しでは true（extraExcluded の影響を受けない）（Requirement 5.4）
	if err != nil {
		t.Fatalf("IsBusinessDay (without extra) returned error: %v", err)
	}
	if !got2 {
		t.Errorf("IsBusinessDay(%v) without extra = false, want true (extra should not persist)", target.Format(dateLayout))
	}
}

// --- WithConfig のテスト ---

// TestWithConfig_ValidYAML_ExcludedDateReturnsFalse は WithConfig が有効な YAML を読み込み、
// 除外日付の日付に false を返すことを確認する。
func TestWithConfig_ValidYAML_ExcludedDateReturnsFalse(t *testing.T) {
	// Given: 3月21日を excluded_dates に含む有効な YAML 設定ファイル
	dir := t.TempDir()
	path := writeTempCalendarFile(t, dir, "config.yaml", `
excluded_dates:
  - month: 3
    day: 21
`)
	opt, err := heijitu.WithConfig(path)
	if err != nil {
		t.Fatalf("WithConfig returned error: %v", err)
	}
	p := emptyProvider()
	bc := heijitu.New(p, opt)
	ctx := context.Background()
	// 2024-03-21 木曜日
	target := time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC)

	// When: YAML で登録した除外日付に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, target)

	// Then: false と nil error が返る（Requirement 4.3, 5.3）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if got {
		t.Errorf("IsBusinessDay(%v) = true, want false for WithConfig excluded date", target.Format(dateLayout))
	}
}

// TestWithConfig_FileNotFound は存在しないファイルパスを渡した場合にエラーが返ることを確認する。
func TestWithConfig_FileNotFound(t *testing.T) {
	// Given: 存在しないファイルパス
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")

	// When: 存在しないパスを WithConfig に渡す
	_, err := heijitu.WithConfig(path)

	// Then: os.ErrNotExist を含むエラーが返る（Requirement 4.5）
	if err == nil {
		t.Fatal("WithConfig: expected error for nonexistent file, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("WithConfig: expected os.ErrNotExist, got %v", err)
	}
}

// TestWithConfig_InvalidYAML はパースできない YAML ファイルを渡した場合にエラーが返ることを確認する。
func TestWithConfig_InvalidYAML(t *testing.T) {
	// Given: 不正な YAML 内容のファイル
	dir := t.TempDir()
	path := writeTempCalendarFile(t, dir, "invalid.yaml", `
excluded_dates:
  - month: [unclosed bracket
`)

	// When: 不正な YAML ファイルを WithConfig に渡す
	_, err := heijitu.WithConfig(path)

	// Then: パースエラーが返る（Requirement 4.5）
	if err == nil {
		t.Fatal("WithConfig: expected parse error for invalid YAML, got nil")
	}
}

// TestWithConfig_ValidJSON_ExcludedDateReturnsFalse は WithConfig が有効な JSON を読み込み、
// 除外日付の日付に false を返すことを確認する。
func TestWithConfig_ValidJSON_ExcludedDateReturnsFalse(t *testing.T) {
	// Given: 9月23日を excluded_dates に含む有効な JSON 設定ファイル
	dir := t.TempDir()
	path := writeTempCalendarFile(t, dir, "config.json", `{
  "excluded_dates": [
    {"month": 9, "day": 23}
  ]
}`)
	opt, err := heijitu.WithConfig(path)
	if err != nil {
		t.Fatalf("WithConfig returned error: %v", err)
	}
	p := emptyProvider()
	bc := heijitu.New(p, opt)
	ctx := context.Background()
	// 2024-09-23 月曜日
	target := time.Date(2024, time.September, 23, 0, 0, 0, 0, time.UTC)

	// When: JSON で登録した除外日付に IsBusinessDay を呼ぶ
	got, err := bc.IsBusinessDay(ctx, target)

	// Then: false と nil error が返る（Requirement 4.3）
	if err != nil {
		t.Fatalf("IsBusinessDay returned error: %v", err)
	}
	if got {
		t.Errorf("IsBusinessDay(%v) = true, want false for WithConfig JSON excluded date", target.Format(dateLayout))
	}
}

// --- WithExcludedDates + WithConfig 併用のテスト ---

// TestWithExcludedDates_And_WithConfig_BothApply は WithExcludedDates と WithConfig を併用した場合に
// 両方の除外日付がマージされることを確認する。
func TestWithExcludedDates_And_WithConfig_BothApply(t *testing.T) {
	// Given: 9月23日を WithExcludedDates で、3月21日を WithConfig（YAML）で登録した BusinessCalendar
	dir := t.TempDir()
	path := writeTempCalendarFile(t, dir, "config.yaml", `
excluded_dates:
  - month: 3
    day: 21
`)
	configOpt, err := heijitu.WithConfig(path)
	if err != nil {
		t.Fatalf("WithConfig returned error: %v", err)
	}
	datesOpt := heijitu.WithExcludedDates([]heijitu.MonthDay{
		{Month: time.September, Day: 23},
	})
	p := emptyProvider()
	bc := heijitu.New(p, datesOpt, configOpt)
	ctx := context.Background()

	tests := []struct {
		name   string
		target time.Time
		want   bool
	}{
		{
			name:   "WithExcludedDates の除外日付（9月23日）は false",
			target: time.Date(2024, time.September, 23, 0, 0, 0, 0, time.UTC), // 月曜日
			want:   false,
		},
		{
			name:   "WithConfig の除外日付（3月21日）は false",
			target: time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC), // 木曜日
			want:   false,
		},
		{
			name:   "いずれの除外日付にも含まれない平日は true",
			target: time.Date(2024, time.March, 22, 0, 0, 0, 0, time.UTC), // 金曜日
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: 各日付に IsBusinessDay を呼ぶ
			got, err := bc.IsBusinessDay(ctx, tt.target)

			// Then: 期待する結果が返る（Requirement 4.4）
			if err != nil {
				t.Fatalf("IsBusinessDay returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsBusinessDay(%v) = %v, want %v", tt.target.Format(dateLayout), got, tt.want)
			}
		})
	}
}
