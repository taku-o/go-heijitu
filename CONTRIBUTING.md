# Contributing

Thank you for your interest in contributing to `go-heijitu`.

## Prerequisites

- Go 1.25 or later.

## Building

```bash
go build ./...
```

## Running tests

Run the standard test suite (no network access required):

```bash
go test ./...
```

### Integration tests

Tests that call the real Google Calendar API are separated with the
`//go:build integration` build tag and are excluded from a normal `go test ./...`.
To run them, supply a Google Calendar API key through an environment variable:

```bash
export GOOGLE_CALENDAR_API_KEY=<your key>
go test -tags integration ./providers/googleCalendar/...
```

If `GOOGLE_CALENDAR_API_KEY` is not set, the integration tests are skipped.
See the [provider guide](./docs/en/providers.md) for how to obtain an API key.

## Static analysis and formatting

```bash
go vet ./...
gofmt -l .
```

`gofmt -l .` should print nothing; format any reported files with `gofmt -w`.

## Coding conventions

- Follow standard Go conventions and `gofmt` formatting.
- Keep external dependencies inside provider packages; the core package depends
  only on the standard library and YAML.
- Add a GoDoc comment to every exported type, function, and method.
