# ギャップ分析: Step 3 — caoCsv（内閣府CSV）プロバイダー

> ⚠️ **本書は「初期探索 → ユーザー判断による決定」の経緯を含む。セクション1〜6は初期探索時の記述で、一部は後続の決定で更新されている。実装の最終確定方針は末尾「決定事項の反映」「設計フェーズへの確定方針」を参照すること（そちらが優先）。**

## 分析サマリー

- **スコープ**: 既存の `HolidayProvider` インターフェースを満たす新規プロバイダー `caoCsv` を `providers/caoCsv/` に追加する。コア（`calendar.go` 等）・既存プロバイダー（holidayjp）の変更は不要。
- **既存パターン**: holidayjp プロバイダーが「`providers/<name>/provider.go` + `Provider` 構造体 + `New()` + 3メソッド実装 + `var _ heijitu.HolidayProvider` のコンパイル時充足チェック + Given-When-Then テスト」という再利用可能なテンプレートを確立済み。caoCsv はこれを踏襲できる。
- **新規要素**: holidayjp と異なり `New(ctx, Options) (*Provider, error)` は I/O とエラーを伴う初めてのプロバイダー。CSVデータの取得・パース・データソース選択が新規（取得・デコード・パースは mikan に委譲）。
- **最終アプローチ**: Option B（新規パッケージ `providers/caoCsv/`）+ `github.com/mikan/syukujitsu-go` 全面委譲（`LoadAndParse` / `FetchAndParse` / `Find`）。`Options` は `CSVPath` のみ（指定=ローカル、空=内閣府オンライン）。工数 S〜M、リスク Low〜Medium。

---

## 1. 現状調査（Current State）

### ディレクトリ・パッケージ構成
```
go-heijitu/
├── calendar.go / holiday.go / monthday.go / provider.go / option.go / config.go  # コア（標準ライブラリ中心）
├── providers/
│   └── holidayjp/
│       ├── provider.go         # 既存プロバイダー実装（参照テンプレート）
│       └── provider_test.go
└── （caoCsv は未作成）
```

### 確立済みの規約（holidayjp から抽出）
- **パッケージ命名**: `providers/<name>/` 配下に `package <name>`。
- **構造体**: `type Provider struct{...}`、コンストラクタ `New(...)`。
- **インターフェース充足の担保**: テストファイルに `var _ heijitu.HolidayProvider = (*<name>.Provider)(nil)` を置きコンパイル時に保証。
- **親パッケージ参照**: `heijitu "github.com/taku-o/go-heijitu"` を import し、戻り値に `heijitu.Holiday` を使用。
- **`HolidaysBetween` の整列**: `slices.SortFunc` で `Date.Compare` による昇順ソート（holidayjp:38-51 が再利用可能な実装パターン）。
- **エラー方針**: `HolidayName` は非祝日時 `("", nil)`。エラーは握りつぶさず伝播（`provider.go` の interface コメントに明記）。
- **テスト**: 外部テストパッケージ（`package <name>_test`）+ Given-When-Then コメント + `time.DateOnly` フォーマット。

### 既存の型・インターフェース（変更不可）
```go
// provider.go
type HolidayProvider interface {
    IsHoliday(ctx context.Context, t time.Time) (bool, error)
    HolidayName(ctx context.Context, t time.Time) (string, error)
    HolidaysBetween(ctx context.Context, from, to time.Time) ([]Holiday, error)
}
// holiday.go
type Holiday struct { Date time.Time; Name string }
```

### go.mod 現状
```
go 1.23.4
require gopkg.in/yaml.v3 v3.0.1
require github.com/holiday-jp/holiday_jp-go v0.0.0-20220125203534-53124b4cc19c
```
caoCsv 用の依存（`golang.org/x/text`、または `github.com/mikan/syukujitsu-go`）は未導入。

---

## 2. 要件 → 資産マッピング（Requirement-to-Asset Map）

_（下表は最終決定を反映済み）_

