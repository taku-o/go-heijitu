# ギャップ分析: Step 4 — googleCalendar（Google Calendar API）プロバイダー実装

## 分析サマリー

- 既存の `providers/` パターン（1プロバイダー＝1パッケージ、`heijitu` エイリアス参照、`var _ heijitu.HolidayProvider` のコンパイル時充足チェック、テーブル駆動＋Given-When-Then、`//go:build integration` での実通信テスト分離）は確立済みで、本プロバイダーも同パターンに自然に乗る。
- `providers/googleCalendar/` パッケージは未作成（Missing）。`google.golang.org/api/calendar/v3` と `golang.org/x/oauth2` は go.mod 未導入（Missing）。
- 認証（APIキー / OAuth2 サービスアカウント）は既存プロバイダーに前例がない**新規パターン**。ただし `calendar.NewService` は構築時にネットワークアクセスしないため、Requirement 1 AC4（両方空 → ネットワークなしでエラー）は opts 検証のみで満たせる。
- 最大の設計判断は「**祝日データを New 時に一括取得して保持するか（caoCsv 流）／ `*calendar.Service` クライアントを保持しメソッド呼び出し毎に問い合わせるか**」。要件（Req 4/5）は観測可能挙動で記述しているため、どちらでも要件を満たせる。Google Calendar は対象期間を引数に取るライブAPIであり、後者（クライアント保持・都度問い合わせ）の方が自然。
- アプローチは既存規約に従い **Option B（新規パッケージ作成）** が妥当。Effort = M、Risk = Medium（認証・通信挙動が公式に明文化されておらず、インテグレーションでのスモークテスト確認が必要）。

---

## 1. 現状調査（Current State）

### 既存資産とパターン

| 資産 | 内容 | 本ステップでの関係 |
|------|------|--------------------|
| `provider.go`（コア） | `HolidayProvider` インターフェース（`IsHoliday` / `HolidayName` / `HolidaysBetween`）。エラーは握りつぶさず伝播する旨をコメントで明記 | 実装対象の契約。変更しない |
| `providers/holidayjp/provider.go` | ステートレスアダプター（`type Provider struct{}`、`New() *Provider`）。外部接続不要 | 認証なしの最小例。直接の参考にはならない |
| `providers/caoCsv/provider.go` | `Options{CSVPath}` を受ける `New(ctx, opts) (*Provider, error)`。ローカル/オンラインの2モード、**New 時に全データを `entries` へ一括ロード**、点照合は外部ライブラリの `Find` に委譲、`HolidaysBetween` のみ自前で範囲フィルタ＋昇順ソート | **最も近い前例**（`New(ctx, opts) (*Provider, error)` シグネチャ・オンライン取得・エラー伝播） |
| `providers/caoCsv/provider_test.go` | `package caoCsv_test`、`var _ heijitu.HolidayProvider = (*caoCsv.Provider)(nil)`、`testdata/` フィクスチャ、Given-When-Then | 通常テストの様式の参考 |
| `providers/caoCsv/provider_integration_test.go` | `//go:build integration`、実オンライン取得 → 既知祝日（元日）判定 | 実API呼び出しテストの様式の参考 |

### 抽出した規約

- **依存方向**: プロバイダー → コア（`heijitu "github.com/taku-o/go-heijitu"`）。コアはプロバイダーを import しない。外部依存はプロバイダーパッケージに閉じ込める。
- **構築**: 外部I/Oを伴うプロバイダーは `New(ctx context.Context, opts Options) (*Provider, error)`。
- **エラー**: 握りつぶさず伝播。`HolidayName` は非祝日のとき `("", nil)`。
- **`HolidaysBetween`**: 両端含む・日付昇順（`slices.SortFunc` + `Date.Compare`）。
- **テスト**: 通常テストはネットワーク非依存、実通信は `//go:build integration` で分離。コンパイル時インターフェース充足チェックを置く。

### 統合面（Integration Surfaces）

