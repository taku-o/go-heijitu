# 作業計画: go-heijitu 実装

## 全体概要

5つの作業ステップに分けて実装を進める。  
各ステップは独立して動作確認できる単位とし、前ステップの成果物を次ステップが利用する形で積み上げる。

---

## ステップ一覧

| ステップ | 内容 | 成果物 |
|---------|------|--------|
| Step 1 | プロジェクト初期化 + コア実装 | コア型・インターフェース・BusinessCalendar骨格 |
| Step 2 | holidayjp プロバイダー + 全API実装 | 全API動作・テスト（holidayjpベース） |
| Step 3 | caoCsv プロバイダー実装 | 内閣府CSV対応・テスト |
| Step 4 | googleCalendar プロバイダー実装 | Google Calendar API対応・テスト |
| Step 5 | example + ドキュメント整備 | README(en/ja)・GoDoc・使い方ガイド(en/ja)・API仕様(en/ja)・プロバイダーガイド(en/ja)・example・最終確認 |

---

## Step 1: プロジェクト初期化 + コア実装

### 目的
ライブラリの骨格を作り、`IsBusinessDay` の判定ロジックまでを実装する。

### 作業内容

**プロジェクト初期化**
- `go.mod` の作成（`module github.com/taku-o/go-heijitu`、`go 1.23`）
- `.gitignore` の確認

**コア型の実装**
- `monthday.go`: `MonthDay` 型 + `Matches(t time.Time) bool` メソッド
- `holiday.go`: `Holiday` 型（Date, Name）

**インターフェースの定義**
- `provider.go`: `HolidayProvider` インターフェース（IsHoliday / HolidayName / HolidaysBetween）

**BusinessCalendar の骨格**
- `calendar.go`: `BusinessCalendar` 構造体 + `New()` コンストラクタ
- `calendar.go`: `IsBusinessDay()` の判定ロジック（土日判定 + プロバイダー呼び出し + 除外日付チェック）

**オプション・設定ファイルの実装**
- `option.go`: `Option` 型 + `WithExcludedDates()` + `WithConfig()`
- `config.go`: YAML/JSON設定ファイルの読み込み（拡張子で自動判別）

**テスト**
- `monthday_test.go`: `Matches()` の動作確認
- `config_test.go`: YAML・JSON設定ファイルの読み込み確認
- `calendar_test.go`: `IsBusinessDay()` のテスト（モックプロバイダーを使用）

### 完了条件
- `go build ./...` が通る
- モックプロバイダーを使い `IsBusinessDay()` のテストが全て通る

---

## Step 2: holidayjp プロバイダー + 全API実装

### 目的
デフォルトプロバイダー（holiday_jp-go）を実装し、BusinessCalendar の全APIを完成させる。

### 作業内容

**holidayjp プロバイダーの実装**
- `providers/holidayjp/provider.go`: `github.com/holiday-jp/holiday_jp-go` を使った `HolidayProvider` 実装
- `go get github.com/holiday-jp/holiday_jp-go`

**BusinessCalendar 残りAPIの実装**
- `calendar.go`: `NextBusinessDay()` — 翌日以降で最初の営業日を探索
- `calendar.go`: `FirstBusinessDayOfMonth()` — 指定年月の1日から営業日を探索
- `calendar.go`: `FirstBusinessDaysOfYear()` — 12ヶ月分を `FirstBusinessDayOfMonth()` で集約
- `calendar.go`: `Holidays()` — プロバイダーの `HolidaysBetween()` を呼び出して返す

**テスト**
- `providers/holidayjp/provider_test.go`: IsHoliday / HolidayName / HolidaysBetween の動作確認
- `calendar_test.go` に追記:
  - `NextBusinessDay()`: 金曜日の翌営業日が月曜日になること、祝日をスキップすること
  - `FirstBusinessDayOfMonth()`: 元旦が1月1日の場合に翌営業日を返すこと
  - `FirstBusinessDaysOfYear()`: 12件返ること
  - `Holidays()`: 指定期間の祝日リストが正しいこと
  - 除外日付（WithExcludedDates / WithConfig）が正しく機能すること

### 完了条件
- `go test ./...` が全て通る
- holidayjp プロバイダーを使って全APIが期待通りに動作する

---

## Step 3: caoCsv プロバイダー実装

### 目的
内閣府公式CSVを使った `HolidayProvider` 実装を追加する。

### 作業内容

**caoCsv プロバイダーの実装**
- `providers/caoCsv/provider.go`: `Options`（CSVPath）を受け取る `New()`
- ローカルCSVファイルの読み込みモード（`CSVPath` 指定時 → `syukujitsu.LoadAndParse`）
- オンライン取得モード（`CSVPath` 空時 → `syukujitsu.FetchAndParse`、内閣府固定URL）
- CSV取得・Shift_JIS デコード・パースは `github.com/mikan/syukujitsu-go` に委譲
- パース結果（`[]Entry`）を保持。点照合（IsHoliday/HolidayName）は mikan `Find` に委譲、`HolidaysBetween` のみ自前で範囲フィルタ＋昇順ソート
- `go get github.com/mikan/syukujitsu-go`

