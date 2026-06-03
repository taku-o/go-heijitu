# 設計案: go-heijitu（日本営業日計算ライブラリ）

## ライブラリ概要

日本の営業日を計算するGoライブラリ。祝日判定の実装を差し替え可能なプロバイダー設計とし、会社独自の休業日を外部設定またはパラメータで指定できる。

---

## パッケージ構成

```
go-heijitu/
├── calendar.go          # BusinessCalendar 本体・公開API
├── holiday.go           # Holiday 型
├── provider.go          # HolidayProvider インターフェース
├── config.go            # 設定ファイル読み込み（YAML/JSON）
├── monthday.go          # MonthDay 型
├── providers/
│   ├── holidayjp/
│   │   └── provider.go  # holiday-jp/holiday_jp-go ベース実装
│   ├── caoCsv/
│   │   └── provider.go  # 内閣府CSV ベース実装
│   └── googleCalendar/
│       └── provider.go  # Google Calendar API ベース実装
└── example/
    └── main.go          # 利用例
```

---

## 型定義

### MonthDay

年をまたいで有効な月日を表す。会社独自の休業日指定に使用。

```go
type MonthDay struct {
    Month time.Month
    Day   int
}
```

### Holiday

祝日1件を表す。

```go
type Holiday struct {
    Date time.Time
    Name string
}
```

### Config

設定ファイルから読み込む内容。

```go
type Config struct {
    ExcludedDates []MonthDay `yaml:"excluded_dates" json:"excluded_dates"`
}
```

---

## HolidayProvider インターフェース

祝日判定の実装を切り替えられるようにするためのインターフェース。

```go
type HolidayProvider interface {
    // IsHoliday は指定日が祝日かどうかを返す
    IsHoliday(ctx context.Context, t time.Time) (bool, error)

    // HolidayName は指定日の祝日名を返す。祝日でなければ空文字を返す
    HolidayName(ctx context.Context, t time.Time) (string, error)

    // HolidaysBetween は指定期間の祝日リストを返す（fromとto両端を含む）
    HolidaysBetween(ctx context.Context, from, to time.Time) ([]Holiday, error)
}
```

---

## プロバイダー実装

### HolidayJPProvider（デフォルト）

`github.com/holiday-jp/holiday_jp-go` を内部で使用。外部接続不要。

```go
// providers/holidayjp/provider.go
type Provider struct{}

func New() *Provider
```

### CAOCSVProvider

内閣府CSVから祝日データを取得・パース。ローカルファイルとオンライン取得の2モード対応。

```go
// providers/caoCsv/provider.go
type Provider struct {
    entries []syukujitsu.Entry // mikan のパース結果を保持
}

type Options struct {
    CSVPath string // ローカルCSVファイルパス。空の場合は内閣府公式データをオンライン取得する
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

CSVの取得・パースには `github.com/mikan/syukujitsu-go` を使用する（`CSVPath` 指定時は `LoadAndParse`、空の場合は `FetchAndParse`）。  
Shift_JIS デコードは同ライブラリが内部で処理するため、`golang.org/x/text` は推移的依存となる。  
`IsHoliday` / `HolidayName` の点照合は mikan の `Find` に委譲する。`HolidaysBetween` は mikan に範囲APIが無いため、保持した `entries` を自前で範囲フィルタ＋昇順ソートして返す。

### GoogleCalendarProvider

Google Calendar APIから日本の祝日カレンダーを取得。

```go
// providers/googleCalendar/provider.go
type Provider struct{}