| 要件 | 必要な技術要素 | 既存資産 | ギャップ種別 |
|------|--------------|---------|-----------|
| Req1 コンストラクタ/データソース選択 | `New(ctx, Options{CSVPath})`、CSVPath 指定=ローカル/空=オンライン | holidayjp の `New()`（引数なし版） | **Missing**: I/O + error を伴う New は新規。Options 型も新規 |
| Req2 ローカルCSVモード | mikan `LoadAndParse`、読込失敗エラー | なし | **Missing**（mikan 委譲） |
| Req3 オンライン取得モード | mikan `FetchAndParse`（内閣府固定URL）、取得失敗エラー、永続キャッシュなし | なし | **Missing**（mikan 委譲） |
| Req4 デコード/パース | Shift_JIS→UTF-8、CSV列解釈、ヘッダ行除外、失敗エラー | なし | **Missing**: mikan が内部処理（`golang.org/x/text` 推移的依存） |
| Req5 HolidayProvider 実装 | IsHoliday/HolidayName（mikan `Find`）、HolidaysBetween（自前・昇順両端含む） | holidayjp の3メソッド実装・ソートパターン | **Reusable**: ソート/契約パターンを流用。点照合は `Find` 委譲、`entries` 保持は新規 |

---

## 3. 外部依存の調査結果

### github.com/mikan/syukujitsu-go（設計案の採用候補）
公開API:
```go
type Entry struct { Year, Month, Day int; Name string }
func FetchAndParse(ctx context.Context) ([]Entry, error)  // 内閣府URLをハードコードして取得
func LoadAndParse(name string) ([]Entry, error)           // ローカルファイル読込
func Parse(data []byte) ([]Entry, error)                  // バイト列をパース
func Find(entries []Entry, t time.Time) (name string, found bool)
```
- **Shift_JIS デコードを内部で実施**（`golang.org/x/text/encoding/japanese` を内部利用）。
- **要件との適合性**:
  - ローカルモード → `LoadAndParse(CSVPath)` または `os.ReadFile + Parse` で対応可。
  - URLモード → `FetchAndParse` は**URLハードコードのため不可**。`http.Get(CSVURL)` でバイト取得 → `Parse(body)` なら任意URLに対応可。
  - `HolidaysBetween` 相当は mikan に範囲取得APIがなく、`[]Entry` を自前で範囲フィルタ+昇順ソートして `[]heijitu.Holiday` に変換する必要がある（holidayjp のソートパターン流用）。
  - `x/text` は mikan の推移的依存として取り込まれ、直接 import は不要になる可能性が高い。

### golang.org/x/text（自前パース時）
- `encoding/japanese.ShiftJIS.NewDecoder()` + `transform.NewReader` で UTF-8 化 → `encoding/csv` でパース。
- mikan を使わず標準ライブラリ + x/text のみで実装する場合の選択肢。

### 内閣府CSV 仕様
- URL: `https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv`
- エンコード: Shift_JIS、2列（`国民の祝日・休日月日`, `国民の祝日・休日名称`）、**先頭にヘッダ行あり**（要件4-3で除外対象）。日付は `YYYY/M/D` 形式。
- ライセンス CC-BY。

---

## 4. 実装アプローチ

### Option A: コア/既存プロバイダーを拡張
- 不適合。caoCsv は独立した責務・依存・ライフサイクルを持ち、コアや holidayjp に同居させる理由がない。

### Option B: 新規パッケージ `providers/caoCsv/` を作成【推奨】
- **構成**: `providers/caoCsv/provider.go`（`Provider` + `Options` + `New` + 3メソッド）、`providers/caoCsv/provider_test.go`、`providers/caoCsv/testdata/syukujitsu_test.csv`。
- **データ取得**: ローカルは `os.ReadFile(CSVPath)`、URLは `http.Get(CSVURL)` でバイト取得し、共通の `Parse`（mikan）または自前デコード+`encoding/csv` でパース。`New` 時に内部表現（`map[string]string` "2006-01-02"→名称、または `[]Entry`）へ確定させ、以降の I/O を不要にする（Req1-4 / Req3-3）。_（※後続決定で `LoadAndParse` / `FetchAndParse` 全面委譲・内部表現は `[]Entry` 保持へ変更）_
- **トレードオフ**: ✅ holidayjp と同型で一貫性が高くテストも独立。✅ 依存を caoCsv に閉じ込められる。❌ ファイル数増（許容範囲）。

