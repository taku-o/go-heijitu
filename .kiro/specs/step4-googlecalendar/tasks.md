# Implementation Plan

- [x] 1. Foundation — Google Calendar API 依存の追加

- [x] 1.1 google.golang.org/api 依存の追加
  - `go get google.golang.org/api/calendar/v3` を実行し `go.mod` / `go.sum` を更新する（`google.golang.org/api/option` も同梱され、`golang.org/x/oauth2` は推移的依存として `// indirect` で取り込まれる）。取得後 `go mod tidy` を実行する
  - 補足: 本依存は要件4・要件5（Calendar API からの祝日取得と HolidayProvider 実装）の前提となるが、本タスクの成果物は依存の追加とビルド成立のみであり、振る舞い検証は後続タスクで行う
  - 観察可能な完了条件: `go build ./...` がエラーなく通り、`go.mod` の `require` に `google.golang.org/api` が追加されていること

- [ ] 2. googleCalendar プロバイダーの実装

- [ ] 2.1 New と認証方式選択の実装
  - `providers/googleCalendar/provider.go` を新規作成する（`package googleCalendar`）。固定 Calendar ID 定数（`ja.japanese.official#holiday@group.v.calendar.google.com`）・`Options`（`APIKey` / `CredentialsFile`）・`Provider`（`service *calendar.Service` を保持）・`New(ctx, opts)` を定義する
  - `New` の認証分岐: `CredentialsFile` が非空なら `option.WithAuthCredentialsFile(option.ServiceAccount, opts.CredentialsFile)` と `option.WithScopes(calendar.CalendarReadonlyScope)` を付与（`APIKey` 併用時も優先。deprecated な `WithCredentialsFile` は使わない）。空かつ `APIKey` 非空なら `option.WithAPIKey`（スコープ指定なし）。両方空なら `calendar.NewService` を呼ばずにエラーを返す（ネットワークアクセスなし）。`NewService` のエラーは握りつぶさず伝播し、成功時に `Provider{service}` を返す
  - 観察可能な完了条件: `go build ./providers/googleCalendar/...` がエラーなく通ること
  - _Depends: 1.1_
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 3.1, 3.2_
  - _Boundary: googleCalendar.Provider_

- [ ] 2.2 祝日取得ヘルパーと単日判定メソッドの実装
  - `provider.go` に private ヘルパー `holidaysInWindow(ctx, timeMin, timeMax, loc)` を実装する。固定 Calendar ID に対し `Events.List` を `TimeMin`/`TimeMax`（UTC・RFC3339）・`SingleEvents(true)`・`OrderBy("startTime")` で発行し、`NextPageToken` で全ページ取得する。終日イベント（`Start.Date` 非空）のみを `heijitu.Holiday{Date, Name: Summary}` に変換する。`Do()` のエラーは握りつぶさず伝播する
  - `IsHoliday` / `HolidayName`: 対象日を含む UTC窓（-1日〜+2日）で `holidaysInWindow` を取得し、`Start.Date` が対象日の壁時計 Y/M/D（`t.Format(time.DateOnly)`）に一致するイベントの有無 / `Summary` を返す。`HolidayName` は非祝日で `("", nil)`、`error` は取得失敗時のみ伝播する
  - 観察可能な完了条件: `go build ./providers/googleCalendar/...` がエラーなく通ること（振る舞い検証はネットワーク依存のためタスク4で行う）
  - _Depends: 2.1_
  - _Requirements: 2.2, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4_
  - _Boundary: googleCalendar.Provider_

- [ ] 2.3 HolidaysBetween の実装
  - `provider.go` に `HolidaysBetween(ctx, from, to)` を追加する。`from.Location()`・0時で `fromDate`/`toDate` を正規化し、`fromDate.After(toDate)`（暦日で逆順）なら `Events.List` を呼ばず空スライス（`[]heijitu.Holiday{}`）と `nil` を返す（`TimeMin > TimeMax` の API エラー回避、holidayjp / caoCsv との挙動一貫性）。そうでなければ `from-1日`〜`to+2日` の UTC窓で `holidaysInWindow` を取得し、`Start.Date` の暦日が範囲内（両端含む）のものを抽出して `Date.Compare` 昇順にソートして返す
  - 観察可能な完了条件: `go build ./providers/googleCalendar/...` がエラーなく通ること（振る舞い検証はタスク4で行う）
  - _Depends: 2.2_
  - _Requirements: 4.4, 5.5_
  - _Boundary: googleCalendar.Provider_