type Options struct {
    APIKey          string // APIキー認証
    CredentialsFile string // OAuth2 サービスアカウント認証
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

---

## BusinessCalendar

営業日計算のメイン構造体。

```go
type BusinessCalendar struct {
    provider      HolidayProvider
    extraHolidays []MonthDay
}
```

### コンストラクタ

```go
func New(provider HolidayProvider, opts ...Option) *BusinessCalendar

type Option func(*BusinessCalendar)

// WithExcludedDates はパラメータで除外日付を指定する
func WithExcludedDates(dates []MonthDay) Option

// WithConfig は設定ファイルのパスを指定して除外日付を読み込む
func WithConfig(configPath string) (Option, error)
```

### 提供API

#### NextBusinessDay

指定日の翌日以降で最初の営業日を返す。

```go
func (bc *BusinessCalendar) NextBusinessDay(ctx context.Context, from time.Time) (time.Time, error)
```

- 返り値: 営業日の `time.Time`（曜日は `.Weekday()` で取得可能）
- `from` 当日は含まない（翌日以降を探索）

#### FirstBusinessDayOfMonth

指定年月の最初の営業日を返す。

```go
func (bc *BusinessCalendar) FirstBusinessDayOfMonth(ctx context.Context, year int, month time.Month) (time.Time, error)
```

#### FirstBusinessDaysOfYear

指定年の各月の最初の営業日リストを返す（12件）。

```go
func (bc *BusinessCalendar) FirstBusinessDaysOfYear(ctx context.Context, year int) ([]time.Time, error)
```

#### IsBusinessDay

指定日が営業日かどうかを返す。`extraExcluded` でこの呼び出し限りの追加除外日付を渡せる。

```go
func (bc *BusinessCalendar) IsBusinessDay(ctx context.Context, t time.Time, extraExcluded ...MonthDay) (bool, error)
```

営業日の判定条件（全て満たす場合に営業日）:
1. 平日（土日でない）
2. 祝日でない（プロバイダーで判定）
3. `extraHolidays`（固定設定）に含まれない
4. `extraExcluded`（呼び出し時パラメータ）に含まれない

#### Holidays

プロバイダーが認識している期間内の祝日リストを返す。

```go
func (bc *BusinessCalendar) Holidays(ctx context.Context, from, to time.Time) ([]Holiday, error)
```

---

## 設定ファイル仕様

### YAML形式（推奨）

```yaml
# heijitu.yaml
excluded_dates:
  - month: 8
    day: 15   # 夏季休業
  - month: 12
    day: 31   # 年末
  - month: 1
    day: 2    # 年始
  - month: 1
    day: 3    # 年始
```

### JSON形式

```json
{
  "excluded_dates": [
    {"month": 8,  "day": 15},
    {"month": 12, "day": 31},
    {"month": 1,  "day": 2},
    {"month": 1,  "day": 3}
  ]
}
```

ファイル拡張子（`.yaml` / `.yml` / `.json`）から形式を自動判別する。

---

## 使用例

### デフォルトプロバイダー（holiday_jp-go）

```go
import (
    heijitu "github.com/taku-o/go-heijitu"
    "github.com/taku-o/go-heijitu/providers/holidayjp"
)

provider := holidayjp.New()
cal := heijitu.New(provider,
    heijitu.WithExcludedDates([]heijitu.MonthDay{
        {Month: time.August,   Day: 15},
        {Month: time.December, Day: 31},
    }),
)

// 次の営業日
next, err := cal.NextBusinessDay(ctx, time.Now())
fmt.Printf("%s (%s)\n", next.Format("2006-01-02"), next.Weekday())

// 指定月の最初の営業日
first, err := cal.FirstBusinessDayOfMonth(ctx, 2026, time.April)

// 指定年の各月の最初の営業日
list, err := cal.FirstBusinessDaysOfYear(ctx, 2026)
```

### 設定ファイル + パラメータ併用

```go
configOpt, err := heijitu.WithConfig("heijitu.yaml")
if err != nil {
    log.Fatal(err)
}

provider := holidayjp.New()
cal := heijitu.New(provider, configOpt)

// この呼び出し限りの追加除外日付
ok, err := cal.IsBusinessDay(ctx, time.Now(),
    heijitu.MonthDay{Month: time.May, Day: 1},  // 会社設立記念日
)
```

### 内閣府CSVプロバイダー（公式データ）

```go
import "github.com/taku-o/go-heijitu/providers/caoCsv"

// ローカルCSVから
provider, err := caoCsv.New(ctx, caoCsv.Options{
    CSVPath: "data/syukujitsu.csv",
})

// オンライン取得（CSVPath を空にすると内閣府公式データを取得）
provider, err := caoCsv.New(ctx, caoCsv.Options{})
```

---

## 外部依存ライブラリ（候補）

| ライブラリ | 用途 | 必須/任意 |
|-----------|------|---------|
| `github.com/holiday-jp/holiday_jp-go` | HolidayJPProvider実装 | holidayjpプロバイダー使用時 |
| `github.com/mikan/syukujitsu-go` | 内閣府CSVパーサー | caoCsvプロバイダー使用時 |
| `golang.org/x/text` | Shift_JISデコード（mikan/syukujitsu-go 経由の推移的依存） | caoCsvプロバイダー使用時 |
| `google.golang.org/api/calendar/v3` | Google Calendar API | googleCalendarプロバイダー使用時 |
| `gopkg.in/yaml.v3` | YAML設定ファイル読み込み | 設定ファイル使用時 |

コア部分（`calendar.go`, `provider.go`）は標準ライブラリのみで実装し、プロバイダーごとに依存を分離する。

---

## 決定事項

| 項目 | 決定内容 |
|------|---------|
| モジュール名 | `github.com/taku-o/go-heijitu` |
| Go バージョン | 1.23 |
| ライセンス | MIT |
| CAOCSVプロバイダーのキャッシュ | キャッシュなし。オンライン取得時は New 呼び出し時に毎回 fetch する |
| CAOCSVプロバイダーのデータソース | `Options.CSVPath` 指定時はローカル、空時は内閣府公式データをオンライン取得。任意URLの指定は受け付けない |
| エラーハンドリング | プロバイダーのエラーは呼び出し元に伝播する。内部で握りつぶさない |
| 振替休日 | 各プロバイダーの実装に委ねる（ライブラリ側では関与しない） |
| 設定ファイルフォーマット | YAML優先、JSON も対応。拡張子で自動判別 |
| 内閣府CSVパーサー | `github.com/mikan/syukujitsu-go` を使用する |
