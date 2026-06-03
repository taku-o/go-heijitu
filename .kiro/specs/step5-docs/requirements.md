# 要件定義書: Step 5 — example・ドキュメント整備

## Introduction

日本の営業日計算ライブラリ `go-heijitu` の開発計画 Step 5「example・ドキュメント整備」の要件を定義する。Step 1〜4 で実装したコア（`BusinessCalendar`・型・`HolidayProvider` インターフェース・設定ファイル読み込み）と3つの祝日プロバイダー（holidayjp / caoCsv / googleCalendar）を、利用者が導入・利用できる状態にするためのサンプルコードと各種ドキュメントを整備する。

具体的には、動作するサンプルコード（`example/`）、全公開シンボルへの GoDoc コメント、多言語の README、リポジトリルートドキュメント（CHANGELOG / CONTRIBUTING / LICENSE）、および多言語のドキュメント群（API仕様・使い方ガイド・プロバイダーガイド）を作成する。プロバイダーガイドには、当初の作業計画には無かった **googleCalendar の APIキー取得手順** と **integration テスト実行手順** を新たに追加する（README にも要点を記載する）。

本ステップはドキュメントとサンプルコードの整備が主目的であり、ライブラリの新機能追加や既存の振る舞い変更は行わない（GoDoc コメントの追加を除く）。

## Boundary Context

- **In scope**:
  - サンプルコード `example/main.go`（および example 実行に必要な設定ファイル・テスト用CSV等の補助ファイル）
  - 既存の全公開型・公開関数・公開メソッドへの GoDoc コメント追加（コメントのみ。振る舞いは変更しない）
  - `README.md`（英語）・`README-ja.md`（日本語）
  - `LICENSE`（MIT）・`CHANGELOG.md`（英語）・`CONTRIBUTING.md`（英語）
  - `docs/en/api-spec.md`・`docs/ja/api-spec.md`（API仕様、英語/日本語）
  - `docs/en/usage.md`・`docs/ja/usage.md`（使い方ガイド、英語/日本語）
  - `docs/en/providers.md`・`docs/ja/providers.md`（プロバイダーガイド、英語/日本語。googleCalendar の APIキー取得手順・integration テスト実行手順を含む）
- **Out of scope**:
  - ライブラリの新しい機能・API・型の追加、既存の振る舞いの変更（GoDoc コメント追加を除く）
  - プロバイダー実装（holidayjp / caoCsv / googleCalendar）のロジック変更
  - `docs/planning/` 配下の既存計画資料の改変
  - pkg.go.dev への公開作業・バージョンタグ付け・リリース手続き・CI/CD 構築
  - 新しい祝日プロバイダーの追加
- **Adjacent expectations**:
  - Step 1〜4 の実装（コア・3プロバイダー・テスト）が完了済みであることを前提とする。
  - googleCalendar の実 API 利用には、利用者自身が用意した有効な APIキー（または認証情報）とネットワーク接続が必要であり、その発行・管理は利用者の責任とする。本ステップはその取得手順を文書化するのみで、認証情報を同梱しない。
  - 内閣府CSV（caoCsv のオンライン取得モード）の利用にはネットワーク接続が必要である。

---

## Requirements

### Requirement 1: サンプルコード（example）

**Objective:** As a ライブラリ利用者, I want 動作するサンプルコードを参照したい, so that 各プロバイダーと公開APIの使い方を実際に動くコードで理解できる

#### Acceptance Criteria

1. The example shall provide an executable program at `example/main.go` that can be run with `go run example/main.go`.
2. The example shall demonstrate holidayjp プロバイダーを用いた全公開API（`IsBusinessDay` / `NextBusinessDay` / `FirstBusinessDayOfMonth` / `FirstBusinessDaysOfYear` / `Holidays`）の呼び出しとその出力.
3. The example shall demonstrate `WithExcludedDates` と `WithConfig` を併用した除外日付の設定.
4. The example shall demonstrate `IsBusinessDay` に `extraExcluded` 引数を渡す呼び出し.
5. The example shall demonstrate caoCsv プロバイダーのローカルCSVモードとオンライン取得（URL）モードの両方の生成方法.
6. While `GOOGLE_CALENDAR_API_KEY` 環境変数が設定されている, the example shall execute googleCalendar プロバイダーを用いた使用例.
7. If `GOOGLE_CALENDAR_API_KEY` 環境変数が未設定である, then the example shall skip the googleCalendar 部分の実行 and 他の例の実行を継続する.
8. When `go run example/main.go` が `GOOGLE_CALENDAR_API_KEY` 未設定の環境で実行される, the example shall 認証情報を要求せずに正常終了する.

---

### Requirement 2: GoDoc コメント

**Objective:** As a ライブラリ利用者, I want すべての公開シンボルに GoDoc コメントが付いている, so that `go doc` や pkg.go.dev で各 API の説明を参照できる

#### Acceptance Criteria

