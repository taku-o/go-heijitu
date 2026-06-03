# Implementation Plan

- [x] 1. Foundation — 依存追加とテストフィクスチャ

- [x] 1.1 mikan/syukujitsu-go の依存追加
  - `go get github.com/mikan/syukujitsu-go` を実行し `go.mod` と `go.sum` を更新する（`golang.org/x/text` は推移的依存として取り込まれる）
  - セキュリティ: mikan が要求する `golang.org/x/text v0.3.6` は既知脆弱性（CVE-2022-32149 / CVE-2021-38561）を含むため、`go get golang.org/x/text@latest` で patched 版（v0.3.8 以降）へ引き上げ、`go mod tidy` を実行する
  - 観察可能な完了条件: `go build ./...` がエラーなく通り、`go.mod` の `golang.org/x/text` が v0.3.8 以降であること（可能なら `govulncheck ./...` で既知脆弱性が報告されないことも確認する）
  - _Requirements: 4.1, 4.2, 4.3_

- [x] 1.2 Shift_JIS テストフィクスチャの作成
  - `providers/caoCsv/testdata/syukujitsu_test.csv` を作成する。内閣府CSVと同形式（先頭ヘッダ行 + `YYYY/M/D, 祝日名` の2列）で、複数の既知の祝日（例: 元日・成人の日・建国記念の日など数件）を収録する
  - ファイルは **Shift_JIS エンコード**で作成する（UTF-8 で作ると mikan のデコードと整合せず文字化け検証が成立しないため）
  - 観察可能な完了条件: ファイルがヘッダ行 + 複数祝日行を含み、かつ **Shift_JIS であることを検証**できること（例: `file providers/caoCsv/testdata/syukujitsu_test.csv` が UTF-8 と表示されない、`nkf -g` / `hexdump` で Shift_JIS バイト列を確認、または UTF-8 へ変換し直して祝日名が一致する等。UTF-8 で誤作成していないことを確認する）。読込検証は 2.2 で行う
  - 補足: フィクスチャの祝日行数を控えておく（2.2・3.2 のヘッダ非混入・件数検証で参照する）
  - _Depends: 1.1_
  - _Requirements: 4.1, 4.2, 4.3_

- [x] 2. caoCsv プロバイダー（読込・点照合）と mikan 実挙動の早期検証

- [x] 2.1 New と点照合（IsHoliday / HolidayName）の最小実装
  - `providers/caoCsv/provider.go` を新規作成する（`package caoCsv`）。`Options`（`CSVPath` のみ）・`Provider`（`entries []syukujitsu.Entry` を保持）・`New(ctx, opts)` を定義する
  - `New`: `CSVPath` が非空なら `syukujitsu.LoadAndParse(opts.CSVPath)`、空なら `syukujitsu.FetchAndParse(ctx)` を呼び、得た `[]Entry` を保持する。読込・取得・デコード・パースのエラーは握りつぶさず `return nil, err` で伝播する。成功後は追加 I/O を行わない
  - `IsHoliday` / `HolidayName`: `syukujitsu.Find(p.entries, t)` に委譲する。`HolidayName` は非祝日（`found == false`）で `("", nil)`、祝日で `(name, nil)` を返す。`error` は常に `nil`
  - 観察可能な完了条件: `go build ./providers/caoCsv/...` がエラーなく通ること
  - _Depends: 1.1_
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.4, 5.1, 5.2, 5.3, 5.4_
  - _Boundary: caoCsv.Provider_

- [x] 2.2 mikan 実挙動の早期検証 + 点照合の契約テスト
  - `providers/caoCsv/provider_test.go` を新規作成する（`package caoCsv_test`）。`New` / `IsHoliday` / `HolidayName` を直接呼ぶテストを書く（この時点では `HolidaysBetween` 未実装のため、インターフェース充足チェックは 3.2 で行う）
  - `New`: フィクスチャを `CSVPath` に指定 → エラーなく読み込め、収録祝日数と一致する件数が得られること（研究D: ヘッダ行が Entry に混入しない）。存在しないパス → エラーが返ること
  - `IsHoliday` / `HolidayName`: 既知の祝日（壁時計 Y/M/D）→ `true` / 期待する祝日名（例: "元日"）と**完全一致し文字化けしない**こと（研究D: `Find` の壁時計突合・Shift_JIS デコードを実挙動で確認）。祝日でない日付 → `false` / 空文字・エラーなし
  - ヘッダ非混入の確認: フィクスチャのヘッダ行（日付でない文字列、例: "国民の祝日・休日月日"）が祝日として誤検出されないこと（フィクスチャに無い任意の日付で `IsHoliday` が `false`、かつ既知祝日のみ `true` になることで間接確認。件数ベースの厳密確認は HolidaysBetween 実装後の 3.2 で行う）
  - 観察可能な完了条件: `go test ./providers/caoCsv/...` で全テストがパスし、研究D（ヘッダ除外・`Find` の壁時計突合・Shift_JIS デコード）が実挙動で確認できること。ここで前提が崩れていれば後続（HolidaysBetween）着手前に検知する
  - _Depends: 1.2, 2.1_
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 4.1, 4.2, 4.3, 5.1, 5.2, 5.3, 5.4_
  - _Boundary: caoCsv.Provider (provider_test.go)_

