# Implementation Plan

- [x] 1. Foundation — holiday_jp-go の依存追加
  - `go get github.com/holiday-jp/holiday_jp-go` を実行し `go.mod` と `go.sum` を更新する
  - `go build ./...` がエラーなく通ること
  - _Requirements: 1.7_

- [x] 2 (P). holidayjp プロバイダーの実装とテスト

- [x] 2.1 IsHoliday / HolidayName / HolidaysBetween の実装
  - `providers/holidayjp/` ディレクトリを作成し `provider.go` を新規作成する（`package holidayjp`）
  - `IsHoliday`: `holiday.IsHoliday(t)` の戻り値を `(bool, nil)` として返す
  - `HolidayName`: `holiday.HolidayName(t)` がエラーを返した場合は `("", nil)` に変換して返す（非祝日の意味のみのエラーであり、I/O なしライブラリのため変換は安全）
  - `HolidaysBetween`: `holiday.Between(from, to)` が返す `map[string]string` を `[]heijitu.Holiday` に変換し、日付を `from.Location()` でパースして日付昇順にソートして返す
  - `go build ./providers/holidayjp/...` がエラーなく通ること
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_
  - _Boundary: holidayjp.Provider_

- [x] 2.2 holidayjp プロバイダーのテスト実装
  - `providers/holidayjp/provider_test.go` を新規作成する（`package holidayjp_test`）
  - IsHoliday: 既知の祝日日付（例: 2025-01-01 元日）で `true` を返すことを確認するテストを書く
  - IsHoliday: 祝日でない平日で `false` を返すことを確認するテストを書く
  - HolidayName: 既知の祝日日付で祝日名が返ることを確認するテストを書く
  - HolidayName: 祝日でない日付で空文字が返り、エラーが `nil` であることを確認するテストを書く
  - HolidaysBetween: 祝日を含む期間で正しい件数が返り、両端の日付を含み、日付昇順で並んでいることを確認するテストを書く
  - HolidaysBetween: `from > to` のとき空スライスと `nil` error が返ることを確認するテストを書く
  - `go test ./providers/holidayjp/...` で全テストがパスすること
  - なお、`holiday_jp-go` の埋め込みデータが対応している年度範囲（通常、現在年度 ± 数年）を確認した上でテスト日付を選択すること
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_
  - _Boundary: holidayjp.Provider_

- [x] 3 (P). BusinessCalendar 残りAPI の実装
  - `NextBusinessDay(ctx, from)`: `from.AddDate(0,0,1)` を起点に `bc.IsBusinessDay(ctx, candidate)` を繰り返し、最初に `true` が返った日付を返す（エラーは即座に返す）
  - `FirstBusinessDayOfMonth(ctx, year, month)`: `time.Date(year, month, 1, 0, 0, 0, 0, time.Local)` を起点に、候補日の月が指定月と異なるまで `bc.IsBusinessDay` を繰り返し、最初に `true` が返った日付を返す（エラーは即座に返す）。月境界チェックを必ず実装し、月をまたいで検索を続けないこと（月内に営業日がない場合の動作はスコープ外）
  - `FirstBusinessDaysOfYear(ctx, year)`: 1月から12月の順に `FirstBusinessDayOfMonth` を呼び出し 12 要素のスライスを返す（いずれかの月でエラーが発生したら即座にそのエラーを返す）
  - `Holidays(ctx, from, to)`: `bc.provider.HolidaysBetween(ctx, from, to)` に委譲して返す（会社独自の除外日付は含まない）
  - `go build ./...` がエラーなく通ること
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 5.1, 5.2, 5.3_
  - _Boundary: BusinessCalendar_

- [x] 4. 統合テストと全体検証
  - 以下のテストを API 単位の各テストファイル（NextBusinessDay 系は `calendar_next_business_day_test.go`、FirstBusinessDayOfMonth / FirstBusinessDaysOfYear 系は `calendar_first_business_day_test.go`、Holidays 系は `calendar_holidays_test.go`）に追記する（`heijitu_test` パッケージ、holidayjp プロバイダーとモックプロバイダーを使い分ける）
  - NextBusinessDay: 金曜日（例: 2025-01-10）を渡すと土日および 2025-01-13（成人の日）をスキップして翌営業日（2025-01-14 火曜日）が返ることを確認する（holidayjp.New() 使用）
  - NextBusinessDay: 祝日前日を渡すと祝日の翌営業日が返ることを確認する（holidayjp.New() 使用）
  - NextBusinessDay: プロバイダーがエラーを返したとき、そのエラーが呼び出し元に伝播することを確認する（モックプロバイダー使用）
  - NextBusinessDay: `WithExcludedDates` で登録した日付が候補からスキップされることを確認する
  - FirstBusinessDayOfMonth: 月初が祝日（例: 2026-01-01 元日）の場合に翌営業日が返ることを確認する（holidayjp.New() 使用）
  - FirstBusinessDayOfMonth: プロバイダーがエラーを返したとき、そのエラーが呼び出し元に伝播することを確認する（モックプロバイダー使用）
  - FirstBusinessDaysOfYear: 2026 年を渡すと 12 件のスライスが返ることを確認する（holidayjp.New() 使用）
  - FirstBusinessDaysOfYear: プロバイダーがエラーを返したとき、そのエラーが呼び出し元に伝播することを確認する（モックプロバイダー使用）
  - Holidays: 指定期間（例: 2026-01-01〜2026-03-31）の祝日リストが正しい件数で返ることを確認する（holidayjp.New() 使用）
  - Holidays: プロバイダーがエラーを返したとき、そのエラーが呼び出し元に伝播することを確認する（モックプロバイダー使用）
  - `go test ./...` で新規テストおよび既存テスト（Step 1 実装分）全てがパスすること
  - _Depends: 2.2, 3_
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 5.1, 5.2, 5.3_
  - _Boundary: BusinessCalendar (calendar_next_business_day_test.go / calendar_first_business_day_test.go / calendar_holidays_test.go)_