1. The ライブラリ shall provide a GoDoc コメント for すべての公開型（`BusinessCalendar`・`HolidayProvider`・`MonthDay`・`Holiday`・各プロバイダーの `Provider` と `Options`）. なお設定ファイルの読み込み型は非公開（`config`）であり公開型一覧には含めない（設定ファイルの仕様は Requirement 5 の設定ファイル仕様で扱う）.
2. The ライブラリ shall provide a GoDoc コメント for すべての公開関数・公開メソッド（`New`・`WithExcludedDates`・`WithConfig`・`IsBusinessDay`・`NextBusinessDay`・`FirstBusinessDayOfMonth`・`FirstBusinessDaysOfYear`・`Holidays`・各プロバイダーの `New` および `IsHoliday`/`HolidayName`/`HolidaysBetween`）.
3. The GoDoc コメント shall 対象シンボル名で始まる Go の Doc 慣習に従う.
4. When `go doc` が任意の公開シンボルに対して実行される, the ライブラリ shall そのシンボルのコメント本文を表示する.

---

### Requirement 3: README（英語・日本語）

**Objective:** As a 初めての利用者, I want README から概要と最短の使い始め方を把握したい, so that 迷わずにライブラリを導入できる

#### Acceptance Criteria

1. The リポジトリ shall provide `README.md`（英語）with 概要・インストール手順・クイックスタート・ライセンス記載.
2. The リポジトリ shall provide `README-ja.md`（日本語）with `README.md` と同等の内容.
3. The README（英語・日本語の両方） shall include googleCalendar の APIキー取得手順の要点と、プロバイダーガイドの該当箇所へのリンク.
4. The README（英語・日本語の両方） shall include 相互の言語版へのリンク.

---

### Requirement 4: リポジトリルートドキュメント（LICENSE・CHANGELOG・CONTRIBUTING）

**Objective:** As a 利用者・コントリビューター, I want ライセンス・変更履歴・貢献方法が整備されている, so that 利用条件と貢献の前提が明確になる

#### Acceptance Criteria

1. The リポジトリ shall provide a `LICENSE` ファイル with MIT ライセンス本文.
2. The リポジトリ shall provide a `CHANGELOG.md`（英語）with バージョン履歴の記載.
3. The リポジトリ shall provide a `CONTRIBUTING.md`（英語）with コントリビューション方法の記載.

---

### Requirement 5: API仕様ドキュメント（英語・日本語）

**Objective:** As a 利用者, I want 全公開API・型・設定ファイルの仕様書を参照したい, so that 詳細な仕様を確認しながら実装できる

#### Acceptance Criteria

1. The リポジトリ shall provide `docs/en/api-spec.md`（英語）covering 全公開型・全公開API・設定ファイル仕様.
2. The リポジトリ shall provide `docs/ja/api-spec.md`（日本語）with `docs/en/api-spec.md` と同等の内容.
3. The API仕様ドキュメント shall 現在のライブラリの公開シグネチャ（型・関数・メソッド）と一致する内容を記載する.

---

### Requirement 6: 使い方ガイド（英語・日本語）

**Objective:** As a 利用者, I want ユースケース別の使い方ガイドを参照したい, so that 目的に応じた実装方法を理解できる

#### Acceptance Criteria

1. The リポジトリ shall provide `docs/en/usage.md`（英語）covering インストールから各ユースケース別の使い方.
2. The リポジトリ shall provide `docs/ja/usage.md`（日本語）with `docs/en/usage.md` と同等の内容.

---

### Requirement 7: プロバイダーガイド（英語・日本語）と APIキー取得・テスト実行手順

**Objective:** As a 利用者, I want 3プロバイダーの選択基準・設定方法・注意点と、googleCalendar の APIキー取得・テスト実行手順を参照したい, so that 適切なプロバイダーを選び、Google Calendar を正しく設定・検証できる

#### Acceptance Criteria

1. The プロバイダーガイド（`docs/en/providers.md` および `docs/ja/providers.md`） shall describe 3プロバイダー（holidayjp / caoCsv / googleCalendar）の選択基準・設定方法・注意点.
2. The プロバイダーガイド shall describe googleCalendar の APIキー取得手順として、Google Cloud Console でのプロジェクト作成（または既存選択）・Google Calendar API の有効化・APIキーの作成・キーを Calendar API のみに制限する推奨設定を含む.
3. The プロバイダーガイド shall describe integration テストの実行手順として、`export GOOGLE_CALENDAR_API_KEY=<取得した鍵>` の設定と `go test -tags integration ./providers/googleCalendar/...` の実行を含む.
4. The プロバイダーガイド shall be provided in 英語・日本語の両方で同等の内容.

---

### Requirement 8: ドキュメント整備後の整合性と品質ゲート

**Objective:** As a メンテナー, I want ドキュメント・GoDoc・example の整備後もビルド・テスト・静的解析・サンプル実行が成功する, so that 公開できる品質が保たれていることを確認できる

#### Acceptance Criteria

1. When `go test ./...`（integration タグなし）が実行される, the ライブラリ shall すべてのテストをパスする.
2. When `go vet ./...` が実行される, the ライブラリ shall 警告・エラーを出さない.
3. When `go build ./...` が実行される, the ライブラリ shall エラーなくビルドされる.
4. The 多言語ドキュメント（README・api-spec・usage・providers） shall 英語版と日本語版で内容が対応している.

> **補足（受け入れ基準ではない）**: 本ステップは Step 1〜4 の実装完了を前提とする。example の caoCsv オンライン取得モードと googleCalendar 使用例（`GOOGLE_CALENDAR_API_KEY` 設定時）はネットワーク接続を要する。認証情報の発行・管理は利用者の責任であり、本ステップはその取得手順を文書化するのみで認証情報を同梱しない。
