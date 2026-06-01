# 要件定義書: Step 2 — holidayjp プロバイダー + 残りAPI実装

## Introduction

日本の営業日計算ライブラリ `go-heijitu` において、デフォルト祝日プロバイダー（holidayjp）を実装し、Step 1 で骨格のみ構築した `BusinessCalendar` の残りすべての公開 API を動作可能な状態にする。

本ステップ完了後、`holidayjp` プロバイダーを使ってすべての `BusinessCalendar` API が期待通りに動作する状態になる。

Step 3（caoCsv）・Step 4（googleCalendar）の各プロバイダー実装は本ステップのスコープ外とする。

## Boundary Context

- **In scope**: `holidayjp` プロバイダーの実装・`NextBusinessDay` / `FirstBusinessDayOfMonth` / `FirstBusinessDaysOfYear` / `Holidays` API の実装・各 API のテスト
- **Out of scope**: `caoCsv` プロバイダー・`googleCalendar` プロバイダー・Step 1 の既存 API（`IsBusinessDay`・型定義・インターフェース定義・設定ファイル読み込み）の変更
- **Adjacent expectations**: `HolidayProvider` インターフェースと `BusinessCalendar` 骨格・`IsBusinessDay` は Step 1 で実装済みであることを前提とする。本ステップは Step 1 の既存実装を変更しない。

---

## Requirements

### Requirement 1: holidayjp プロバイダー

**Objective:** As a ライブラリ利用者, I want 外部接続不要で日本の祝日を判定できるデフォルトプロバイダー, so that 最小設定でライブラリを使い始められる

#### Acceptance Criteria

1. The go-heijitu ライブラリ shall provide a `holidayjp.New()` factory function that returns a value implementing `HolidayProvider`, requiring no constructor arguments.
2. When `IsHoliday` is called on the holidayjp provider for a date recognized as a Japanese national holiday, the go-heijitu ライブラリ shall return `true` and `nil` error.
3. When `IsHoliday` is called on the holidayjp provider for a date not recognized as a Japanese national holiday, the go-heijitu ライブラリ shall return `false` and `nil` error.
4. When `HolidayName` is called on the holidayjp provider for a date recognized as a Japanese national holiday, the go-heijitu ライブラリ shall return the holiday name as a non-empty string and `nil` error.
5. When `HolidayName` is called on the holidayjp provider for a date not recognized as a Japanese national holiday, the go-heijitu ライブラリ shall return an empty string and `nil` error.
6. When `HolidaysBetween` is called on the holidayjp provider with a `from` date and a `to` date, the go-heijitu ライブラリ shall return all Japanese national holidays in that range with both endpoints inclusive.
7. The go-heijitu ライブラリ shall not require any external network connection for holidayjp provider operations.

---

### Requirement 2: NextBusinessDay API

**Objective:** As a ライブラリ利用者, I want 指定日の翌日以降で最初の営業日を取得できる, so that 期日計算・締め切り計算に利用できる

#### Acceptance Criteria

1. When `NextBusinessDay(ctx, from)` is called, the go-heijitu ライブラリ shall return the first business day strictly after the `from` date (the `from` date itself is never returned as the result).
2. When the day immediately after `from` falls on a Saturday or Sunday, the go-heijitu ライブラリ shall skip it and continue searching for the next business day.
3. When a candidate date is recognized as a holiday by the `HolidayProvider`, the go-heijitu ライブラリ shall skip it and continue searching for the next business day.
4. When a candidate date matches any excluded date registered via `WithExcludedDates` or `WithConfig`, the go-heijitu ライブラリ shall skip it and continue searching for the next business day.
5. If the `HolidayProvider` returns an error during a `NextBusinessDay` search, the go-heijitu ライブラリ shall propagate that error to the caller without suppressing it.

---

### Requirement 3: FirstBusinessDayOfMonth API

**Objective:** As a ライブラリ利用者, I want 指定年月の最初の営業日を取得できる, so that 月初め処理の日付計算に利用できる

#### Acceptance Criteria

1. When `FirstBusinessDayOfMonth(ctx, year, month)` is called, the go-heijitu ライブラリ shall return the first calendar day of that month that qualifies as a business day.
2. When the 1st day of the specified month is not a business day (weekend, holiday, or excluded date), the go-heijitu ライブラリ shall continue searching subsequent days of that month until a business day is found.
3. If the `HolidayProvider` returns an error during a `FirstBusinessDayOfMonth` search, the go-heijitu ライブラリ shall propagate that error to the caller without suppressing it.

---

### Requirement 4: FirstBusinessDaysOfYear API

**Objective:** As a ライブラリ利用者, I want 指定年の各月の最初の営業日リストを一括取得できる, so that 年間の月初め営業日カレンダーを効率的に生成できる

#### Acceptance Criteria

1. When `FirstBusinessDaysOfYear(ctx, year)` is called, the go-heijitu ライブラリ shall return a slice of exactly 12 `time.Time` values, where index 0 corresponds to January and index 11 corresponds to December of the specified year.
2. When `FirstBusinessDaysOfYear(ctx, year)` is called, each element in the returned slice shall equal the result of `FirstBusinessDayOfMonth` for the corresponding month of that year.
3. If the `HolidayProvider` returns an error for any month during a `FirstBusinessDaysOfYear` call, the go-heijitu ライブラリ shall propagate that error to the caller without suppressing it.

---

### Requirement 5: Holidays API

**Objective:** As a ライブラリ利用者, I want 指定期間内の祝日リストを取得できる, so that 祝日一覧の表示や期間内の祝日数の計算に利用できる

#### Acceptance Criteria

1. When `Holidays(ctx, from, to)` is called, the go-heijitu ライブラリ shall return the list of holidays in the specified range as provided by the `HolidayProvider`, with both `from` and `to` dates inclusive.
2. The go-heijitu ライブラリ shall not include company-specific excluded dates (registered via `WithExcludedDates` or `WithConfig`) in the `Holidays` return value.
3. If the `HolidayProvider` returns an error during a `Holidays` call, the go-heijitu ライブラリ shall propagate that error to the caller without suppressing it.
