package heijitu

import (
	"testing"
	"time"
)

// exampleYAMLPath は example ディレクトリの設定ファイルパス。
const exampleYAMLPath = "example/heijitu.yaml"

// --- loadConfig による YAML ファイルの読み込みテスト ---

// TestExampleYAML_Loadable は example/heijitu.yaml が loadConfig で読み込めることを確認する。
func TestExampleYAML_Loadable(t *testing.T) {
	// Given: example ディレクトリに配置された YAML 設定ファイル

	// When: loadConfig で読み込む
	cfg, err := loadConfig(exampleYAMLPath)

	// Then: エラーなしで設定が返る
	if err != nil {
		t.Fatalf("loadConfig(%q) returned error: %v", exampleYAMLPath, err)
	}
	if cfg == nil {
		t.Fatalf("loadConfig(%q) returned nil config", exampleYAMLPath)
	}
}

// TestExampleYAML_HasExpectedEntryCount は example/heijitu.yaml が期待件数の excluded_dates を含むことを確認する。
func TestExampleYAML_HasExpectedEntryCount(t *testing.T) {
	// Given: example ディレクトリに配置された YAML 設定ファイル

	// When: loadConfig で読み込む
	cfg, err := loadConfig(exampleYAMLPath)
	if err != nil {
		t.Fatalf("loadConfig(%q) returned error: %v", exampleYAMLPath, err)
	}

	// Then: 5件の除外日付（12/29, 12/30, 12/31, 1/2, 1/3）が含まれる
	const wantCount = 5
	if len(cfg.ExcludedDates) != wantCount {
		t.Fatalf("ExcludedDates: got %d items, want %d", len(cfg.ExcludedDates), wantCount)
	}
}

// TestExampleYAML_ExcludedDatesContent は example/heijitu.yaml の各除外日付が期待する月日であることを確認する。
func TestExampleYAML_ExcludedDatesContent(t *testing.T) {
	// Given: example ディレクトリに配置された YAML 設定ファイル
	cfg, err := loadConfig(exampleYAMLPath)
	if err != nil {
		t.Fatalf("loadConfig(%q) returned error: %v", exampleYAMLPath, err)
	}

	// Then: 年末年始の休業日が正しい順序で含まれる
	want := []MonthDay{
		{Month: time.December, Day: 29},
		{Month: time.December, Day: 30},
		{Month: time.December, Day: 31},
		{Month: time.January, Day: 2},
		{Month: time.January, Day: 3},
	}

	if len(cfg.ExcludedDates) != len(want) {
		t.Fatalf("ExcludedDates: got %d items, want %d", len(cfg.ExcludedDates), len(want))
	}

	for i, w := range want {
		got := cfg.ExcludedDates[i]
		if got.Month != w.Month || got.Day != w.Day {
			t.Errorf("ExcludedDates[%d]: got month=%v day=%d, want month=%v day=%d",
				i, got.Month, got.Day, w.Month, w.Day)
		}
	}
}

// --- WithConfig（公開API）による統合テスト ---

// TestExampleYAML_WithConfig_Loadable は WithConfig が example/heijitu.yaml をエラーなしで読み込めることを確認する。
func TestExampleYAML_WithConfig_Loadable(t *testing.T) {
	// Given: example ディレクトリに配置された YAML 設定ファイル

	// When: WithConfig で読み込む
	opt, err := WithConfig(exampleYAMLPath)

	// Then: エラーなしでオプションが返る
	if err != nil {
		t.Fatalf("WithConfig(%q) returned error: %v", exampleYAMLPath, err)
	}
	if opt == nil {
		t.Fatalf("WithConfig(%q) returned nil option", exampleYAMLPath)
	}
}
