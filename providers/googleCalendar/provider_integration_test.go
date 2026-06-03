//go:build integration

package googleCalendar_test

import (
	"context"
	"os"
	"testing"
	"time"

	googleCalendar "github.com/taku-o/go-heijitu/providers/googleCalendar"
)

// apiKeyProvider は環境変数 GOOGLE_CALENDAR_API_KEY からAPIキーを取得し、
// googleCalendar プロバイダーを構築して返す。未設定の場合は t.Skip でスキップする。
func apiKeyProvider(t *testing.T) *googleCalendar.Provider {
	t.Helper()
	apiKey := os.Getenv("GOOGLE_CALENDAR_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_CALENDAR_API_KEY is not set; skipping integration test")
	}
	ctx := context.Background()
	p, err := googleCalendar.New(ctx, googleCalendar.Options{APIKey: apiKey})
	if err != nil {
		t.Fatalf("New(APIKey) returned error: %v", err)
	}
	return p
}

// --- New ---

// TestNew_APIKey は有効なAPIキーでプロバイダーが構築できることを確認する（要件 1.1, 1.3, 2.1）。
func TestNew_APIKey(t *testing.T) {
	// Given: 環境変数から取得したAPIキー
	apiKey := os.Getenv("GOOGLE_CALENDAR_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_CALENDAR_API_KEY is not set; skipping integration test")
	}
	ctx := context.Background()
	opts := googleCalendar.Options{APIKey: apiKey}

	// When: New を呼ぶ
	p, err := googleCalendar.New(ctx, opts)

	// Then: エラーなしでプロバイダーが得られる
	if err != nil {
		t.Fatalf("New(APIKey) returned error: %v", err)
	}
	if p == nil {
		t.Fatal("New(APIKey) returned nil provider")
	}
}

// --- IsHoliday ---

// TestIsHoliday_KnownHoliday は既知の祝日（元日）に対して IsHoliday が true を返すことを確認する（要件 5.1, 4.1, 4.2）。
func TestIsHoliday_KnownHoliday(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダーと既知の祝日（元日）
	p := apiKeyProvider(t)
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

// TestIsHoliday_Weekday は平日の日付に対して IsHoliday が false を返すことを確認する（要件 5.2）。
func TestIsHoliday_Weekday(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダーと平日（2025-01-06 月曜）
	p := apiKeyProvider(t)
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

// --- HolidayName ---

// TestHolidayName_KnownHoliday は既知の祝日に対して正しい祝日名を返し、文字化けしないことを確認する（要件 5.3）。
func TestHolidayName_KnownHoliday(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダーと既知の祝日（元日）
	p := apiKeyProvider(t)
	ctx := context.Background()
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 祝日の日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: 祝日名 "元日" と完全一致し、nil error が返る
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got != "元日" {
		t.Errorf("HolidayName(%s) = %q, want %q", target.Format(time.DateOnly), got, "元日")
	}
}

// TestHolidayName_NonHoliday は非祝日の日付に対して空文字と nil error を返すことを確認する（要件 5.4）。
func TestHolidayName_NonHoliday(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダーと平日（2025-01-06）
	p := apiKeyProvider(t)
	ctx := context.Background()
	target := time.Date(2025, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: 平日の日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: 空文字と nil error が返る
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got != "" {
		t.Errorf("HolidayName(%s) = %q, want empty string", target.Format(time.DateOnly), got)
	}
}

// --- HolidaysBetween ---

// TestHolidaysBetween_IncludesEndpoints は HolidaysBetween が両端の日付を含む祝日を昇順で返すことを確認する（要件 5.5）。
func TestHolidaysBetween_IncludesEndpoints(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダー、from=元日（2025-01-01）、to=成人の日（2025-01-13）
	p := apiKeyProvider(t)
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
	if got[0].Date.Year() != wantFirst.Year() || got[0].Date.Month() != wantFirst.Month() || got[0].Date.Day() != wantFirst.Day() {
		t.Errorf("HolidaysBetween: got[0].Date = %s, want %s", got[0].Date.Format(time.DateOnly), wantFirst.Format(time.DateOnly))
	}
	wantSecond := time.Date(2025, time.January, 13, 0, 0, 0, 0, time.UTC)
	if got[1].Date.Year() != wantSecond.Year() || got[1].Date.Month() != wantSecond.Month() || got[1].Date.Day() != wantSecond.Day() {
		t.Errorf("HolidaysBetween: got[1].Date = %s, want %s", got[1].Date.Format(time.DateOnly), wantSecond.Format(time.DateOnly))
	}
}

// TestHolidaysBetween_AscendingOrder は HolidaysBetween が日付昇順でソートされた結果を返すことを確認する（要件 5.5）。
func TestHolidaysBetween_AscendingOrder(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダー、2025年の1月〜3月
	p := apiKeyProvider(t)
	ctx := context.Background()
	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC)

	// When: 複数の祝日を含む期間で HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, from, to)

	// Then: 結果が日付昇順でソートされている
	if err != nil {
		t.Fatalf("HolidaysBetween returned error: %v", err)
	}
	for i := 1; i < len(got); i++ {
		if got[i].Date.Before(got[i-1].Date) {
			t.Errorf("HolidaysBetween: got[%d].Date (%s) is before got[%d].Date (%s), want ascending order",
				i, got[i].Date.Format(time.DateOnly), i-1, got[i-1].Date.Format(time.DateOnly))
		}
	}
}

// TestHolidaysBetween_FromAfterTo は from > to の場合に空スライスと nil error が返ることを確認する。
// ガードが API 呼び出し前に短絡し、holidayjp/caoCsv との挙動が一致することを検証する。
func TestHolidaysBetween_FromAfterTo(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダー、from > to（逆順）
	p := apiKeyProvider(t)
	ctx := context.Background()
	from := time.Date(2025, time.February, 24, 0, 0, 0, 0, time.UTC)
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

// TestHolidaysBetween_FromEqualsTo_Holiday は from == to で祝日の場合に1件が返ることを確認する（要件 5.5 境界条件）。
func TestHolidaysBetween_FromEqualsTo_Holiday(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダー、from == to == 元日（2025-01-01）
	p := apiKeyProvider(t)
	ctx := context.Background()
	day := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: from == to で祝日の日付を指定して HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, day, day)

	// Then: 1件の祝日が返る
	if err != nil {
		t.Fatalf("HolidaysBetween(from==to, holiday) returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("HolidaysBetween(from==to, holiday): got %d holidays, want 1", len(got))
	}
	if got[0].Name != "元日" {
		t.Errorf("HolidaysBetween(from==to, holiday): got[0].Name = %q, want %q", got[0].Name, "元日")
	}
}

// TestHolidaysBetween_FromEqualsTo_NonHoliday は from == to で非祝日の場合に0件が返ることを確認する（要件 5.5 境界条件）。
func TestHolidaysBetween_FromEqualsTo_NonHoliday(t *testing.T) {
	// Given: APIキー認証で構築したプロバイダー、from == to == 平日（2025-01-06）
	p := apiKeyProvider(t)
	ctx := context.Background()
	day := time.Date(2025, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: from == to で非祝日の日付を指定して HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, day, day)

	// Then: 0件が返る
	if err != nil {
		t.Fatalf("HolidaysBetween(from==to, non-holiday) returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("HolidaysBetween(from==to, non-holiday): got %d holidays, want 0", len(got))
	}
}
