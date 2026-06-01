package heijitu_test

import (
	"context"
	"testing"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
)

// dateLayout は testProvider のマップキーと Format 呼び出しで使用する日付フォーマット。
const dateLayout = "2006-01-02"

// testProvider は HolidayProvider インターフェースを満たすテスト専用の実装。
// このファイルのコンパイルが成功することで、インターフェース定義の正しさを保証する。
// err フィールドが非 nil の場合、各メソッドがそのエラーを返す。
type testProvider struct {
	holidays map[string]string // dateLayout キー → 祝日名
	err      error
}

// コンパイル時のインターフェース充足チェック。
var _ heijitu.HolidayProvider = (*testProvider)(nil)

func (p *testProvider) IsHoliday(ctx context.Context, t time.Time) (bool, error) {
	if p.err != nil {
		return false, p.err
	}
	_, ok := p.holidays[t.Format(dateLayout)]
	return ok, nil
}

func (p *testProvider) HolidayName(ctx context.Context, t time.Time) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	name := p.holidays[t.Format(dateLayout)]
	return name, nil
}

func (p *testProvider) HolidaysBetween(ctx context.Context, from, to time.Time) ([]heijitu.Holiday, error) {
	if p.err != nil {
		return nil, p.err
	}
	if from.After(to) {
		return []heijitu.Holiday{}, nil
	}
	var result []heijitu.Holiday
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format(dateLayout)
		if name, ok := p.holidays[key]; ok {
			result = append(result, heijitu.Holiday{Date: d, Name: name})
		}
	}
	return result, nil
}

// TestHolidayProvider_IsHoliday_HolidayDate は祝日の日付に対して IsHoliday が true を返すことを確認する。
func TestHolidayProvider_IsHoliday_HolidayDate(t *testing.T) {
	// Given: 元日が祝日として登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	ctx := context.Background()
	target := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 祝日の日付に対して IsHoliday を呼ぶ
	got, err := p.IsHoliday(ctx, target)

	// Then: true と nil error が返る（Requirement 3.2）
	if err != nil {
		t.Fatalf("IsHoliday returned error: %v", err)
	}
	if !got {
		t.Errorf("IsHoliday(%v) = false, want true", target.Format(dateLayout))
	}
}

// TestHolidayProvider_IsHoliday_NonHolidayDate は祝日でない日付に対して IsHoliday が false を返すことを確認する。
func TestHolidayProvider_IsHoliday_NonHolidayDate(t *testing.T) {
	// Given: 元日のみ祝日として登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	ctx := context.Background()
	target := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)

	// When: 祝日でない日付に対して IsHoliday を呼ぶ
	got, err := p.IsHoliday(ctx, target)

	// Then: false と nil error が返る（Requirement 3.3）
	if err != nil {
		t.Fatalf("IsHoliday returned error: %v", err)
	}
	if got {
		t.Errorf("IsHoliday(%v) = true, want false", target.Format(dateLayout))
	}
}

// TestHolidayProvider_HolidayName_HolidayDate は祝日の日付に対して HolidayName が非空文字列を返すことを確認する。
func TestHolidayProvider_HolidayName_HolidayDate(t *testing.T) {
	// Given: 元日が登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	ctx := context.Background()
	target := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 祝日の日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: 祝日名と nil error が返る（Requirement 3.4）
	if err != nil {
		t.Fatalf("HolidayName returned error: %v", err)
	}
	if got != "元日" {
		t.Errorf("HolidayName(%v) = %q, want %q", target.Format(dateLayout), got, "元日")
	}
}

// TestHolidayProvider_HolidayName_NonHolidayDate は祝日でない日付に対して HolidayName が空文字を返すことを確認する。
func TestHolidayProvider_HolidayName_NonHolidayDate(t *testing.T) {
	// Given: 元日のみ登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	ctx := context.Background()
	target := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)

	// When: 祝日でない日付に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: 空文字と nil error が返る（Requirement 3.5）
	if err != nil {
		t.Fatalf("HolidayName returned error: %v", err)
	}
	if got != "" {
		t.Errorf("HolidayName(%v) = %q, want empty string", target.Format(dateLayout), got)
	}
}

// TestHolidayProvider_HolidaysBetween_IncludesEndpoints は HolidaysBetween が両端の日付を含むことを確認する。
func TestHolidayProvider_HolidaysBetween_IncludesEndpoints(t *testing.T) {
	// Given: from と to 当日が祝日として登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
			"2024-01-02": "振替休日",
			"2024-01-03": "三が日最終日",
		},
	}
	ctx := context.Background()
	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)

	// When: from〜to の範囲で HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, from, to)

	// Then: 両端を含む3件が返る（Requirement 3.6）
	if err != nil {
		t.Fatalf("HolidaysBetween returned error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("HolidaysBetween: got %d holidays, want 3", len(got))
	}
}

// TestHolidayProvider_HolidaysBetween_FromAfterTo は from > to の場合に空スライスと nil error が返ることを確認する。
func TestHolidayProvider_HolidaysBetween_FromAfterTo(t *testing.T) {
	// Given: 祝日が登録されたプロバイダー
	p := &testProvider{
		holidays: map[string]string{
			"2024-01-01": "元日",
		},
	}
	ctx := context.Background()
	from := time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC) // from > to

	// When: from が to より後の場合に HolidaysBetween を呼ぶ
	got, err := p.HolidaysBetween(ctx, from, to)

	// Then: 空スライスと nil error が返る（Requirement 3.8）
	if err != nil {
		t.Fatalf("HolidaysBetween returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("HolidaysBetween(from>to): got %d holidays, want 0", len(got))
	}
}