- コアの `Holiday{Date time.Time, Name string}` 型へマッピングする必要がある。
- `BusinessCalendar` 側の変更は不要（プロバイダー注入のみ）。go.mod に2依存を追加。

---

## 2. 要件実現性分析（Requirement-to-Asset Map）

| 要件 | 必要な技術要素 | 既存資産 | ギャップ種別 |
|------|----------------|----------|--------------|
| Req 1: 生成・認証方式選択 | `New(ctx, Options{APIKey, CredentialsFile})`、`CredentialsFile` 優先、両方空 → ネットワークなしでエラー | caoCsv の `New` 形を踏襲可。`calendar.NewService` は構築時に通信しない | Missing（パッケージ未作成）。両方空チェックは opts 検証のみで実現可（Constraint なし） |
| Req 2: APIキー認証 | `option.WithAPIKey(apiKey)` → `calendar.NewService`。公開祝日カレンダーは APIキーのみで読める | 前例なし | Missing（新規認証パターン） |
| Req 3: OAuth2 サービスアカウント認証 | `option.WithCredentialsFile(path)` + `CalendarReadonlyScope`（または `WithAuthCredentialsFile` / `google.CredentialsFromJSON`）。ファイル不正は `NewService` でネットワークなしにエラーになりうる | 前例なし | Missing（新規認証パターン）。`WithCredentialsFile` は pkg.go.dev 上で deprecated 注記あり（要設計判断） |
| Req 4: 祝日データ取得 | 固定 Calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com` への `Events.List(...).TimeMin/TimeMax(RFC3339).SingleEvents(true).OrderBy("startTime").Do()`。終日イベントは `event.Start.Date`（yyyy-mm-dd、`DateTime` は空）、祝日名は `event.Summary`。`NextPageToken` でページング | 前例なし | Missing。データ取得＝ライブAPI（ネットワーク必須） |
| Req 5: HolidayProvider 実装 | `IsHoliday`/`HolidayName`/`HolidaysBetween` を上記取得結果から構成。両端含む・昇順 | caoCsv の `HolidaysBetween` ロジックが参考 | Missing（実装）。挙動契約は既存と同一 |

### 複雑性シグナル

- 外部サービス統合（認証＋ライブHTTP API＋ページング＋日付マッピング）。単純CRUDではない。
- ネットワーク・認証情報必須でオフラインモードなし（caoCsv のローカルモードに相当するものがない）。

---

## 3. 実装アプローチ選択肢

### Option A: 既存プロバイダーの拡張
- 既存の holidayjp / caoCsv は責務・データソースが別個であり、googleCalendar を相乗りさせる余地はない。
- ❌ 不適。採用しない。

### Option B: 新規パッケージ作成（推奨）
- `providers/googleCalendar/provider.go` を新設し、`Options{APIKey, CredentialsFile}` ＋ `New(ctx, opts) (*Provider, error)` を実装。既存の1プロバイダー＝1パッケージ規約に合致。
- 統合点: コアの `HolidayProvider` を実装し `Holiday` 型を返すのみ。コア・他プロバイダーへ影響なし。
- ✅ 既存規約と整合、独立テスト容易、関心の分離が明快。
- ❌ go.mod に外部依存2件（＋推移的依存）追加。
- **トレードオフ評価**: 本プロジェクトの確立済みパターンそのもの。最も妥当。

### Option C: ハイブリッド
- 本機能は単一プロバイダーの追加で完結し、コア改変や段階的移行を要しない。過剰。
- ❌ 不要。

---

## 4. データ保持モデル（設計フェーズで決定すべき主要論点）

要件は観測可能挙動で書かれているため以下いずれでも要件充足可能だが、実装方針として設計フェーズで確定が必要。

- **モデル1: クライアント保持・都度問い合わせ（推奨候補）**
  `Provider` が `*calendar.Service` を保持。`IsHoliday`/`HolidayName` は対象日の1日窓を、`HolidaysBetween` は `from`〜`to` 窓を `Events.List` で問い合わせる。Google Calendar が対象期間を引数に取るライブAPIである性質に合致。`Provider` 構造体は `struct{}` ではなくサービスクライアントを持つ（design.md / structure.md の `type Provider struct{}` は雛形であり、保持フィールドが必要になる点を設計で明記）。
- **モデル2: New 時に一括取得（caoCsv 流）**
  取得には対象期間が必要だが、Google Calendar は無限範囲を持つため「いつからいつまでを New 時に取るか」を決められない。固定範囲を切ると将来日・過去日の判定が破綻する。**本データソースには不適**。

→ モデル1を基本線として設計することを推奨。

---

## 5. Effort / Risk

- **Effort: M（3〜7日）** — 新規外部統合（認証2方式・ライブHTTP・ページング・終日イベントの日付/名称マッピング）。既存テスト様式は流用できるが、実装パターン自体は新規。
- **Risk: Medium** — 以下が公式に明文化されておらず、インテグレーションでのスモークテスト確認が前提:
  - APIキーのみで公開祝日カレンダーを読める点（前例ライブラリ `haruotsu/go-jpholiday` 等で実績はあるが公式明記なし）。
  - `calendar.NewService` が構築時に通信しない点（生成系Goクライアントの定着挙動だが公式プロミスではない）。
  - Calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com` の安定性（広く使われるが Google 公式ドキュメント化はされていない）。

