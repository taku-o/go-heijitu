# ギャップ分析: Step 5 — example・ドキュメント整備

## 分析サマリー

- Step 1〜4 の実装（コア・3プロバイダー・テスト）は完了済み。本ステップは**新規ファイル作成が中心**（example・各種ドキュメント）で、既存コードへの変更は GoDoc の**パッケージコメント追加**に限定される。
- コアおよび3プロバイダーの公開型・関数・メソッドは**既に GoDoc コメントを保有**している。一方、**全パッケージでパッケージレベルの doc コメント（`// Package ...`）が欠落**しており、これが GoDoc の主な残作業。
- **公開 `Config` 型は存在しない**（実装は非公開の `config` 型 + `loadConfig`）。`docs/planning/api-spec.md` は公開 `Config` を記載しているが実態と乖離。新 API 仕様ドキュメントは**実際の公開シグネチャ**に合わせる必要がある（要件2の `Config` 言及も要修正）。
- `docs/planning/` 配下の既存資料（api-spec.md / design.md / structure.md / usage 相当の情報）が、新ドキュメント（docs/en|ja・README）の**有力な元資料**になる。ただし計画資料であり実態との差分（Config・最新シグネチャ）に注意。
- example を `go run` で動かすには **補助ファイル**（`WithConfig` 用の設定ファイル、caoCsv ローカルモード用の Shift_JIS CSV）が必要。

---

## 1. 現状調査（Current State）

### 既存コードの公開シンボル（ドキュメント対象）

| パッケージ | 公開シンボル | GoDoc コメント | 備考 |
|-----------|------------|---------------|------|
| `heijitu`（ルート） | `BusinessCalendar`（型）/ `New` / `IsBusinessDay` / `NextBusinessDay` / `FirstBusinessDayOfMonth` / `FirstBusinessDaysOfYear` / `Holidays` | あり | `calendar.go` |
| `heijitu` | `Option` / `WithExcludedDates` / `WithConfig` | あり | `option.go` |
| `heijitu` | `HolidayProvider`（インターフェース） | あり | `provider.go` |
| `heijitu` | `MonthDay` / `Matches` | あり | `monthday.go` |
| `heijitu` | `Holiday` | あり | `holiday.go` |
| `heijitu` | （`config` は非公開） | — | `config.go`：`type config`（小文字）+ `loadConfig`。**公開 `Config` は無い** |
| `providers/holidayjp` | `Provider` / `New` / 3メソッド | あり | |
| `providers/caoCsv` | `Options` / `Provider` / `New` / 3メソッド | あり | |
| `providers/googleCalendar` | `Options` / `Provider` / `New` / 3メソッド | あり | |

- **欠落**: 全4パッケージ（`heijitu`・`providers/holidayjp`・`providers/caoCsv`・`providers/googleCalendar`）に**パッケージ doc コメントが無い**。`go doc <pkg>` のパッケージ概要が空になる。

### 既存ドキュメント・成果物の状況

| 対象 | 状況 |
|------|------|
| `README.md` | 存在するがほぼ空（13バイト） |
| `README-ja.md` | 無し |
| `LICENSE` | 無し（product.md は MIT を明記） |
| `CHANGELOG.md` / `CONTRIBUTING.md` | 無し |
| `example/` | 無し |
| `docs/en/` `docs/ja/` | 無し |
| `docs/planning/` | あり（api-spec.md / design.md / structure.md / workplan.md / investigation.md）＝元資料 |

### 規約（steering より）

- 多言語方針: README en/ja・api-spec en/ja・usage en/ja・providers en/ja。CHANGELOG/CONTRIBUTING は英語のみ。
- 1ファイル1責務、`gofmt` 準拠。テストは `対象_test.go`。
- ライセンス: MIT（product.md）。

---

## 2. 要件→資産マップ（ギャップタグ: Missing / Unknown / Constraint）

| 要件 | 必要資産 | 現状 | ギャップ |
|------|---------|------|---------|
| R1 example | `example/main.go` ＋補助ファイル（設定ファイル・テスト用CSV） | 無し | **Missing**（新規）。caoCsv URL / googleCalendar はネットワーク/鍵依存（**Constraint**） |
| R2 GoDoc | 全公開シンボルのコメント | 大部分あり | **Missing**: パッケージ doc コメント（4パッケージ）。`Config` は公開シンボルとして存在せず要件記述の**修正対象**（**Constraint**） |
| R3 README | `README.md` / `README-ja.md` | ほぼ空 / 無し | **Missing**（新規作成） |
| R4 ルートドキュメント | `LICENSE`(MIT) / `CHANGELOG.md` / `CONTRIBUTING.md` | 無し | **Missing**。LICENSE の著作権者・年は**Unknown** |
| R5 API仕様 | `docs/en/api-spec.md` / `docs/ja/api-spec.md` | 無し | **Missing**。元資料 `docs/planning/api-spec.md` は公開 `Config` 等で実態と差異（**Constraint**：実シグネチャに合わせる） |
| R6 使い方 | `docs/en/usage.md` / `docs/ja/usage.md` | 無し | **Missing**（新規） |
| R7 プロバイダーガイド | `docs/en/providers.md` / `docs/ja/providers.md` ＋ APIキー取得・integration テスト手順 | 無し | **Missing**（新規。当初計画に無かった APIキー手順を含む） |
| R8 品質ゲート | `go build`/`go test`/`go vet` 成功・example 実行 | コードは通過する状態 | example 追加後の `go run` 成功（**Constraint**：オフライン安全性） |

---

## 3. 実装アプローチ

