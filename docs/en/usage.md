# Usage Guide

This guide walks through installing `go-heijitu` and using it for common tasks.

## Installation

```bash
go get github.com/taku-o/go-heijitu
```

The core package depends only on the standard library and YAML. Each holiday provider lives in its own subpackage so that its external dependencies are isolated.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "time"

    heijitu "github.com/taku-o/go-heijitu"
    "github.com/taku-o/go-heijitu/providers/holidayjp"
)

func main() {
    ctx := context.Background()
    cal := heijitu.New(holidayjp.New())

    ok, _ := cal.IsBusinessDay(ctx, time.Now())
    fmt.Println("today is a business day:", ok)
}
```

## Use cases

### Judge whether a day is a business day

```go
ok, err := cal.IsBusinessDay(ctx, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
// → false, because January 1 is a holiday
```

### Find the next business day

```go
next, err := cal.NextBusinessDay(ctx, time.Now())
fmt.Printf("%s (%s)\n", next.Format("2006-01-02"), next.Weekday())
```

`from` itself is excluded; the result is the first business day after it.

### First business day of a month

```go
first, err := cal.FirstBusinessDayOfMonth(ctx, 2026, time.April)
```

### First business day of every month in a year

```go
days, err := cal.FirstBusinessDaysOfYear(ctx, 2026)
for i, d := range days {
    fmt.Printf("%2d: %s\n", i+1, d.Format("2006-01-02"))
}
```

The slice always has 12 elements (index 0 is January).

### List holidays in a period

```go
from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.Local)
holidays, err := cal.Holidays(ctx, from, to)
for _, h := range holidays {
    fmt.Printf("%s: %s\n", h.Date.Format("2006-01-02"), h.Name)
}
```

### Exclude company holidays via parameter

```go
cal := heijitu.New(holidayjp.New(),
    heijitu.WithExcludedDates([]heijitu.MonthDay{
        {Month: time.December, Day: 31},
        {Month: time.January, Day: 2},
    }),
)
```

### Exclude company holidays via a configuration file

Create `heijitu.yaml`:

```yaml
excluded_dates:
  - month: 12
    day: 31
  - month: 1
    day: 2
```

```go
opt, err := heijitu.WithConfig("heijitu.yaml")
if err != nil {
    log.Fatal(err)
}
cal := heijitu.New(holidayjp.New(), opt)
```

`WithExcludedDates` and `WithConfig` can be combined; their excluded dates are merged.

### Exclude a date for a single call only

```go
ok, err := cal.IsBusinessDay(ctx,
    time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
    heijitu.MonthDay{Month: time.May, Day: 1},
)
// → false, because 5/1 is treated as excluded for this call only
```

### Switch the holiday data source

Any `HolidayProvider` can be passed to `heijitu.New`. To use the Cabinet Office CSV instead of the embedded data:

```go
provider, err := caoCsv.New(ctx, caoCsv.Options{}) // fetches official data online
if err != nil {
    log.Fatal(err)
}
cal := heijitu.New(provider)
```

See the [provider guide](./providers.md) for how to choose and configure each provider.