---

## 6. 設計フェーズへの申し送り（Research Needed）

1. **データ保持モデルの確定**: モデル1（クライアント保持・都度問い合わせ）で確定するか。`Provider` 構造体が保持するフィールド（`*calendar.Service` 等）を design.md に明記する。
2. **認証構築APIの選定**: サービスアカウントは `option.WithCredentialsFile` + `option.WithScopes(calendar.CalendarReadonlyScope)` か、deprecated 回避の `option.WithAuthCredentialsFile(option.ServiceAccount, path)` / `google.CredentialsFromJSON` か。APIキーは `option.WithAPIKey`。優先順位（CredentialsFile 優先）の分岐点を設計に落とす。
3. **`New` での検証範囲**: 両方空 → ネットワークなしでエラー（Req 1 AC4）。`CredentialsFile` のファイル読み込み/パース失敗を `New` でネットワークなしに検出できる範囲（`NewService` 構築時に出る分）と、実認証成否（`.Do()` が必要＝integration 側）の切り分けを明記。
4. **`Events.List` クエリ詳細**: `TimeMin`/`TimeMax`（RFC3339、`Z` 必須、`TimeMax` は排他上限）、`SingleEvents(true)` + `OrderBy("startTime")`、終日イベントは `event.Start.Date`（`End.Date` は翌日＝排他）・名称は `event.Summary`。`IsHoliday`/`HolidayName` の単日問い合わせ窓の取り方（当日0時〜翌日0時）。
5. **ページング方針**: `NextPageToken` ループ（または `Pages` ヘルパー）で全件取得する（1年窓なら通常1ページだが正しくループする）。
6. **テスト分離**: 通常テストは Req 1 AC4（両方空→エラー）等のネットワーク非依存契約のみ。実API取得（Req 2/3/4/5）は `//go:build integration` で分離（caoCsv の `provider_integration_test.go` 様式に倣う）。実認証情報の与え方（環境変数等）を設計で決める。

---

## 次のステップ

- 本ギャップ分析を踏まえて `/kiro-spec-design step4-googlecalendar` で設計書を作成する（要件を自動承認して進む場合は `-y`）。
- 設計時は上記「申し送り」1〜6、特に**データ保持モデル**と**認証構築API選定**を先に確定する。

_調査出典: pkg.go.dev（`google.golang.org/api/calendar/v3`・`.../option`）、`golang.org/x/oauth2/google`、参考実装 `github.com/haruotsu/go-jpholiday`。項目「APIキーでの公開カレンダー読み取り」「NewService 構築時に非通信」は公式明文ではなく定着挙動のため、インテグレーションでの確認を要する。_
