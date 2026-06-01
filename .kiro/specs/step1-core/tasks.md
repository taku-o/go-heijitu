# Implementation Plan

- [x] 1. プロジェクト基盤のセットアップ

- [x] 1.1 Go モジュールと外部依存の初期化
  - `github.com/taku-o/go-heijitu` モジュールを Go 1.23 で `go.mod` を作成する
  - `go get gopkg.in/yaml.v3` で YAML ライブラリを追加し `go.sum` を生成する
  - `go build ./...` がエラーなく通ること

- [x] 2. コアデータ型の実装

- [x] 2.1 (P) MonthDay 型と月日一致判定の実装
  - `Month`（`time.Month`）と `Day`（`int`）の 2 つの公開フィールドを持つ構造体を定義する
  - 月と日が一致したら `true`、どちらか異なれば `false` を返す `Matches(t time.Time) bool` メソッドを実装する（年は無視する）
  - `monthday.go` ファイルが作成され `go build ./...` がエラーなく通ること
  - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - _Boundary: MonthDay_

- [x] 2.2 (P) Holiday 型の実装
  - `Date`（`time.Time`）と `Name`（`string`）の 2 つの公開フィールドを持つ構造体を定義する
  - `holiday.go` ファイルが作成され `go build ./...` がエラーなく通ること
  - _Requirements: 2.1_
  - _Boundary: Holiday_

- [x] 3. HolidayProvider インターフェースと設定読み込みの実装

- [x] 3.1 (P) HolidayProvider インターフェースの定義
  - `IsHoliday(ctx, t) (bool, error)`、`HolidayName(ctx, t) (string, error)`、`HolidaysBetween(ctx, from, to) ([]Holiday, error)` の 3 メソッドを持つインターフェースを定義する
  - エラーは即座に呼び出し元へ伝播することをインターフェース契約に明示する（コメントで記載）
  - `provider.go` ファイルが作成され `go build ./...` がエラーなく通ること
  - _Requirements: 3.1, 3.7_
  - _Boundary: HolidayProvider_
  - _Depends: 2.2_

- [x] 3.2 (P) 設定ファイル読み込み機能の実装
  - `excluded_dates`（`[]MonthDay`）フィールドを持つ非公開 `config` 型を YAML/JSON タグ付きで定義する
  - `loadConfig(path string) (*config, error)` を実装し、拡張子（`.yaml` / `.yml` → YAML、`.json` → JSON）でパース形式を自動判別する
  - 不明な拡張子の場合はサポート外フォーマットエラーを返す
  - ファイル不存在・パース失敗のいずれもエラーを即返す（フォールバックなし）
  - `config.go` ファイルが作成され、有効な YAML/JSON ファイルを正しくパースできること
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_
  - _Boundary: config_
  - _Depends: 2.1_

- [x] 4. BusinessCalendar と Option 関数の実装

- [x] 4.1 Option 型・BusinessCalendar 構造体・コンストラクタの実装
  - `type Option func(*BusinessCalendar)` を `option.go` に定義する
  - `provider HolidayProvider` と `excludedDates []MonthDay` を持つ `BusinessCalendar` 構造体を `calendar.go` に定義する
  - `New(provider HolidayProvider, opts ...Option) *BusinessCalendar` コンストラクタを実装する（渡されたオプションを順に適用する）
  - `WithExcludedDates(dates []MonthDay) Option` を実装する（`excludedDates` に指定した日付群を追記するオプションを返す）
  - `go build ./...` がエラーなく通ること
  - _Requirements: 4.1, 4.2_

- [x] 4.2 WithConfig オプション関数の実装
  - `WithConfig(configPath string) (Option, error)` を実装する
  - コンストラクタ呼び出し前（`New()` の外）でファイルを読み込み、エラーが発生したら即返す
  - 読み込んだ `excluded_dates` を `excludedDates` に追記するオプション関数を返す
  - `WithExcludedDates` と併用した場合、両方の除外日付が `excludedDates` にマージされること
  - `go build ./...` がエラーなく通ること
  - _Requirements: 4.3, 4.4, 4.5_
  - _Depends: 3.2_

- [x] 4.3 IsBusinessDay 判定ロジックの実装
  - `IsBusinessDay(ctx context.Context, t time.Time, extraExcluded ...MonthDay) (bool, error)` を実装する
  - 土曜・日曜 → `false`、`provider.IsHoliday` が祝日と判定 → `false`（エラーは即伝播）、`excludedDates` に一致 → `false`、`extraExcluded` に一致 → `false`（この呼び出し限り）、いずれにも該当しない → `true` の順で判定する
  - excludedDates と extraExcluded の両方に対して除外日付チェックロジックを共通化して重複を排除する
  - `go build ./...` がエラーなく通ること
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 5. テストの実装

- [x] 5.1 (P) MonthDay.Matches() テスト
  - 月と日が一致する場合（異なる年でも）`true` を返すことを確認するテストを書く
  - 月または日が異なる場合に `false` を返すことを確認するテストを書く
  - 2月29日を指定した MonthDay が閏年では `true`、平年では `false` を返すことを確認するテストを書く
  - テーブル駆動テスト形式で複数パターンをカバーする
  - `go test ./...` で全テストがパスすること
  - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - _Boundary: MonthDay_

- [x] 5.2 (P) loadConfig() テスト
  - YAML ファイルから `excluded_dates` を正しく読み込めることを確認するテストを書く
  - JSON ファイルから `excluded_dates` を正しく読み込めることを確認するテストを書く
  - サポート外拡張子でエラーを返すことを確認するテストを書く
  - 不正フォーマットのファイル内容でパースエラーを返すことを確認するテストを書く
  - `go test ./...` で全テストがパスすること
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_
  - _Boundary: config_

- [x] 5.3 IsBusinessDay テスト（モックプロバイダー使用）
  - `calendar_test.go` 内にテスト専用 `mockProvider` をローカル定義する（`HolidayProvider` インターフェースを満たすシンプルな実装）
  - `HolidayName` が非祝日に対して空文字を返すこと、`HolidaysBetween` が両端の日付を含む範囲を返すこと、`HolidaysBetween` で `from > to` の場合に空スライスと nil error を返すことを `mockProvider` の実装で保証する
  - 土曜・日曜に `false` を返すことを確認するテストを書く
  - モックが祝日と認識する日付に `false` を返すことを確認するテストを書く
  - `WithExcludedDates` / `WithConfig` で登録した除外日付に `false` を返すことを確認するテストを書く
  - `extraExcluded` が当該呼び出し限りの除外として機能し、他の呼び出しに影響しないことを確認するテストを書く
  - 平日・非祝日・除外日付なしの場合に `true` を返すことを確認するテストを書く
  - `WithExcludedDates` と `WithConfig` 併用時に両方の除外日付が有効なことを確認するテストを書く
  - モックがエラーを返したとき `IsBusinessDay` がそのエラーを伝播することを確認するテストを書く
  - `go test ./...` で全テストがパスすること
  - _Requirements: 3.2, 3.3, 3.4, 3.5, 3.6, 3.8, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_
  - _Depends: 4.3, 5.1, 5.2_
