# API仕様

## 型定義

### MonthDay

年をまたいで有効な月日を表す型。除外日付の指定に使用する。

```go
type MonthDay struct {
    Month time.Month // 月（time.January〜time.December）
    Day   int        // 日（1〜31）
}
```

**メソッド**

```go
// Matches は指定の time.Time がこの月日と一致するか返す
func (md MonthDay) Matches(t time.Time) bool
```

---

### Holiday

祝日1件を表す型。

```go
type Holiday struct {
    Date time.Time // 祝日の日付
    Name string    // 祝日名（例: "元日"）
}
```

---

### Config

設定ファイルから読み込む内容。

```go
type Config struct {
    ExcludedDates []MonthDay `yaml:"excluded_dates" json:"excluded_dates"`
}
```

---

## HolidayProvider インターフェース

```go
type HolidayProvider interface {
    IsHoliday(ctx context.Context, t time.Time) (bool, error)
    HolidayName(ctx context.Context, t time.Time) (string, error)
    HolidaysBetween(ctx context.Context, from, to time.Time) ([]Holiday, error)
}
```

| メソッド | 説明 |
|---------|------|
| `IsHoliday` | 指定日が祝日かどうかを返す |
| `HolidayName` | 指定日の祝日名を返す。祝日でなければ空文字を返す |
| `HolidaysBetween` | 指定期間の祝日リストを返す。`from` と `to` 両端を含む |

---

## BusinessCalendar

### 構造体

```go
type BusinessCalendar struct {
    // エクスポートしないフィールド
    provider      HolidayProvider
    extraHolidays []MonthDay
}
```

### コンストラクタ

```go
func New(provider HolidayProvider, opts ...Option) *BusinessCalendar
```

| 引数 | 型 | 説明 |
|------|----|------|
| `provider` | `HolidayProvider` | 祝日判定の実装 |
| `opts` | `...Option` | 除外日付の設定（後述） |

---

### Option

```go
type Option func(*BusinessCalendar)
```

#### WithExcludedDates

パラメータとして除外日付を渡す。

```go
func WithExcludedDates(dates []MonthDay) Option
```

```go
// 使用例
cal := heijitu.New(provider,
    heijitu.WithExcludedDates([]heijitu.MonthDay{
        {Month: time.August,   Day: 15},
        {Month: time.December, Day: 31},
    }),
)
```

#### WithConfig

設定ファイルのパスを指定して除外日付を読み込む。  
拡張子 `.yaml` / `.yml` → YAML、`.json` → JSON として処理する。

```go
func WithConfig(configPath string) (Option, error)
```

```go
// 使用例
opt, err := heijitu.WithConfig("heijitu.yaml")
if err != nil {
    log.Fatal(err)
}
cal := heijitu.New(provider, opt)
```

`WithExcludedDates` と `WithConfig` は併用可能。両方の除外日付がマージされる。

---

### NextBusinessDay

指定日の翌日以降で最初の営業日を返す。指定日当日は含まない。

```go
func (bc *BusinessCalendar) NextBusinessDay(ctx context.Context, from time.Time) (time.Time, error)
```

| 引数 | 説明 |
|------|------|
| `from` | 起点となる日付。この日自体は結果に含まれない |

| 返り値 | 説明 |
|--------|------|
| `time.Time` | 次の営業日。`.Weekday()` で曜日を取得できる |
| `error` | プロバイダーがエラーを返した場合 |

**営業日の判定条件**（全て満たす場合に営業日）
1. 土日でない
2. `HolidayProvider.IsHoliday()` が false を返す
3. `WithExcludedDates` / `WithConfig` で設定した除外日付に含まれない

**例**
```go
next, err := cal.NextBusinessDay(ctx, time.Now())
// → 例: 2026-05-27 Wednesday
fmt.Printf("%s (%s)\n", next.Format("2006-01-02"), next.Weekday())
```

---

### FirstBusinessDayOfMonth

指定年月の最初の営業日を返す。

```go
func (bc *BusinessCalendar) FirstBusinessDayOfMonth(ctx context.Context, year int, month time.Month) (time.Time, error)
```

| 引数 | 説明 |
|------|------|
| `year` | 対象年（例: 2026） |
| `month` | 対象月（例: `time.April`） |

| 返り値 | 説明 |
|--------|------|
| `time.Time` | 指定月の最初の営業日 |
| `error` | プロバイダーがエラーを返した場合 |

**例**
```go
first, err := cal.FirstBusinessDayOfMonth(ctx, 2026, time.April)
// → 2026-04-01 Wednesday
```

---

### FirstBusinessDaysOfYear

指定年の各月の最初の営業日リスト（12件）を返す。

