# API Specification

`go-heijitu` is a Go library for calculating Japanese business days. This document describes every public type, function, method, and the configuration file format.

Import path: `github.com/taku-o/go-heijitu`

## Types

### MonthDay

A month/day value that is valid across years. Used to specify excluded dates (company holidays).

```go
type MonthDay struct {
    Month time.Month // month (time.January .. time.December)
    Day   int        // day (1 .. 31)
}
```

#### Method

```go
// Matches reports whether t falls on this month and day (the year is ignored).
func (md MonthDay) Matches(t time.Time) bool
```

### Holiday

A single holiday.

```go
type Holiday struct {
    Date time.Time // the holiday date
    Name string    // the holiday name (e.g. "元日")
}
```

### HolidayProvider

The interface that abstracts the holiday data source. Each provider package implements it.

```go
type HolidayProvider interface {
    IsHoliday(ctx context.Context, t time.Time) (bool, error)
    HolidayName(ctx context.Context, t time.Time) (string, error)
    HolidaysBetween(ctx context.Context, from, to time.Time) ([]Holiday, error)
}
```

| Method | Description |
|--------|-------------|
| `IsHoliday` | Reports whether the given date is a holiday. |
| `HolidayName` | Returns the holiday name for the given date; returns an empty string for a non-holiday. |
| `HolidaysBetween` | Returns the holidays within the range; both `from` and `to` are inclusive, in ascending date order. |

## BusinessCalendar

The core type that calculates business days.

```go
type BusinessCalendar struct {
    // unexported fields
}
```

### Constructor

```go
func New(provider HolidayProvider, opts ...Option) *BusinessCalendar
```

| Argument | Type | Description |
|----------|------|-------------|
| `provider` | `HolidayProvider` | the holiday-judgment implementation (must not be nil; passing nil panics) |
| `opts` | `...Option` | excluded-date options (see below) |

## Option

```go
type Option func(*BusinessCalendar)
```

### WithExcludedDates

Registers excluded dates passed as a parameter.

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

Loads excluded dates from a configuration file. The format is selected by extension: `.yaml` / `.yml` → YAML, `.json` → JSON. Returns an error if the file cannot be read or parsed.

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

`WithExcludedDates` and `WithConfig` can be combined; the excluded dates from both are merged.

## BusinessCalendar methods

A business day is a day that satisfies all of the following:

1. It is not Saturday or Sunday.
2. `HolidayProvider.IsHoliday` returns false.
3. It is not in the excluded dates registered via `WithExcludedDates` / `WithConfig`.
4. It is not in `extraExcluded` (for `IsBusinessDay` only).

### IsBusinessDay

```go
func (bc *BusinessCalendar) IsBusinessDay(ctx context.Context, t time.Time, extraExcluded ...MonthDay) (bool, error)
```

| Argument | Description |
|----------|-------------|
| `t` | the date to judge |
| `extraExcluded` | additional excluded dates for this call only (variadic, optional) |

Returns `true` when `t` is a business day.

### NextBusinessDay

Returns the first business day after `from`. `from` itself is not included.

```go
func (bc *BusinessCalendar) NextBusinessDay(ctx context.Context, from time.Time) (time.Time, error)
```

### FirstBusinessDayOfMonth

Returns the first business day of the given year and month.

```go
func (bc *BusinessCalendar) FirstBusinessDayOfMonth(ctx context.Context, year int, month time.Month) (time.Time, error)
```

### FirstBusinessDaysOfYear

Returns the first business day of each month of the given year. The result always has 12 elements (index 0 is January, 11 is December).

```go
func (bc *BusinessCalendar) FirstBusinessDaysOfYear(ctx context.Context, year int) ([]time.Time, error)
```

### Holidays

Returns the holidays the provider recognizes within the period. Company excluded dates are not included.

```go
func (bc *BusinessCalendar) Holidays(ctx context.Context, from, to time.Time) ([]Holiday, error)
```

| Argument | Description |
|----------|-------------|
| `from` | start of the period (inclusive) |
| `to` | end of the period (inclusive) |

The result is in ascending date order.

## Provider constructors

### holidayjp.New

```go
import "github.com/taku-o/go-heijitu/providers/holidayjp"

func New() *Provider
```

Uses the holiday data embedded in `holiday-jp/holiday_jp-go`. No network access is required.

### caoCsv.New

```go
import "github.com/taku-o/go-heijitu/providers/caoCsv"

type Options struct {
    CSVPath string // path to a local CSV file; if empty, the official data is fetched online
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

When `CSVPath` is set, the local CSV file is read. When it is empty, the official Cabinet Office holiday data is fetched online. CSV retrieval, Shift_JIS decoding, and parsing are delegated to `github.com/mikan/syukujitsu-go`.

### googleCalendar.New

```go
import "github.com/taku-o/go-heijitu/providers/googleCalendar"

type Options struct {
    APIKey          string // API key authentication
    CredentialsFile string // path to an OAuth2 service account JSON file
}

func New(ctx context.Context, opts Options) (*Provider, error)
```

When both `APIKey` and `CredentialsFile` are set, `CredentialsFile` takes precedence. When both are empty, an error is returned (without network access).

## Configuration file format

The configuration file holds excluded dates under `excluded_dates`. Each entry has a numeric `month` (1–12) and `day`. There is no exported configuration type; the file is read internally by `WithConfig`.

### YAML (`.yaml` / `.yml`)

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

### JSON (`.json`)

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