- [ ] 3. ネットワーク非依存の契約テスト
  - `providers/googleCalendar/provider_test.go` を新規作成する（`package googleCalendar_test`）
  - `New`: `Options{}`（`APIKey`・`CredentialsFile` 両方空）→ エラーが返り、ネットワークアクセスが発生しないこと（要件 1.4）。存在しないパスを `CredentialsFile` に指定 → `calendar.NewService` のファイル読込失敗としてエラーが返ること（ネットワークなし、要件 3.2）
  - `var _ heijitu.HolidayProvider = (*googleCalendar.Provider)(nil)` を置き、3メソッド揃った `Provider` のインターフェース充足をコンパイル時に保証する
  - 補足: `HolidaysBetween` の `from > to` ガードはネットワーク非依存ロジックだが、外部テストパッケージでは `New` の成功（＝有効な認証情報）なしに `Provider` 実体を構築できないため、その挙動検証はタスク4（integration）で行う（`New` の振る舞いに依存しない単体検証はここでは行わない）
  - 観察可能な完了条件: `go test ./providers/googleCalendar/...`（タグなし）で全テストがパスすること
  - _Depends: 2.3_
  - _Requirements: 1.1, 1.4, 3.2_
  - _Boundary: googleCalendar.Provider (provider_test.go)_

- [ ] 4. (P) 実API取得の integration テスト
  - `providers/googleCalendar/provider_integration_test.go` を新規作成する。ファイル先頭に `//go:build integration` タグを付与し、通常の `go test ./...` ではビルド対象外にする。APIキーは環境変数 `GOOGLE_CALENDAR_API_KEY` から取得し、未設定なら `t.Skip` でスキップする（認証情報の無い環境でも `go test -tags integration` がエラーにならないようにする）
  - `New`（APIキー認証）→ エラーなくプロバイダーが得られる（要件 1.1, 1.3, 2.1）
  - `IsHoliday`: 既知の祝日（例: 1月1日 元日）→ `true`、平日 → `false`（要件 5.1, 5.2, 4.1, 4.2）。`HolidayName`: 既知の祝日 → 期待する祝日名（文字化けなし）、非祝日 → 空文字（要件 5.3, 5.4）
  - `HolidaysBetween`: 祝日を含む期間 → 件数が両端含めて正しく、日付昇順で並ぶ。境界として `from == to`（同一暦日・その日が祝日なら1件/非祝日なら0件）、`from > to` → 空スライスと `nil`（ガードが API 呼び出し前に短絡することの確認）、および境界日付の祝日が UTC クエリ窓で取りこぼされないこと（要件 5.5, 4.4）
  - 観察可能な完了条件: `go test -tags integration ./providers/googleCalendar/...` がパスし、かつ `go test ./providers/googleCalendar/...`（タグなし）ではこのファイルがビルドされない（ネットワーク非依存が保たれる）こと
  - _Depends: 2.3_
  - _Requirements: 1.1, 1.3, 2.1, 2.2, 4.1, 4.2, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5_
  - _Boundary: googleCalendar.Provider (provider_integration_test.go)_

- [ ] 5. 全体検証
  - `go build ./...` がエラーなく通ること
  - `go test ./...`（integration タグなし）で新規テストおよび既存テスト（Step 1〜3 実装分）が全てパスすること。なお `//go:build integration` のオンラインテスト（タスク4）はネットワーク依存のため本コマンドの対象外であり、その検証はタスク4で別途実施済みとする（本タスクでは再実行しない）
  - `go vet ./...` がエラーなく通ること
  - 観察可能な完了条件: 上記コマンドが全てエラー0件で完了すること
  - _Depends: 3, 4_
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 3.1, 3.2, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5_
