# Provider Guide

A `HolidayProvider` supplies the holiday data that `BusinessCalendar` uses. `go-heijitu` ships three providers. This guide helps you choose one, configure it, and understand its caveats.

## Choosing a provider

| Provider | Data source | Network | Authentication | Works offline |
|----------|-------------|---------|----------------|---------------|
| `holidayjp` | Data embedded in `holiday-jp/holiday_jp-go` | Not required | None | Yes |
| `caoCsv` | Cabinet Office official CSV (local file or online) | Required for the online mode | None | Yes (local CSV mode) |
| `googleCalendar` | Google Calendar Japanese holiday calendar | Required | API key or OAuth2 service account | No |

- Choose **holidayjp** for the simplest, dependency-light, fully offline option.
- Choose **caoCsv** when you want the official Cabinet Office data, either from a bundled CSV (offline) or fetched online.
- Choose **googleCalendar** when you want data backed by Google's managed holiday calendar.

## holidayjp

```go
import "github.com/taku-o/go-heijitu/providers/holidayjp"

provider := holidayjp.New()
cal := heijitu.New(provider)
```

No options, no network access. The holiday data is embedded in the dependency.

## caoCsv

```go
import "github.com/taku-o/go-heijitu/providers/caoCsv"

// Local CSV mode
provider, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: "syukujitsu.csv"})

// Online mode (fetches the official data)
provider, err := caoCsv.New(ctx, caoCsv.Options{})
```

- When `CSVPath` is set, the local CSV is read (works offline).
- When `CSVPath` is empty, the official Cabinet Office data is fetched online; this requires a network connection.
- CSV retrieval, Shift_JIS decoding, and parsing are delegated to `github.com/mikan/syukujitsu-go`.

**Caveat:** the online mode performs an HTTP request, so it depends on network availability.

## googleCalendar

```go
import "github.com/taku-o/go-heijitu/providers/googleCalendar"

// API key authentication
provider, err := googleCalendar.New(ctx, googleCalendar.Options{APIKey: apiKey})

// OAuth2 service account authentication
provider, err := googleCalendar.New(ctx, googleCalendar.Options{CredentialsFile: "service-account.json"})
```

- When both `APIKey` and `CredentialsFile` are set, `CredentialsFile` takes precedence.
- When both are empty, `New` returns an error without accessing the network.
- The provider reads from the fixed calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com`.

**Caveat (call cost):** the provider issues a Calendar API request on each method call. `NextBusinessDay` / `FirstBusinessDayOfMonth` / `FirstBusinessDaysOfYear` call `IsHoliday` day by day internally, so they incur one API round-trip per day searched. Compared with holidayjp / caoCsv (in-memory), latency and API quota usage are higher.

### Obtaining a Google Calendar API key

1. Create (or select) a project in the [Google Cloud Console](https://console.cloud.google.com/).
2. Open **APIs & Services → Library** and enable **Google Calendar API**.
3. Open **APIs & Services → Credentials → Create credentials → API key**.
4. Restrict the created key to the Calendar API only (recommended): in the key settings, set **API restrictions** so that only the Calendar API is allowed.

The issuance, storage, and rotation of credentials are the user's responsibility. Do not commit API keys into source control.

### Running the integration tests

The real Calendar API tests are separated with the `//go:build integration` tag and are excluded from a normal `go test ./...`. To run them, supply the API key through an environment variable:

```bash
export GOOGLE_CALENDAR_API_KEY=<your key>
go test -tags integration ./providers/googleCalendar/...
```

If `GOOGLE_CALENDAR_API_KEY` is not set, the integration tests are skipped, so `go test -tags integration` does not fail in an environment without credentials.
