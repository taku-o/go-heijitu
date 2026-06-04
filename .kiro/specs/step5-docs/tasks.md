# Implementation Plan

- [x] 1. Foundation — example 補助ファイルの準備

- [x] 1.1 example 用の設定ファイルと caoCsv ローカル用CSVを用意
  - `WithConfig` 用の設定ファイルを example 配下に作成する（`excluded_dates` を含む YAML）。後続 2.1 の `WithConfig` 併用例がこのファイルを参照する
  - caoCsv ローカルモード用CSVを、既存 `providers/caoCsv/testdata/syukujitsu_test.csv` の内容を流用して example 配下に作成する（内閣府CSV形式・Shift_JIS を保持し、`syukujitsu.LoadAndParse` のパース成功を保証）。後続 2.1 の caoCsv ローカルモード例がこのファイルを参照する
  - 観察可能な完了条件: 設定ファイルと Shift_JIS CSV が example 配下に存在し、リポジトリルートからの相対パスで参照でき、CSV が既存テストデータと同一フォーマットであること
  - _Requirements: 1.3, 1.5_

- [ ] 2. Core — example・GoDoc・各ドキュメントの作成

- [ ] 2.1 (P) サンプルプログラムの実装
  - holidayjp プロバイダーで、営業日判定・次の営業日・指定月の最初の営業日・指定年の各月初営業日・期間の祝日一覧を呼び出し、結果を標準出力に表示する
  - `WithExcludedDates` と `WithConfig`（1.1 の設定ファイル）を併用した構築例、および `IsBusinessDay` に呼び出し限定の追加除外日付を渡す例を表示する
  - caoCsv をローカルCSV（1.1 のCSV）・オンラインURL の両モードで生成する例を示し、googleCalendar は `GOOGLE_CALENDAR_API_KEY` 設定時のみ実行し未設定時はスキップを表示する
  - 全プロバイダーセクションをエラーガード＋ログ継続とし、ネットワーク/認証失敗でも異常終了しない
  - 観察可能な完了条件: `GOOGLE_CALENDAR_API_KEY` 未設定で `go run example/main.go` が exit 0 で終了し、holidayjp の各API結果・除外日付の効果・caoCsv ローカル結果・googleCalendar のスキップ表示が出力されること
  - _Depends: 1.1_
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8_
  - _Boundary: example program_

- [ ] 2.2 (P) GoDoc パッケージコメントの追加と公開シンボルの点検
  - ルートパッケージ（単一パッケージ `heijitu`。複数ファイルで構成）にパッケージ doc コメント（ライブラリ概要）を追加し、3プロバイダーパッケージ（holidayjp / caoCsv / googleCalendar）に各パッケージコメントを追加する（コメントのみ。宣言・シグネチャ・import は変更しない）
  - 全公開型・公開関数・公開メソッドを点検し、コメントが無いもの／Go Doc 慣習（対象シンボル名で始まる）違反のものを補い、結果として全公開シンボルがコメントを持つ状態にする
  - 観察可能な完了条件: `go doc github.com/taku-o/go-heijitu` と `go doc ./providers/...` でパッケージ概要が表示され、全公開シンボルにコメントが表示されること。`go build ./...` / `go vet ./...` がエラーなし
  - _Requirements: 2.1, 2.2, 2.3, 2.4_
  - _Boundary: GoDoc comments_

- [ ] 2.3 (P) API仕様ドキュメント（英語・日本語）
  - 公開型（`MonthDay` / `Holiday` / `HolidayProvider` / `BusinessCalendar` / 各プロバイダー `Provider`・`Options`）・公開API シグネチャ・設定ファイル仕様（`excluded_dates` の YAML/JSON）を、実コードの公開シグネチャに整合して記載する。公開 `Config` 型は存在しないため記載しない
  - 観察可能な完了条件: `docs/en/api-spec.md` と `docs/ja/api-spec.md` が同一の章構成で存在し、記載シグネチャが実コードと一致すること
  - _Requirements: 5.1, 5.2, 5.3_
  - _Boundary: docs api-spec_

- [ ] 2.4 (P) 使い方ガイド（英語・日本語）
  - インストールから、営業日判定・次の営業日・月初/年間営業日・祝日一覧・除外日付（パラメータ/設定ファイル）・呼び出し限定の追加除外日付・プロバイダー切替の各ユースケース別の使い方を記載する
  - 観察可能な完了条件: `docs/en/usage.md` と `docs/ja/usage.md` が同一の章構成で存在すること
  - _Requirements: 6.1, 6.2_
  - _Boundary: docs usage_

