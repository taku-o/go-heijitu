# API仕様

`go-heijitu` は日本の営業日を計算する Go ライブラリです。本書では全ての公開型・公開関数・公開メソッドと、設定ファイルの形式を説明します。

インポートパス: `github.com/taku-o/go-heijitu`

## 型

### MonthDay

年をまたいで有効な月日を表す型。除外日付（会社の休業日）の指定に使用します。

```go
type MonthDay struct {
    Month time.Month // 月（time.January 〜 time.December）
    Day   int        // 日（1 〜 31）
}
```

#### メソッド

```go
// Matches は t がこの月日に一致するかを返す（年は無視する）。
func (md MonthDay) Matches(t time.Time) bool
```

### Holiday

祝日1件を表す型。

```go
type Holiday struct {
    Date time.Time // 祝日の日付
    Name string    // 祝日名（例: "元日"）
}
```

### HolidayProvider

祝日データソースを抽象化するインターフェース。各プロバイダーパッケージが実装します。

```go
type HolidayProvider interface {
    IsHoliday(ctx context.Context, t time.Time) (bool, error)
    HolidayName(ctx context.Context, t time.Time) (string, error)
    HolidaysBetween(ctx context.Context, from, to time.Time) ([]Holiday, error)
}
```

| メソッド | 説明 |
|---------|------|
| `IsHoliday` | 指定日が祝日かどうかを返す。 |
| `HolidayName` | 指定日の祝日名を返す。祝日でない場合は空文字を返す。 |
| `HolidaysBetween` | 期間内の祝日を返す。`from` と `to` の両端を含み、日付昇順。 |

## BusinessCalendar

営業日を計算するコア型。

```go
type BusinessCalendar struct {
    // 非公開フィールド
}
```

### コンストラクタ

```go
func New(provider HolidayProvider, opts ...Option) *BusinessCalendar
```

| 引数 | 型 | 説明 |
|------|----|------|
| `provider` | `HolidayProvider` | 祝日判定の実装（nil 不可。nil を渡すとパニックする） |
| `opts` | `...Option` | 除外日付のオプション（後述） |

## Option

```go
type Option func(*BusinessCalendar)
```

### WithExcludedDates

パラメータとして渡した除外日付を登録します。

```go
func WithExcludedDates(dates []MonthDay) Option
```

```go
cal := heijitu.New(provider,
    heijitu.WithExcludedDates([]heijitu.MonthDay{
        {Month: time.August,   Day: 15},
        {Month: time.December, Day: 31},
    }),
)
```

### WithConfig

設定ファイルから除外日付を読み込みます。形式は拡張子で判別します（`.yaml` / `.yml` → YAML、`.json` → JSON）。ファイルの読み込みやパースに失敗した場合はエラーを返します。

```go
func WithConfig(path string) (Option, error)
```

```go
opt, err := heijitu.WithConfig("heijitu.yaml")
if err != nil {
    log.Fatal(err)
}
cal := heijitu.New(provider, opt)
```

`WithExcludedDates` と `WithConfig` は併用可能で、両方の除外日付がマージされます。

## BusinessCalendar のメソッド

営業日は次を全て満たす日です。

1. 土曜・日曜でない。
2. `HolidayProvider.IsHoliday` が false を返す。
3. `WithExcludedDates` / `WithConfig` で登録した除外日付に含まれない。
4. `extraExcluded` に含まれない（`IsBusinessDay` のみ）。

### IsBusinessDay

```go
func (bc *BusinessCalendar) IsBusinessDay(ctx context.Context, t time.Time, extraExcluded ...MonthDay) (bool, error)
```

| 引数 | 説明 |
|------|------|
| `t` | 判定する日付 |
| `extraExcluded` | この呼び出し限りで追加する除外日付（可変長、省略可） |

`t` が営業日のとき `true` を返します。

### NextBusinessDay

`from` の翌日以降で最初の営業日を返します。`from` 当日は含みません。

```go
func (bc *BusinessCalendar) NextBusinessDay(ctx context.Context, from time.Time) (time.Time, error)
```

### FirstBusinessDayOfMonth

指定した年月の最初の営業日を返します。

```go
func (bc *BusinessCalendar) FirstBusinessDayOfMonth(ctx context.Context, year int, month time.Month) (time.Time, error)
```

### FirstBusinessDaysOfYear

指定した年の各月の最初の営業日を返します。結果は必ず12件です（インデックス0が1月、11が12月）。

```go
func (bc *BusinessCalendar) FirstBusinessDaysOfYear(ctx context.Context, year int) ([]time.Time, error)
```

### Holidays

プロバイダーが認識する期間内の祝日を返します。会社の除外日付は含まれません。

```go
func (bc *BusinessCalendar) Holidays(ctx context.Context, from, to time.Time) ([]Holiday, error)
```

| 引数 | 説明 |
|------|------|
| `from` | 取得期間の開始日（この日を含む） |
| `to` | 取得期間の終了日（この日を含む） |

結果は日付昇順です。

## プロバイダーコンストラクタ

### holidayjp.New

```go
import "github.com/taku-o/go-heijitu/providers/holidayjp"

func New() *Provider
```

`holiday-jp/holiday_jp-go` に埋め込まれた祝日データを使用します。ネットワークアクセスは不要です。

### caoCsv.New

```go
import "github.com/taku-o/go-heijitu/providers/caoCsv"

type Options struct {
    CSVPath string // ローカルCSVファイルのパス。空の場合は公式データをオンライン取得する
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

`CSVPath` を指定した場合はそのローカルCSVファイルを読み込みます。空の場合は内閣府公式の祝日データをオンライン取得します。CSVの取得・Shift_JISデコード・パースは `github.com/mikan/syukujitsu-go` に委譲します。

### googleCalendar.New

```go
import "github.com/taku-o/go-heijitu/providers/googleCalendar"

type Options struct {
    APIKey          string // APIキー認証
    CredentialsFile string // OAuth2 サービスアカウントJSONファイルのパス
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

`APIKey` と `CredentialsFile` の両方を指定した場合は `CredentialsFile` を優先します。両方が空の場合はエラーを返します（ネットワークアクセスなし）。

## 設定ファイルの形式

設定ファイルは `excluded_dates` の下に除外日付を保持します。各要素は数値の `month`（1〜12）と `day` を持ちます。公開された設定型はなく、ファイルは `WithConfig` が内部で読み込みます。

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