ドキュメント整備のため、大半は **Option B（新規作成）**、GoDoc のみ **Option A（既存ファイルへ追記）** の組み合わせ（実質ハイブリッド）。

### GoDoc（Option A: 既存ファイルへ追記）
- 各パッケージに `// Package ...` doc コメントを追加（`doc.go` 新設、または既存ファイル先頭に付与のいずれか＝設計で決定）。
- 既存の型/関数/メソッドコメントは概ね揃っているため、**抜け漏れの確認と統一**が主作業。
- ✅ 振る舞い変更なし・低リスク。❌ コメントのみとはいえ全公開シンボルの再点検が必要。

### example・各種ドキュメント（Option B: 新規作成）
- `example/main.go` ＋ 設定ファイル（YAML/JSON）＋ caoCsv ローカル用 Shift_JIS CSV。
- `docs/en|ja/` 3種（api-spec / usage / providers）＋ README en/ja ＋ ルートドキュメント。
- 元資料 `docs/planning/` を基に、**実コードのシグネチャへ整合**させて作成。
- ✅ 既存に影響しない。❌ ファイル数が多く、en/ja の内容対応の維持コスト。

---

## 4. 複雑度・リスク

- **Effort: L（1〜2週間相当）** — 実装難度は低いが、example＋多言語ドキュメント9点＋ルートドキュメント＋GoDoc点検と量が多い。
- **Risk: Low〜Medium** —
  - Low: コード変更が GoDoc コメント追加に限定され、アーキテクチャ的リスクなし。
  - Medium 要素: (a) ドキュメントと実コードの**整合性維持**（特に planning 資料の `Config` 等の陳腐化を引き継がない）、(b) example の**オフライン安全性**（caoCsv URL はネットワーク依存。googleCalendar は鍵ガード済み方針）、(c) en/ja の内容対応。

---

## 5. 設計フェーズへの申し送り（Research Needed / 決定事項）

1. **GoDoc パッケージコメントの配置**: 各パッケージに `doc.go` を新設するか、既存ファイル先頭に付与するか（4パッケージ分）。
2. **要件2の `Config` 言及の扱い**: 公開 `Config` 型は存在しない。要件/設計では「設定ファイル仕様（`excluded_dates`）」として文書化し、公開型一覧から `Config` を除く（API仕様は実シグネチャに整合）。
3. **example の補助ファイルと実行モデル**:
   - `WithConfig` 用の設定ファイル（YAML/JSON のいずれか/両方）を `example/` 配下に用意。
   - caoCsv ローカルモード用の Shift_JIS テストCSV を用意（既存 `providers/caoCsv/testdata/` の再利用可否を設計で判断）。
   - caoCsv オンライン（URL）モードと googleCalendar（鍵ガード）の**ネットワーク非依存環境での実行挙動**（スキップ/エラーハンドリング）を設計で確定（要件: 鍵未設定で `go run` が正常終了）。
4. **ドキュメントの目次・粒度**: docs/en|ja の各ファイルの章立て。元資料 `docs/planning/api-spec.md` を基に実コードへ整合。
5. **LICENSE / CHANGELOG の具体値**: ライセンス著作権者・年（MIT、`taku-o` / 2026 想定だが確認）、CHANGELOG の初版バージョン表記。
6. **多言語対応の担保方法**: en/ja の内容対応をどう維持・検証するか（章構成を揃える等）。
7. **APIキー取得手順の記載内容**: ユーザー提示の5手順（Cloud Console プロジェクト作成→Calendar API 有効化→APIキー作成→Calendar API のみへ制限→`export GOOGLE_CALENDAR_API_KEY` で integration テスト）をプロバイダーガイド(en/ja)に詳細、README(en/ja)に要点リンク。

---

## 設計フェーズ Synthesis（design.md 作成時）

### 1. Generalization（一般化）
- README/API仕様/使い方/プロバイダーガイドはすべて「**実コードの公開シグネチャを唯一の真実とする**」という共通問題の変種。各ドキュメントは元資料 `docs/planning/` ではなく実コードを参照源とする方針で統一。
- en/ja は同一章構成のミラーとして扱い、内容対応を構造で担保する（個別最適化しない）。

### 2. Build vs. Adopt（採用判断）
- **採用**: ドキュメント本文の素材は既存 `docs/planning/`（api-spec/design/structure/usage 相当情報）を流用。ただし陳腐化箇所（公開 `Config` 型など）は実コードに合わせて補正。
- **採用**: LICENSE は標準 MIT 本文を採用（独自ライセンス文を作らない）。
- **新規作成**: example・docs/en|ja・README 本文・CHANGELOG/CONTRIBUTING はコードに整合する新規作成。

### 3. Simplification（簡素化）
- パッケージ doc コメントはルートのみ独立 `doc.go`、プロバイダーは既存 `provider.go` 先頭に1行付与（新ファイルを増やしすぎない）。
- example は1ファイル `main.go` に集約し、補助は最小限（設定1ファイル＋最小CSV）。仮想的な将来用途のドキュメント・サンプルは作らない（YAGNI）。
- example のネットワーク/認証依存部はガード＋ログ継続で `go run` の常時正常終了を担保（追加の抽象化やフォールバック機構は導入しない）。

### 4. 確定した設計判断
- 公開 `Config` 型は存在しないため、API仕様・GoDoc 対象から除外し、設定は「設定ファイル仕様（`excluded_dates`）」として文書化（要件2修正済みと整合）。
- example はリポジトリルートからの `go run example/main.go` を前提に補助ファイルを相対参照。
- LICENSE 著作権者・年（MIT / taku-o / 2026 想定）と CHANGELOG 初版表記は実装時に最終確定。