- [ ] 2.5 (P) プロバイダーガイド（英語・日本語）と APIキー取得・テスト実行手順
  - 3プロバイダーの選択基準（データソース・ネットワーク要否・認証・オフライン可否）・設定方法・注意点（googleCalendar の lazy query コスト、caoCsv URL のネットワーク依存）を記載する
  - googleCalendar の APIキー取得手順（Google Cloud Console でのプロジェクト作成→Calendar API 有効化→APIキー作成→Calendar API のみへの制限推奨）と、`export GOOGLE_CALENDAR_API_KEY=<鍵>` → `go test -tags integration ./providers/googleCalendar/...` の実行手順を記載する
  - 観察可能な完了条件: `docs/en/providers.md` と `docs/ja/providers.md` が同一の章構成で存在し、APIキー取得手順と integration テスト実行手順を含むこと
  - _Requirements: 7.1, 7.2, 7.3, 7.4_
  - _Boundary: docs providers_

- [ ] 2.6 (P) README（英語・日本語）
  - 概要・インストール・クイックスタート（holidayjp の最小例）・プロバイダー要約・docs へのリンク・googleCalendar APIキー取得の要点とプロバイダーガイドへのリンク・MIT ライセンス記載・相互言語リンクを記載する
  - 補足: リンク先（docs/providers 等）のパスは設計で確定済みのため 2.3〜2.5 と並行作成可能（リンク先ファイルの内容には依存しない）
  - 観察可能な完了条件: `README.md`（英語）と `README-ja.md`（日本語）が存在し、相互言語リンクと APIキー要点リンクを持つこと
  - _Requirements: 3.1, 3.2, 3.3, 3.4_
  - _Boundary: README_

- [ ] 2.7 (P) リポジトリルートドキュメント（LICENSE・CHANGELOG・CONTRIBUTING）
  - `LICENSE`（MIT 本文・著作権者 `taku-o`・年 `2026`）・`CHANGELOG.md`（バージョン履歴・初版エントリ、英語）・`CONTRIBUTING.md`（ビルド・`go test ./...`・`go test -tags integration ...`・`go vet`・`gofmt` を含む貢献手順、英語）を作成する
  - 観察可能な完了条件: `LICENSE`・`CHANGELOG.md`・`CONTRIBUTING.md` が存在し、CONTRIBUTING にテスト実行手順が含まれること
  - _Requirements: 4.1, 4.2, 4.3_
  - _Boundary: root docs_

- [ ] 3. Validation — 整備後の全体検証

- [ ] 3.1 整備後の品質ゲートと多言語整合の確認
  - `go build ./...` / `go vet ./...` / `go test ./...`（integration タグなし）が全てエラー0件であることを確認する
  - `GOOGLE_CALENDAR_API_KEY` 未設定で `go run example/main.go` が正常終了し、期待した出力が得られることを確認する
  - `go doc` でパッケージ概要・公開シンボルコメントが表示されることを確認する
  - docs の英語版と日本語版で見出し構成（章・節）が一致し対応するコード例のセットが一致していること、api-spec の記載シグネチャが実コードと一致することを確認する
  - 観察可能な完了条件: 上記コマンドが全てエラー0件で、docs の en/ja の見出し構成一致・対応コード例セット一致・README 相互リンクが確認できること
  - _Depends: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_
  - _Requirements: 8.1, 8.2, 8.3, 8.4_

- [ ] 4. Cleanup — 計画資料のアーカイブ

- [ ] 4.1 docs/planning をアーカイブ場所へ移動し元ディレクトリを削除
  - `.kiro/specs/initial-planning/planning/` ディレクトリを作成し、`docs/planning/` 配下の全ファイル（`api-spec.md`・`design.md`・`structure.md`・`investigation.md`・`workplan.md`）を内容を改変せずそこへ移動する
  - 移動後、`docs/planning/` ディレクトリを削除する
  - 全ドキュメント整備・検証（タスク3.1）の後に実施し、docs 作成で `docs/planning/` を元資料として参照し終えてから移動する
  - 観察可能な完了条件: `docs/planning/` が存在せず、`.kiro/specs/initial-planning/planning/` に元の全ファイルが内容そのままで存在すること。`go build ./...` / `go test ./...`（タグなし）がエラー0件のままであること
  - _Depends: 3.1_
  - _Requirements: 9.1, 9.2, 9.3_
  - _Boundary: planning archive_
