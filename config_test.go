package heijitu

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTempFile はテスト用の一時ファイルをディレクトリに作成してパスを返す。
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}

// TestLoadConfig_YAML_ValidFile は有効な .yaml ファイルから excluded_dates を正しく読み込めることを確認する。
func TestLoadConfig_YAML_ValidFile(t *testing.T) {
	// Given: 2件の excluded_dates を含む有効な YAML ファイル
	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.yaml", `
excluded_dates:
  - month: 1
    day: 1
  - month: 12
    day: 31
`)

	// When: .yaml ファイルを loadConfig で読み込む
	cfg, err := loadConfig(path)

	// Then: エラーなしで 2件の MonthDay が返る
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if len(cfg.ExcludedDates) != 2 {
		t.Fatalf("ExcludedDates: got %d items, want 2", len(cfg.ExcludedDates))
	}
	if cfg.ExcludedDates[0].Month != time.January || cfg.ExcludedDates[0].Day != 1 {
		t.Errorf("ExcludedDates[0]: got month=%v day=%d, want month=January day=1",
			cfg.ExcludedDates[0].Month, cfg.ExcludedDates[0].Day)
	}
	if cfg.ExcludedDates[1].Month != time.December || cfg.ExcludedDates[1].Day != 31 {
		t.Errorf("ExcludedDates[1]: got month=%v day=%d, want month=December day=31",
			cfg.ExcludedDates[1].Month, cfg.ExcludedDates[1].Day)
	}
}

// TestLoadConfig_YML_ValidFile は有効な .yml ファイルから excluded_dates を正しく読み込めることを確認する。
func TestLoadConfig_YML_ValidFile(t *testing.T) {
	// Given: 1件の excluded_dates を含む有効な YML ファイル
	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.yml", `
excluded_dates:
  - month: 8
    day: 15
`)

	// When: .yml ファイルを loadConfig で読み込む
	cfg, err := loadConfig(path)

	// Then: エラーなしで 1件の MonthDay が返る
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if len(cfg.ExcludedDates) != 1 {
		t.Fatalf("ExcludedDates: got %d items, want 1", len(cfg.ExcludedDates))
	}
	if cfg.ExcludedDates[0].Month != time.August || cfg.ExcludedDates[0].Day != 15 {
		t.Errorf("ExcludedDates[0]: got month=%v day=%d, want month=August day=15",
			cfg.ExcludedDates[0].Month, cfg.ExcludedDates[0].Day)
	}
}

// TestLoadConfig_JSON_ValidFile は有効な .json ファイルから excluded_dates を正しく読み込めることを確認する。
func TestLoadConfig_JSON_ValidFile(t *testing.T) {
	// Given: 2件の excluded_dates を含む有効な JSON ファイル
	// MonthDay の json:"month"/"day" タグにより小文字キーを使用する
	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.json", `{
  "excluded_dates": [
    {"month": 3, "day": 20},
    {"month": 9, "day": 23}
  ]
}`)

	// When: .json ファイルを loadConfig で読み込む
	cfg, err := loadConfig(path)

	// Then: エラーなしで 2件の MonthDay が返る
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if len(cfg.ExcludedDates) != 2 {
		t.Fatalf("ExcludedDates: got %d items, want 2", len(cfg.ExcludedDates))
	}
	if cfg.ExcludedDates[0].Month != time.March || cfg.ExcludedDates[0].Day != 20 {
		t.Errorf("ExcludedDates[0]: got month=%v day=%d, want month=March day=20",
			cfg.ExcludedDates[0].Month, cfg.ExcludedDates[0].Day)
	}
	if cfg.ExcludedDates[1].Month != time.September || cfg.ExcludedDates[1].Day != 23 {
		t.Errorf("ExcludedDates[1]: got month=%v day=%d, want month=September day=23",
			cfg.ExcludedDates[1].Month, cfg.ExcludedDates[1].Day)
	}
}

// TestLoadConfig_UnsupportedExtension はサポート外拡張子のファイルでエラーが返ることを確認する。
func TestLoadConfig_UnsupportedExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: ".toml 拡張子", filename: "config.toml"},
		{name: ".csv 拡張子", filename: "config.csv"},
		{name: ".txt 拡張子", filename: "config.txt"},
		{name: "拡張子なし", filename: "config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: サポート外拡張子のファイル
			dir := t.TempDir()
			path := writeTempFile(t, dir, tt.filename, "some content")

			// When: サポート外拡張子のパスを loadConfig に渡す
			_, err := loadConfig(path)

			// Then: エラーが返る（Requirement 6.4）
			if err == nil {
				t.Errorf("loadConfig(%q): expected error for unsupported extension, got nil", tt.filename)
			}
		})
	}
}

// TestLoadConfig_FileNotFound は存在しないファイルパスでエラーが返ることを確認する。
func TestLoadConfig_FileNotFound(t *testing.T) {
	// Given: 存在しないパス
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")

	// When: 存在しないパスを loadConfig に渡す
	_, err := loadConfig(path)

	// Then: os.ErrNotExist を含むエラーが返る（Requirement 6.4）
	if err == nil {
		t.Fatal("loadConfig: expected error for nonexistent file, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("loadConfig: expected os.ErrNotExist, got %v", err)
	}
}

// TestLoadConfig_InvalidYAMLContent は不正な YAML 内容でパースエラーが返ることを確認する。
func TestLoadConfig_InvalidYAMLContent(t *testing.T) {
	// Given: 不正な YAML 内容のファイル
	dir := t.TempDir()
	path := writeTempFile(t, dir, "invalid.yaml", `
excluded_dates:
  - month: [unclosed bracket
`)

	// When: 不正な YAML ファイルを loadConfig で読み込む
	_, err := loadConfig(path)

	// Then: パースエラーが返る（Requirement 6.5）
	if err == nil {
		t.Fatal("loadConfig: expected parse error for invalid YAML, got nil")
	}
}

// TestLoadConfig_InvalidJSONContent は不正な JSON 内容でパースエラーが返ることを確認する。
func TestLoadConfig_InvalidJSONContent(t *testing.T) {
	// Given: 不正な JSON 内容のファイル
	dir := t.TempDir()
	path := writeTempFile(t, dir, "invalid.json", `{
  "excluded_dates": [
    {"Month": 1, "Day": 1
  ]
}`)

	// When: 不正な JSON ファイルを loadConfig で読み込む
	_, err := loadConfig(path)

	// Then: パースエラーが返る（Requirement 6.5）
	if err == nil {
		t.Fatal("loadConfig: expected parse error for invalid JSON, got nil")
	}
}

// TestLoadConfig_EmptyExcludedDates は excluded_dates が空の YAML ファイルで空スライスが返ることを確認する。
func TestLoadConfig_EmptyExcludedDates(t *testing.T) {
	// Given: excluded_dates が空の有効な YAML ファイル
	dir := t.TempDir()
	path := writeTempFile(t, dir, "empty.yaml", `
excluded_dates: []
`)

	// When: excluded_dates が空の YAML ファイルを loadConfig で読み込む
	cfg, err := loadConfig(path)

	// Then: エラーなしで空スライスが返る
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if len(cfg.ExcludedDates) != 0 {
		t.Errorf("ExcludedDates: got %d items, want 0", len(cfg.ExcludedDates))
	}
}

// TestLoadConfig_MultipleEntries_YAML は複数件の excluded_dates を含む YAML ファイルで全件読み込めることを確認する。
func TestLoadConfig_MultipleEntries_YAML(t *testing.T) {
	// Given: 全月の代表日を含む YAML ファイル
	dir := t.TempDir()
	path := writeTempFile(t, dir, "multi.yaml", `
excluded_dates:
  - month: 1
    day: 1
  - month: 2
    day: 11
  - month: 3
    day: 21
  - month: 4
    day: 29
  - month: 5
    day: 3
`)

	// When: 5件の excluded_dates を含む YAML ファイルを loadConfig で読み込む
	cfg, err := loadConfig(path)

	// Then: エラーなしで 5件の MonthDay が返る
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if len(cfg.ExcludedDates) != 5 {
		t.Fatalf("ExcludedDates: got %d items, want 5", len(cfg.ExcludedDates))
	}
}
