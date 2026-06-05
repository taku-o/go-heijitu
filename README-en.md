# go-heijitu

[日本語](./README.md)

`go-heijitu` is a Go library for calculating Japanese business days. It judges whether a day is a business day (not a weekend, holiday, or excluded date) and finds the next business day, the first business day of a month or year, and the holidays in a period. The holiday data source is pluggable through the `HolidayProvider` interface.

## Installation

```bash
go get github.com/taku-o/go-heijitu
```

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

## Providers

The holiday data source is selected by passing a `HolidayProvider` to `heijitu.New`:

- **holidayjp** — embedded holiday data; no network access.
- **caoCsv** — Cabinet Office official CSV; local file or online.
- **googleCalendar** — Google Calendar holiday calendar; API key or service account.

The `googleCalendar` provider requires a Google Calendar API key. In short: create a project in the Google Cloud Console, enable the Google Calendar API, create an API key under **APIs & Services → Credentials**, and (recommended) restrict the key to the Calendar API only. See the [provider guide](./docs/en/providers.md#obtaining-a-google-calendar-api-key) for the full steps and for running the integration tests.

## Documentation

- [API specification](./docs/en/api-spec.md)
- [Usage guide](./docs/en/usage.md)
- [Provider guide](./docs/en/providers.md)

## License

MIT License. See [LICENSE](./LICENSE).