### Option C: ハイブリッド
- 不要。本機能は単一プロバイダーに収まり段階導入の必要がない。

### CSVパーサーの選択（設計フェーズで確定すべき小決定）
_（※後続決定で確定: B-1 mikan/syukujitsu-go を全面採用。取得も `LoadAndParse` / `FetchAndParse` に委譲し、自前 `http.Get` は使わない。）_
- **B-1: mikan/syukujitsu-go 採用**（確定）: 取得・Shift_JIS デコード・パースを全面委譲。点照合は `Find`。
- **B-2: x/text 自前実装**: 依存を `golang.org/x/text` のみに留めパース挙動を完全制御。実装量増のため不採用。

---

## 5. 工数・リスク

- **工数: S〜M（2〜4日）** — 既存プロバイダーのテンプレートが明確で、新規要素は I/O・デコード・データソース分岐に限定される。
- **リスク: Low〜Medium**
  - *Medium*: テスト用 Shift_JIS CSV フィクスチャ（`testdata/syukujitsu_test.csv`）を Shift_JIS エンコードで用意する必要がある（UTF-8 で作ると文字化け検証が成立しない）。生成手段の確認が必要。
  - *Medium*: mikan `FetchAndParse` の URL ハードコード制約（URLモードは `Parse` 経由で回避する設計合意が必要）。
  - *Low*: `time.Time` の Location 取り扱い。Entry（Year/Month/Day）から `time.Time` を組む際の Location 方針を、`IsHoliday` は引数 `t` の Location、`HolidaysBetween` は `from` の Location に揃えるなど統一が必要（holidayjp は `from.Location()` を使用）。

---

## 6. 設計フェーズへの申し送り

> ⚠️ **本セクションは初期探索時の内容で、後続のユーザー判断により更新済み。** 当初は「mikan `Parse([]byte)` 中核 + 自前 `http.Get`（任意CSVURL対応）」を想定していたが、任意CSVURL廃止に伴い `LoadAndParse` / `FetchAndParse` / `Find` への**全面委譲**へ変更された。最新の確定方針は下記「設計フェーズへの確定方針（2026-06-03 ユーザー判断反映）」を参照（そちらが優先）。残る実挙動確認項目（ヘッダ行除外・go 1.23.4 互換・`Find` の突合基準・Shift_JIS フィクスチャ）も同セクション D に集約済み。

---

## 決定事項の反映（2026-06-03）: 任意CSVURL を廃止し mikan に委譲

ユーザー判断により、**オンライン取得の URL は mikan/syukujitsu-go の責任範囲に委ねる**方針へ確定。これに伴い当初分析の「重要な制約（mikan FetchAndParse の URL ハードコード）」は制約ではなくなった。

### 確定した仕様
- `Options` は `CSVPath string` のみ（`CSVURL` フィールドを廃止）。
- `CSVPath` 指定時 → ローカルCSVモード（mikan `LoadAndParse(CSVPath)`）。
- `CSVPath` 空時 → オンライン取得モード（mikan `FetchAndParse(ctx)`、内閣府固定URL）。
- 「CSVPath / CSVURL の両方空ならエラー」「CSVPath 優先」という当初要件は**消滅**（`CSVPath` 空＝オンラインの正常系）。
- Shift_JIS デコード・CSVパース・ヘッダ行処理は mikan が内部で実施するため、自前の `http.Get` / `golang.org/x/text` 直接利用・`encoding/csv` 自前パースは不要。`golang.org/x/text` は mikan 経由の推移的依存になる。

