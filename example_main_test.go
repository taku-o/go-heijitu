package heijitu_test

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// --- example/main.go の実行（1回だけ実行して結果を共有） ---

var (
	exampleOnce   sync.Once
	exampleOutput string
	exampleErr    error
)

// runExampleOnce は go run example/main.go を GOOGLE_CALENDAR_API_KEY 未設定で1回だけ実行し、
// 結合出力と実行エラーをキャッシュして返す。複数テストでビルド/実行/ネットワーク取得を共有する。
func runExampleOnce() (string, error) {
	exampleOnce.Do(func() {
		cmd := exec.Command("go", "run", "example/main.go")
		cmd.Env = filterEnv("GOOGLE_CALENDAR_API_KEY")
		out, err := cmd.CombinedOutput()
		exampleOutput = string(out)
		exampleErr = err
	})
	return exampleOutput, exampleErr
}

// runExampleMain は共有実行結果の出力を返す。実行失敗時はテストを Fatal する。
func runExampleMain(t *testing.T) string {
	t.Helper()
	out, err := runExampleOnce()
	if err != nil {
		t.Fatalf("go run example/main.go failed: %v\noutput:\n%s", err, out)
	}
	return out
}

// filterEnv は現在のプロセス環境変数から指定キーを除外した環境変数スライスを返す。
func filterEnv(excludeKey string) []string {
	var filtered []string
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, excludeKey+"=") {
			filtered = append(filtered, env)
		}
	}
	return filtered
}

// requireContains は output に substr が含まれることを検証する。
// 含まれない場合は t.Errorf で報告する。
func requireContains(t *testing.T, output, substr, msg string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("%s: output does not contain %q", msg, substr)
	}
}

// --- example/main.go の実行テスト ---

// TestExampleMain_ExitZero は GOOGLE_CALENDAR_API_KEY 未設定で
// go run example/main.go が exit 0 で正常終了することを確認する。
func TestExampleMain_ExitZero(t *testing.T) {
	// Given/When: GOOGLE_CALENDAR_API_KEY 未設定で example を1回実行する
	out, err := runExampleOnce()

	// Then: exit 0 で正常終了する（要件 1.8）
	if err != nil {
		t.Fatalf("go run example/main.go failed with error: %v\noutput:\n%s", err, out)
	}
}

// TestExampleMain_HolidayjpOutput は holidayjp プロバイダーの各API結果が
// 出力に含まれることを確認する。
func TestExampleMain_HolidayjpOutput(t *testing.T) {
	// Given: GOOGLE_CALENDAR_API_KEY 未設定で example を実行
	output := runExampleMain(t)

	// Then: holidayjp セクションの出力に各APIの結果が含まれる（要件 1.2）
	requireContains(t, output, "IsBusinessDay", "holidayjp の IsBusinessDay 結果が出力に含まれる")
	requireContains(t, output, "NextBusinessDay", "holidayjp の NextBusinessDay 結果が出力に含まれる")
	requireContains(t, output, "FirstBusinessDayOfMonth", "holidayjp の FirstBusinessDayOfMonth 結果が出力に含まれる")
	requireContains(t, output, "FirstBusinessDaysOfYear", "holidayjp の FirstBusinessDaysOfYear 結果が出力に含まれる")
	requireContains(t, output, "Holidays", "holidayjp の Holidays 結果が出力に含まれる")
}

// TestExampleMain_ExcludedDatesOutput は WithExcludedDates + WithConfig 併用による
// 除外日付の効果が出力に含まれることを確認する。
func TestExampleMain_ExcludedDatesOutput(t *testing.T) {
	// Given: GOOGLE_CALENDAR_API_KEY 未設定で example を実行
	output := runExampleMain(t)

	// Then: 除外日付の効果を示す出力が含まれる（要件 1.3）
	requireContains(t, output, "WithConfig", "WithConfig 併用例の出力が含まれる")
	requireContains(t, output, "WithExcludedDates", "WithExcludedDates 併用例の出力が含まれる")
}

// TestExampleMain_ExtraExcludedOutput は extraExcluded の呼び出し例が
// 出力に含まれることを確認する。
func TestExampleMain_ExtraExcludedOutput(t *testing.T) {
	// Given: GOOGLE_CALENDAR_API_KEY 未設定で example を実行
	output := runExampleMain(t)

	// Then: extraExcluded の効果を示す出力が含まれる（要件 1.4）
	requireContains(t, output, "extraExcluded", "extraExcluded 呼び出し例の出力が含まれる")
}

// TestExampleMain_CaoCsvLocalOutput は caoCsv ローカルCSV モードの結果が
// 出力に含まれることを確認する。
func TestExampleMain_CaoCsvLocalOutput(t *testing.T) {
	// Given: GOOGLE_CALENDAR_API_KEY 未設定で example を実行
	output := runExampleMain(t)

	// Then: caoCsv ローカルCSV の結果が含まれる（要件 1.5）
	requireContains(t, output, "caoCsv", "caoCsv ローカルCSV の出力が含まれる")
}

// TestExampleMain_GoogleCalendarSkipOutput は GOOGLE_CALENDAR_API_KEY 未設定時に
// googleCalendar のスキップ表示が出力に含まれることを確認する。
func TestExampleMain_GoogleCalendarSkipOutput(t *testing.T) {
	// Given: GOOGLE_CALENDAR_API_KEY 未設定で example を実行
	output := runExampleMain(t)

	// Then: googleCalendar のスキップメッセージが含まれる（要件 1.7）
	requireContains(t, output, "GOOGLE_CALENDAR_API_KEY", "googleCalendar スキップ表示が出力に含まれる")
}
