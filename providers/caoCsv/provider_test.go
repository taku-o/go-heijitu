package caoCsv_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
	caoCsv "github.com/taku-o/go-heijitu/providers/caoCsv"
)

// コンパイル時のインターフェース充足チェック。
// このファイルがコンパイルされることで Provider が HolidayProvider を満たすことを保証する。
var _ heijitu.HolidayProvider = (*caoCsv.Provider)(nil)

// testdataDir はテストフィクスチャのディレクトリパスを返す。
func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata")
}

// fixtureCSVPath はテスト用の Shift_JIS CSVファイルのパスを返す。
func fixtureCSVPath() string {
	return filepath.Join(testdataDir(), "syukujitsu_test.csv")
}

// fixtureHolidayCount はフィクスチャ内の祝日行数（ヘッダ行を除く）。
// フィクスチャの内容と一致させること。
const fixtureHolidayCount = 5

// --- New ---

// TestNew_ValidCSVPath は有効なCSVパスを指定した場合にエラーなしでプロバイダーが得られることを確認する。
func TestNew_ValidCSVPath(t *testing.T) {
	// Given: テストフィクスチャの有効なCSVパス
	ctx := context.Background()
	opts := caoCsv.Options{CSVPath: fixtureCSVPath()}

	// When: New を呼ぶ
	p, err := caoCsv.New(ctx, opts)

	// Then: エラーなしでプロバイダーが得られる
	if err != nil {
		t.Fatalf("New(valid CSV) returned error: %v", err)
	}
	if p == nil {
		t.Fatal("New(valid CSV) returned nil provider")
	}
}

// TestNew_NonExistentPath は存在しないパスを指定した場合にエラーが返ることを確認する。
func TestNew_NonExistentPath(t *testing.T) {
	// Given: 存在しないCSVパス
	ctx := context.Background()
	opts := caoCsv.Options{CSVPath: "/non/existent/path.csv"}

	// When: New を呼ぶ
	p, err := caoCsv.New(ctx, opts)

	// Then: エラーが返り、プロバイダーは nil
	if err == nil {
		t.Fatal("New(non-existent path) returned nil error, want error")
	}
	if p != nil {
		t.Errorf("New(non-existent path) returned non-nil provider")
	}
}

// --- IsHoliday ---

// TestIsHoliday_KnownHoliday は既知の祝日に対して IsHoliday が true を返すことを確認する。
func TestIsHoliday_KnownHoliday(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダーと既知の祝日（元日）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestIsHoliday_NonHoliday は祝日でない日付に対して IsHoliday が false を返すことを確認する。
func TestIsHoliday_NonHoliday(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダーと平日（2025-01-06）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidayName_KnownHoliday は既知の祝日に対して正しい祝日名を返し、文字化けしないことを確認する。
// Shift_JIS デコードの間接検証を兼ねる。
func TestHolidayName_KnownHoliday(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダーと既知の祝日（元日）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidayName_MultiByteCharacters は複数バイト文字を含む祝日名が文字化けなく返ることを確認する。
func TestHolidayName_MultiByteCharacters(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダーと「建国記念の日」
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	target := time.Date(2025, time.February, 11, 0, 0, 0, 0, time.UTC)

	// When: 祝日の日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: "建国記念の日" と完全一致し文字化けしない
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got != "建国記念の日" {
		t.Errorf("HolidayName(%s) = %q, want %q", target.Format(time.DateOnly), got, "建国記念の日")
	}
}

// TestHolidayName_NonHoliday は祝日でない日付に対して空文字と nil error を返すことを確認する。
func TestHolidayName_NonHoliday(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダーと平日（2025-01-06）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidaysBetween_IncludesEndpoints は HolidaysBetween が両端の日付を含む祝日を昇順で返すことを確認する。
func TestHolidaysBetween_IncludesEndpoints(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、from=元日（2025-01-01）、to=成人の日（2025-01-13）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidaysBetween_AscendingOrder は HolidaysBetween が日付昇順でソートされた結果を返すことを確認する。
func TestHolidaysBetween_AscendingOrder(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、全祝日を含む広い期間
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)

	// When: 全祝日を含む期間で HolidaysBetween を呼ぶ
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

// TestHolidaysBetween_HeaderNotContaminated はフィクスチャの全祝日を含む範囲で取得した件数が
// フィクスチャの祝日行数（ヘッダ行を除く）と一致し、ヘッダ行が混入していないことを確認する。
func TestHolidaysBetween_HeaderNotContaminated(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、全祝日を含む十分広い期間
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)

	// When: フィクスチャの全祝日を含む十分広い期間で HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, from, to)

	// Then: 返却件数がフィクスチャの祝日行数と一致（ヘッダ行が混入していない）
	if err != nil {
		t.Fatalf("HolidaysBetween returned error: %v", err)
	}
	if len(got) != fixtureHolidayCount {
		t.Errorf("HolidaysBetween: got %d holidays, want %d (fixture holiday count excluding header)",
			len(got), fixtureHolidayCount)
	}
}

// TestHolidaysBetween_FromAfterTo は from > to の場合に空スライスと nil error が返ることを確認する。
func TestHolidaysBetween_FromAfterTo(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、from > to（逆順）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidaysBetween_FromEqualsTo_Holiday は from == to で祝日の場合に1件が返ることを確認する。
func TestHolidaysBetween_FromEqualsTo_Holiday(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、from == to == 元日（2025-01-01）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidaysBetween_FromEqualsTo_NonHoliday は from == to で非祝日の場合に0件が返ることを確認する。
func TestHolidaysBetween_FromEqualsTo_NonHoliday(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、from == to == 平日（2025-01-06）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
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

// TestHolidaysBetween_DifferentLocation は異なる Location でも同一暦日範囲なら同じ祝日集合が返ることを確認する。
func TestHolidaysBetween_DifferentLocation(t *testing.T) {
	// Given: フィクスチャから作成したプロバイダー、UTC と JST で同一暦日範囲を指定
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: fixtureCSVPath()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	jst := time.FixedZone("JST", 9*60*60)

	fromUTC := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	toUTC := time.Date(2025, time.February, 24, 0, 0, 0, 0, time.UTC)
	fromJST := time.Date(2025, time.January, 1, 0, 0, 0, 0, jst)
	toJST := time.Date(2025, time.February, 24, 0, 0, 0, 0, jst)

	// When: UTC と JST で同じ暦日範囲を指定
	gotUTC, err := p.HolidaysBetween(ctx, fromUTC, toUTC)
	if err != nil {
		t.Fatalf("HolidaysBetween(UTC) returned error: %v", err)
	}
	gotJST, err := p.HolidaysBetween(ctx, fromJST, toJST)
	if err != nil {
		t.Fatalf("HolidaysBetween(JST) returned error: %v", err)
	}

	// Then: 同じ件数の祝日が返る
	if len(gotUTC) != len(gotJST) {
		t.Errorf("HolidaysBetween: UTC returned %d holidays, JST returned %d holidays, want same count",
			len(gotUTC), len(gotJST))
	}
	// 各祝日の暦日（Y/M/D）が一致する
	for i := 0; i < len(gotUTC) && i < len(gotJST); i++ {
		utcDate := gotUTC[i].Date
		jstDate := gotJST[i].Date
		if utcDate.Year() != jstDate.Year() || utcDate.Month() != jstDate.Month() || utcDate.Day() != jstDate.Day() {
			t.Errorf("HolidaysBetween: holiday[%d] UTC=%s JST=%s, want same calendar date",
				i, utcDate.Format(time.DateOnly), jstDate.Format(time.DateOnly))
		}
	}
}