### 更新後の推奨実装（Option B + mikan 全面委譲・Find 採用）
```go
type Options struct { CSVPath string }
type Provider struct { entries []syukujitsu.Entry } // mikan のパース結果をそのまま保持

func New(ctx context.Context, opts Options) (*Provider, error) {
    var entries []syukujitsu.Entry
    var err error
    if opts.CSVPath != "" {
        entries, err = syukujitsu.LoadAndParse(opts.CSVPath) // ローカル
    } else {
        entries, err = syukujitsu.FetchAndParse(ctx)         // 内閣府固定URL
    }
    if err != nil { return nil, err }
    return &Provider{entries: entries}, nil
}

// IsHoliday / HolidayName は mikan の Find に委譲（独自の日付突合を持たない）
func (p *Provider) IsHoliday(_ context.Context, t time.Time) (bool, error) {
    _, found := syukujitsu.Find(p.entries, t)
    return found, nil
}
func (p *Provider) HolidayName(_ context.Context, t time.Time) (string, error) {
    name, found := syukujitsu.Find(p.entries, t)
    if !found { return "", nil }
    return name, nil
}
```
- `HolidaysBetween` は mikan に範囲APIがないため、`p.entries` を範囲フィルタ+昇順ソートして `[]heijitu.Holiday` に変換する（holidayjp のソートパターン流用）。各 Entry の Year/Month/Day から日付を組み、返り値 `Holiday.Date` は `from.Location()` で構築する。

### 設計フェーズへの確定方針（2026-06-03 ユーザー判断反映）

**A. CSVパーサーと内部データ表現（確定: Find 採用）**
- 「ライブラリを使う・独自実装を避ける」方針に従い、点照合（`IsHoliday` / `HolidayName`）は mikan の `Find` に委譲する。自前の `map[string]string` 構築・日付突合は持たない。
- Provider は `LoadAndParse`（ローカル）/ `FetchAndParse`（オンライン）が返す `[]syukujitsu.Entry` をそのまま保持する（design.md の `map[string]string` から内部表現を変更。公開 API には影響しない軽微な変更）。
- `HolidaysBetween` のみ mikan に範囲APIが無いため自前実装が必須（map/Find いずれを選んでも同じ）。`p.entries` を範囲フィルタし `slices.SortFunc` で昇順ソート（holidayjp:38-51 のパターン流用）。

**B. time.Time の Location/正規化（確定: 壁時計暦日基準）**
- 正規化の原則は「壁時計の暦日（Y/M/D）のみで突合、時刻成分・タイムゾーンは無視、ゾーン変換しない」。コアの既存規約（`MonthDay.Matches` は `t.Month()/t.Day()`、`IsBusinessDay` は `t.Weekday()`）と一致させる。
- `IsHoliday` / `HolidayName`: 点照合は mikan `Find` に委譲する（A）。`Find` の突合が壁時計 Y/M/D かは D で実挙動確認する（一致する想定）。
- `HolidaysBetween(from, to)`: 自前実装側でこの原則を明示的に適用する。範囲判定は各 Entry の Y/M/D を `from` 側の暦日と突き合わせて両端含む、返り値 `Holiday.Date` は holidayjp と同様 `from.Location()` で構築する（プロバイダー差し替え時の挙動一貫性のため）。

**C. オンラインモードのテスト（確定: integration タグで分離）**
- ローカルモード: `testdata/syukujitsu_test.csv`（Shift_JIS）で IsHoliday / HolidayName / HolidaysBetween を網羅。存在しない `CSVPath` → `New` がエラー、もユニットテスト。
- オンラインモード（`FetchAndParse`）: ネットワーク依存のため `//go:build integration` タグで分離し、通常の `go test ./...` から除外する（workplan Step 4 の googleCalendar と同じ方針で統一）。

**D. mikan の未確認点（実装着手時に実挙動で確認）**
- (1) `LoadAndParse` / `FetchAndParse` がヘッダ行を内部除外するか、(2) go 1.23.4 での `go get` / `go build` 互換性、(3) `Find` の日付突合が壁時計 Y/M/D 基準か（B との整合）。
- これらはソース推測で決めず、実装の最初に「`go get` → フィクスチャを読み込み、既知の祝日1件・非祝日1件・総件数を assert」する小さな確認で確定する。`Find` が B と異なる突合をする場合のみ、`HolidaysBetween` と挙動を揃える追加処理を検討する。

### planning docs の整合
本決定に合わせ `docs/planning/{api-spec.md, design.md, structure.md, workplan.md}` の `Options` 定義（`CSVURL` 廃止）・依存ライブラリ（Step3 を `mikan/syukujitsu-go` に）・「両方空ならエラー」記述を更新済み。
