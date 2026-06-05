# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `BusinessCalendar` core API: `IsBusinessDay`, `NextBusinessDay`,
  `FirstBusinessDayOfMonth`, `FirstBusinessDaysOfYear`, `Holidays`.
- `HolidayProvider` interface and the `MonthDay` / `Holiday` value types.
- Excluded-date options: `WithExcludedDates` (parameter) and `WithConfig`
  (YAML/JSON configuration file).
- Holiday providers: `holidayjp` (embedded data), `caoCsv` (Cabinet Office CSV,
  local file or online), and `googleCalendar` (Google Calendar API, API key or
  service account).
- Example program (`example/main.go`) and documentation (README, API
  specification, usage guide, and provider guide in English and Japanese).