```go
func (bc *BusinessCalendar) FirstBusinessDaysOfYear(ctx context.Context, year int) ([]time.Time, error)
```

| 引数 | 説明 |
|------|------|
| `year` | 対象年（例: 2026） |

| 返り値 | 説明 |
|--------|------|
| `[]time.Time` | インデックス0が1月、11が12月。必ず12件返る |
| `error` | プロバイダーがエラーを返した場合 |

**例**
```go
list, err := cal.FirstBusinessDaysOfYear(ctx, 2026)
for i, d := range list {
    fmt.Printf("%2d月の最初の営業日: %s (%s)\n", i+1, d.Format("2006-01-02"), d.Weekday())
}
```

---

### IsBusinessDay

指定日が営業日かどうかを返す。`extraExcluded` でこの呼び出し限りの追加除外日付を渡せる。

```go
func (bc *BusinessCalendar) IsBusinessDay(ctx context.Context, t time.Time, extraExcluded ...MonthDay) (bool, error)
```

| 引数 | 説明 |
|------|------|
| `t` | 判定する日付 |
| `extraExcluded` | この呼び出し限りで追加する除外日付（可変長、省略可） |

| 返り値 | 説明 |
|--------|------|
| `bool` | 営業日なら `true` |
| `error` | プロバイダーがエラーを返した場合 |

**営業日の判定条件**（全て満たす場合に営業日）
1. 土日でない
2. `HolidayProvider.IsHoliday()` が false を返す
3. `WithExcludedDates` / `WithConfig` で設定した除外日付に含まれない
4. `extraExcluded` に含まれない

**例**
```go
// 通常の判定
ok, err := cal.IsBusinessDay(ctx, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
// → false（元旦）

// この呼び出し限り 5/1 を休業日として扱う
ok, err := cal.IsBusinessDay(ctx, time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
    heijitu.MonthDay{Month: time.May, Day: 1},
)
// → false
```

---

### Holidays

プロバイダーが認識している期間内の祝日リストを返す。  
会社独自の除外日付（`extraHolidays`）は含まれない。

```go
func (bc *BusinessCalendar) Holidays(ctx context.Context, from, to time.Time) ([]Holiday, error)
```

| 引数 | 説明 |
|------|------|
| `from` | 取得期間の開始日（この日を含む） |
| `to` | 取得期間の終了日（この日を含む） |

| 返り値 | 説明 |
|--------|------|
| `[]Holiday` | 期間内の祝日リスト。日付昇順 |
| `error` | プロバイダーがエラーを返した場合 |

**例**
```go
from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
to   := time.Date(2026, 3, 31, 0, 0, 0, 0, time.Local)
holidays, err := cal.Holidays(ctx, from, to)
for _, h := range holidays {
    fmt.Printf("%s: %s\n", h.Date.Format("2006-01-02"), h.Name)
}
// 2026-01-01: 元日
// 2026-01-12: 成人の日
// 2026-02-11: 建国記念の日
// ...
```

---

## プロバイダーコンストラクタ

### holidayjp.New

```go
import "github.com/taku-o/go-heijitu/providers/holidayjp"

func New() *Provider
```

外部接続不要。`holiday-jp/holiday_jp-go` ライブラリに埋め込まれたデータを使用。

---

### caoCsv.New

```go
import "github.com/taku-o/go-heijitu/providers/caoCsv"

type Options struct {
    CSVPath string // ローカルCSVファイルのパス（優先）
    CSVURL  string // CSVのURL（CSVPathが空の場合に使用）
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

`CSVPath` と `CSVURL` が両方指定された場合は `CSVPath` を優先する。  
どちらも空の場合はエラーを返す。

---

### googleCalendar.New

```go
import "github.com/taku-o/go-heijitu/providers/googleCalendar"

type Options struct {
    APIKey          string // APIキー認証
    CredentialsFile string // OAuth2サービスアカウントJSONファイルのパス
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

`APIKey` と `CredentialsFile` が両方指定された場合は `CredentialsFile` を優先する。  
どちらも空の場合はエラーを返す。

---

## 設定ファイル仕様

### YAML形式（`.yaml` / `.yml`）

```yaml
excluded_dates:
  - month: 1
    day: 2
  - month: 1
    day: 3
  - month: 8
    day: 15
  - month: 12
    day: 31
```

### JSON形式（`.json`）

```json
{
  "excluded_dates": [
    {"month": 1,  "day": 2},
    {"month": 1,  "day": 3},
    {"month": 8,  "day": 15},
    {"month": 12, "day": 31}
  ]
}
```

`month` は数値で指定する（1〜12）。