- [x] 3. HolidaysBetween の実装とテスト

- [x] 3.1 HolidaysBetween の実装
  - `provider.go` に `HolidaysBetween(ctx, from, to)` を追加する。`from.Location()`・0時0分0秒に正規化した暦日で範囲判定する: `fromDate`/`toDate`/`entryDate` を同一 Location の0時で構築し、`!entryDate.Before(fromDate) && !entryDate.After(toDate)` を満たす Entry を `heijitu.Holiday`（`Date` は `from.Location()` で構築）へ変換し、`slices.SortFunc` で日付昇順にソートして返す（両端含む。`from > to` は空スライス）
  - 観察可能な完了条件: `go build ./providers/caoCsv/...` がエラーなく通ること
  - _Depends: 2.2_
  - _Requirements: 5.5_
  - _Boundary: caoCsv.Provider_

- [x] 3.2 HolidaysBetween のテストとインターフェース充足チェック
  - `provider_test.go` に追記する。`var _ heijitu.HolidayProvider = (*caoCsv.Provider)(nil)` を置き、3メソッド揃った Provider のインターフェース充足をコンパイル時に保証する
  - `HolidaysBetween`: 祝日を含む期間で正しい件数が両端含めて返り、日付昇順で並ぶことを確認する
  - ヘッダ非混入の件数確認: フィクスチャの全祝日を含む十分広い期間を指定し、返却件数が **フィクスチャの祝日行数（ヘッダ行を除いた行数）と一致**すること（ヘッダが Entry に混入していないことを件数で確認）
  - 異なる Location での範囲判定: 同一の暦日範囲を `time.UTC` と JST（`time.FixedZone` 等）で指定しても、期待どおり両端含む同じ祝日集合が返ること（壁時計暦日基準の確認）
  - `from > to`（暦日で逆順）で空スライスと `nil` error が返ることを確認する（holidayjp との挙動一貫性）
  - `from == to` で、その日が祝日なら1件・非祝日なら0件が返ることを確認する（両端含むの境界）
  - 観察可能な完了条件: `go test ./providers/caoCsv/...` で全テストがパスすること
  - _Depends: 3.1_
  - _Requirements: 1.1, 4.3, 5.5_
  - _Boundary: caoCsv.Provider (provider_test.go)_

- [x] 4. (P) オンラインモードの integration テスト
  - `providers/caoCsv/provider_integration_test.go` を新規作成する。ファイル先頭に `//go:build integration` タグを付与し、通常の `go test ./...` ではビルド対象外にする
  - `New`: `CSVPath` 空（`Options{}`）で内閣府公式データをオンライン取得 → エラーなくプロバイダーが得られ、既知の祝日（例: 1月1日 元日）を `IsHoliday` が `true` と判定することを確認する
  - 観察可能な完了条件: `go test -tags integration ./providers/caoCsv/...` でテストがパスし、かつ `go test ./providers/caoCsv/...`（タグなし）ではこのファイルがビルドされない（ネットワーク非依存が保たれる）こと
  - _Depends: 2.1_
  - _Requirements: 1.3, 3.1_
  - _Boundary: caoCsv.Provider (provider_integration_test.go)_

- [ ] 5. 全体検証
  - `go build ./...` がエラーなく通ること
  - `go test ./...`（integration タグなし）で新規テストおよび既存テスト（Step 1・Step 2 実装分）が全てパスすること。なお `//go:build integration` のオンラインテスト（タスク4）はネットワーク依存のため本コマンドの対象外であり、その検証はタスク4で別途実施済みとする（本タスクでは再実行しない）
  - `go vet ./...` がエラーなく通ること
  - 観察可能な完了条件: 上記コマンドが全てエラー0件で完了すること
  - _Depends: 3.2, 4_
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5_