**テスト**
- `providers/caoCsv/provider_test.go`:
  - ローカルCSVファイルを使った IsHoliday / HolidayName / HolidaysBetween の動作確認
  - テスト用に最小限のCSVファイル（`testdata/syukujitsu_test.csv`、Shift_JIS）を用意する
  - `CSVPath` 指定時にローカルファイルが読み込まれること
  - オンライン取得モード（`FetchAndParse`）はネットワーク依存のため、通常テストでの扱いを設計時に決める

### 完了条件
- `go test ./providers/caoCsv/...` が全て通る
- ローカルCSVモードで holidayjp プロバイダーと同等の結果が得られること

---

## Step 4: googleCalendar プロバイダー実装

### 目的
Google Calendar APIを使った `HolidayProvider` 実装を追加する。

### 作業内容

**googleCalendar プロバイダーの実装**
- `providers/googleCalendar/provider.go`: `Options`（APIKey / CredentialsFile）を受け取る `New()`
- APIキー認証と OAuth2 サービスアカウント認証の両対応
- Calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com` から祝日を取得
- `go get google.golang.org/api/calendar/v3`
- `go get golang.org/x/oauth2`

**テスト**
- `providers/googleCalendar/provider_test.go`:
  - APIキーが空の場合にエラーが返ること（インターフェース契約のテスト）
  - 実際のAPI呼び出しは `//go:build integration` タグで分離し、通常の `go test` では実行しない

### 完了条件
- `go build ./providers/googleCalendar/...` が通る
- インテグレーションタグなしの `go test ./...` が全て通る

---

## Step 5: example + ドキュメント整備

### 目的
利用者向けのサンプルコードと各種ドキュメントを整備し、ライブラリとして公開できる状態にする。

### 作業内容

**example の実装**
- `example/main.go`: 以下の全パターンを示すサンプルコード
  - holidayjp プロバイダーで全APIを呼び出す例
  - `WithExcludedDates` + `WithConfig` の併用例
  - caoCsv プロバイダー（ローカルCSV・URL両モード）の例
  - `IsBusinessDay` に `extraExcluded` を渡す例

**GoDoc コメントの追加**
- 全公開型・全公開関数・全公開メソッドに GoDoc コメントを追加
  - `BusinessCalendar` 構造体
  - `New()` / `WithExcludedDates()` / `WithConfig()`
  - `NextBusinessDay()` / `FirstBusinessDayOfMonth()` / `FirstBusinessDaysOfYear()`
  - `IsBusinessDay()` / `Holidays()`
  - `HolidayProvider` インターフェース
  - `MonthDay` / `Holiday` 型

**README・リポジトリルートドキュメントの作成**
- `README.md`（英語）: 概要・インストール・クイックスタート・ライセンス
- `README-ja.md`（日本語）: 同上の日本語版
- `CHANGELOG.md`（英語のみ）: バージョン履歴
- `CONTRIBUTING.md`（英語のみ）: コントリビューション方法
- `LICENSE`: ライセンスファイル

**API仕様ドキュメントの作成**
- `docs/en/api-spec.md`: 全公開型・全API・設定ファイル仕様（英語）
- `docs/ja/api-spec.md`: 同上の日本語版

**使い方ガイドの作成**
- `docs/en/usage.md`: インストール〜各ユースケース別の使い方（英語）
- `docs/ja/usage.md`: 同上の日本語版

**プロバイダーガイドの作成**
- `docs/en/providers.md`: 3プロバイダーの選択基準・設定方法・注意点（英語）
- `docs/ja/providers.md`: 同上の日本語版

**最終確認**
- `go test ./...` が全て通ること
- `go vet ./...` が通ること
- `example/main.go` が実行できること（`go run example/main.go`）
- `go doc` でドキュメントが正しく表示されること

### 完了条件
- `go test ./...` / `go vet ./...` がエラーなし
- `example/main.go` が実行でき、期待通りの出力が得られること
- `README.md` を読んで初めての利用者がライブラリを使い始められること

---

## 依存ライブラリ導入タイミング

| ライブラリ | 導入ステップ |
|-----------|------------|
| `github.com/holiday-jp/holiday_jp-go` | Step 2 |
| `gopkg.in/yaml.v3` | Step 1 |
| `github.com/mikan/syukujitsu-go` | Step 3 |
| `golang.org/x/text`（mikan 経由の推移的依存） | Step 3 |
| `google.golang.org/api/calendar/v3` | Step 4 |
| `golang.org/x/oauth2` | Step 4 |

---

## 各ステップの依存関係

```
Step 1（コア）
  └── Step 2（holidayjp + 全API）
        ├── Step 3（caoCsv）
        ├── Step 4（googleCalendar）
        └── Step 5（example + ドキュメント）
```

Step 3 と Step 4 は Step 2 完了後であれば並行して進められる。  
Step 5 は Step 3・4 の完了を待って実施する。
