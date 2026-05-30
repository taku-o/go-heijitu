# 要件定義書: Step 1 — プロジェクト初期化 + コア実装

## Introduction

日本の営業日を計算する Go ライブラリ `go-heijitu` のコア部分を実装する。
本ステップでは、ライブラリの骨格となる型定義・インターフェース・`BusinessCalendar` 構造体を構築し、
`IsBusinessDay()` の判定ロジックまでを動作可能な状態にする。

祝日判定の具体的な実装（各プロバイダー）と `IsBusinessDay` 以外の API は後続ステップで実装する。

## Boundary Context

- **In scope**: `MonthDay` 型・`Holiday` 型・`HolidayProvider` インターフェース・`BusinessCalendar` の構築・`IsBusinessDay()` の判定ロジック・設定ファイル（YAML/JSON）の読み込み・モックプロバイダーを使ったテスト
- **Out of scope**: 実際の祝日プロバイダー実装（holidayjp / caoCsv / googleCalendar）・`IsBusinessDay` 以外の API（`NextBusinessDay` / `FirstBusinessDayOfMonth` / `FirstBusinessDaysOfYear` / `Holidays`）
- **Adjacent expectations**: `HolidayProvider` インターフェースに準拠した実装は後続ステップが提供する。本ステップでテストに使用するモックプロバイダーは本番利用を目的としない。

---

## Requirements

### Requirement 1: MonthDay 型

**Objective:** As a ライブラリ利用者, I want 年をまたいで有効な月日を指定する型, so that 会社独自の休業日を簡潔に表現できる

#### Acceptance Criteria

1. The go-heijitu ライブラリ shall provide a `MonthDay` type with a `Month` field (type `time.Month`) and a `Day` field (type `int`) as public fields.
2. When `Matches(t time.Time)` is called on a `MonthDay` value and the month and day of `t` match, the go-heijitu ライブラリ shall return `true`, regardless of the year.
3. When `Matches(t time.Time)` is called on a `MonthDay` value and either the month or the day of `t` does not match, the go-heijitu ライブラリ shall return `false`.
4. The go-heijitu ライブラリ shall perform no validation on the `Month` and `Day` fields of `MonthDay`; `Matches` shall perform only a direct equality comparison. A `MonthDay` specifying a date that does not exist in non-leap years (e.g., February 29) shall return `false` for all dates on which that month-day combination does not occur.

---

### Requirement 2: Holiday 型

**Objective:** As a ライブラリ利用者, I want 祝日を日付と名称で表現する型, so that 祝日情報を構造的に扱える

#### Acceptance Criteria

1. The go-heijitu ライブラリ shall provide a `Holiday` type with a `Date` field (type `time.Time`) representing the holiday date and a `Name` field (type `string`) representing the holiday name.

---

### Requirement 3: HolidayProvider インターフェース

**Objective:** As a ライブラリ利用者, I want 祝日判定の実装を差し替えられる仕組み, so that プロジェクトの要件に応じてデータソースを選択できる

#### Acceptance Criteria

1. The go-heijitu ライブラリ shall provide a `HolidayProvider` interface with the following three methods: `IsHoliday`, `HolidayName`, and `HolidaysBetween`.
2. When `IsHoliday` is called for a date recognized as a holiday, the go-heijitu ライブラリ shall return `true` and `nil` error.
3. When `IsHoliday` is called for a date not recognized as a holiday, the go-heijitu ライブラリ shall return `false` and `nil` error.
4. When `HolidayName` is called for a date recognized as a holiday, the go-heijitu ライブラリ shall return the holiday name as a non-empty string and `nil` error.
5. When `HolidayName` is called for a date not recognized as a holiday, the go-heijitu ライブラリ shall return an empty string and `nil` error.
6. When `HolidaysBetween` is called with a `from` date and a `to` date, the go-heijitu ライブラリ shall include holidays that fall on the `from` date and the `to` date themselves (both endpoints inclusive).
7. If a `HolidayProvider` method encounters an error, the go-heijitu ライブラリ shall return that error to the caller without suppressing it.

---

### Requirement 4: BusinessCalendar の構築

**Objective:** As a ライブラリ利用者, I want BusinessCalendar を柔軟に構築する手段, so that 会社の休業日ポリシーをカレンダーに反映できる

#### Acceptance Criteria

1. The go-heijitu ライブラリ shall provide a `New(provider HolidayProvider, opts ...Option) *BusinessCalendar` constructor that accepts a `HolidayProvider` and zero or more `Option` values.
2. When `WithExcludedDates(dates []MonthDay)` is passed as an option to `New()`, the go-heijitu ライブラリ shall register those dates as fixed excluded dates applied to all subsequent `IsBusinessDay` calls on that calendar.
3. When `WithConfig(configPath string)` is passed as an option to `New()`, the go-heijitu ライブラリ shall read the specified config file and register the excluded dates defined in it.
4. When both `WithExcludedDates` and `WithConfig` are used together, the go-heijitu ライブラリ shall merge both sets of excluded dates so that all specified dates are treated as excluded.
5. If the file specified in `WithConfig` does not exist or cannot be read, `WithConfig` shall return an error to the caller.

---

### Requirement 5: IsBusinessDay 判定

**Objective:** As a ライブラリ利用者, I want 指定日が営業日かどうかを判定できる, so that 日付に応じたビジネスロジックを実装できる

#### Acceptance Criteria

1. When `IsBusinessDay` is called for a date that falls on a Saturday or Sunday, the go-heijitu ライブラリ shall return `false`.
2. When `IsBusinessDay` is called for a date that the `HolidayProvider` identifies as a holiday, the go-heijitu ライブラリ shall return `false`.
3. When `IsBusinessDay` is called for a date that matches any excluded date registered via `WithExcludedDates` or `WithConfig`, the go-heijitu ライブラリ shall return `false`.
4. When `IsBusinessDay` is called with one or more `extraExcluded MonthDay` arguments and the target date matches any of them, the go-heijitu ライブラリ shall return `false` for that call only, without affecting other calls.
5. When `IsBusinessDay` is called for a weekday that is not a holiday and does not match any excluded date, the go-heijitu ライブラリ shall return `true`.
6. If the `HolidayProvider` returns an error during the `IsBusinessDay` call, the go-heijitu ライブラリ shall propagate that error to the caller without suppressing it.

---

### Requirement 6: 設定ファイルの読み込み

**Objective:** As a ライブラリ利用者, I want 設定ファイルで休業日を管理できる, so that コードを変更せずに休業日ポリシーを更新できる

#### Acceptance Criteria

1. When a config file path ending in `.yaml` or `.yml` is specified, the go-heijitu ライブラリ shall parse it as YAML.
2. When a config file path ending in `.json` is specified, the go-heijitu ライブラリ shall parse it as JSON.
3. The go-heijitu ライブラリ shall read an `excluded_dates` list from the config file, where each entry has a `month` integer (1–12) and a `day` integer (1–31).
4. If the config file path has an extension other than `.yaml`, `.yml`, or `.json`, the go-heijitu ライブラリ shall return an unsupported format error.
5. If the config file content is malformed or cannot be parsed, the go-heijitu ライブラリ shall return a parse error.
