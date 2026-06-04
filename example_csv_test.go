package heijitu_test

import (
	"context"
	"os"
	"testing"
	"time"

	caoCsv "github.com/taku-o/go-heijitu/providers/caoCsv"
)

// exampleCSVPath は example ディレクトリの caoCsv ローカルモード用CSVパス。
const exampleCSVPath = "example/testdata/syukujitsu.csv"

// sourceCSVPath はコピー元のテストフィクスチャCSVパス。
const sourceCSVPath = "providers/caoCsv/testdata/syukujitsu_test.csv"

// --- CSV ファイルの読み込みテスト ---

// TestExampleCSV_LoadableByCaoCsv は example/testdata/syukujitsu.csv が caoCsv.New で読み込めることを確認する。
func TestExampleCSV_LoadableByCaoCsv(t *testing.T) {
	// Given: example ディレクトリに配置された Shift_JIS CSV ファイル
	ctx := context.Background()
	opts := caoCsv.Options{CSVPath: exampleCSVPath}

	// When: caoCsv.New で読み込む
	p, err := caoCsv.New(ctx, opts)

	// Then: エラーなしでプロバイダーが返る
	if err != nil {
		t.Fatalf("caoCsv.New(CSVPath=%q) returned error: %v", exampleCSVPath, err)
	}
	if p == nil {
		t.Fatal("caoCsv.New returned nil provider")
	}
}

// TestExampleCSV_KnownHolidayRecognized は example CSV から作成したプロバイダーが既知の祝日を認識することを確認する。
func TestExampleCSV_KnownHolidayRecognized(t *testing.T) {
	// Given: example CSV から作成したプロバイダーと元日（2025-01-01）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: exampleCSVPath})
	if err != nil {
		t.Fatalf("caoCsv.New returned error: %v", err)
	}
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 元日に対して IsHoliday を呼ぶ
	got, err := p.IsHoliday(ctx, target)

	// Then: true が返る（元日は祝日）
	if err != nil {
		t.Fatalf("IsHoliday(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if !got {
		t.Errorf("IsHoliday(%s) = false, want true", target.Format(time.DateOnly))
	}
}

// TestExampleCSV_HolidayNameNotGarbled は example CSV から取得した祝日名が文字化けしないことを確認する。
// Shift_JIS エンコーディングの正しいデコードを間接的に検証する。
func TestExampleCSV_HolidayNameNotGarbled(t *testing.T) {
	// Given: example CSV から作成したプロバイダーと元日（2025-01-01）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: exampleCSVPath})
	if err != nil {
		t.Fatalf("caoCsv.New returned error: %v", err)
	}
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// When: 元日に対して HolidayName を呼ぶ
	got, err := p.HolidayName(ctx, target)

	// Then: "元日" と完全一致し文字化けしない
	if err != nil {
		t.Fatalf("HolidayName(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got != "元日" {
		t.Errorf("HolidayName(%s) = %q, want %q", target.Format(time.DateOnly), got, "元日")
	}
}

// TestExampleCSV_NonHolidayNotRecognized は example CSV から作成したプロバイダーが非祝日を正しく false で返すことを確認する。
func TestExampleCSV_NonHolidayNotRecognized(t *testing.T) {
	// Given: example CSV から作成したプロバイダーと平日（2025-01-06）
	ctx := context.Background()
	p, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: exampleCSVPath})
	if err != nil {
		t.Fatalf("caoCsv.New returned error: %v", err)
	}
	target := time.Date(2025, time.January, 6, 0, 0, 0, 0, time.UTC)

	// When: 平日に対して IsHoliday を呼ぶ
	got, err := p.IsHoliday(ctx, target)

	// Then: false が返る
	if err != nil {
		t.Fatalf("IsHoliday(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if got {
		t.Errorf("IsHoliday(%s) = true, want false", target.Format(time.DateOnly))
	}
}

// --- CSV ファイルのバイナリ同一性テスト ---

// TestExampleCSV_BinaryIdenticalToSource は example CSV がコピー元のフィクスチャCSVとバイナリ同一であることを確認する。
// Shift_JIS エンコーディングがテキスト変換で壊れていないことを保証する。
func TestExampleCSV_BinaryIdenticalToSource(t *testing.T) {
	// Given: コピー元のフィクスチャCSVと example CSV の両方のバイト列
	sourceData, err := os.ReadFile(sourceCSVPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", sourceCSVPath, err)
	}
	exampleData, err := os.ReadFile(exampleCSVPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", exampleCSVPath, err)
	}

	// Then: バイト列が完全に一致する
	if len(sourceData) != len(exampleData) {
		t.Fatalf("file size mismatch: source=%d bytes, example=%d bytes", len(sourceData), len(exampleData))
	}
	for i := range sourceData {
		if sourceData[i] != exampleData[i] {
			t.Fatalf("byte mismatch at offset %d: source=0x%02x, example=0x%02x", i, sourceData[i], exampleData[i])
		}
	}
}
