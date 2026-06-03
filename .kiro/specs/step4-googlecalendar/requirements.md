# 要件定義書: Step 4 — googleCalendar（Google Calendar API）プロバイダー実装

## Introduction

日本の営業日計算ライブラリ `go-heijitu` において、Google Calendar が提供する日本の祝日カレンダーを祝日データソースとして利用する `googleCalendar` プロバイダーを実装する。

`googleCalendar` プロバイダーは Step 1・Step 2 で確立済みの `HolidayProvider` インターフェースを満たす実装であり、APIキー認証と OAuth2 サービスアカウント認証の2方式に対応する。認証方式は `Options.APIKey` / `Options.CredentialsFile` で指定し、両方が指定された場合は `CredentialsFile`（OAuth2 サービスアカウント）を優先する。祝日データは固定の Calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com` から取得する。本ステップ完了後、利用者は `holidayjp` / `caoCsv` プロバイダーの代わりに Google が管理する祝日カレンダーに基づく `googleCalendar` プロバイダーを `BusinessCalendar` に渡して、すべての既存 API を利用できる状態になる。

Step 1・Step 2・Step 3 で実装済みのコア機能（`BusinessCalendar` の各 API・型定義・インターフェース定義・設定ファイル読み込み・`holidayjp` プロバイダー・`caoCsv` プロバイダー）の変更は、本ステップのスコープ外とする。

## Boundary Context

- **In scope**: `googleCalendar` プロバイダーの実装（コンストラクタ `New`・APIキー認証・OAuth2 サービスアカウント認証・認証方式の選択と優先順位・固定 Calendar ID からの祝日取得・`HolidayProvider` の3メソッド実装）・本プロバイダーのテスト（ネットワーク非依存の契約テストと、`//go:build integration` タグで分離した実 API 呼び出しテスト）
- **Out of scope**: `holidayjp` / `caoCsv` プロバイダーの変更・`BusinessCalendar` の各 API や型定義・インターフェース定義・設定ファイル読み込みの変更・取得した祝日データの永続キャッシュ機構・固定 Calendar ID 以外の任意カレンダーの指定・APIクォータ制限やリトライの独自制御
- **Adjacent expectations**: `HolidayProvider` インターフェース（`IsHoliday` / `HolidayName` / `HolidaysBetween`）と `BusinessCalendar` は Step 1・Step 2 で実装済みであることを前提とする。`googleCalendar` プロバイダーは既存のインターフェース契約（祝日でない日には空文字を返す、`HolidaysBetween` は両端を含み日付昇順、エラーは握りつぶさず伝播する）に従う。Google Calendar API へのアクセスにはネットワーク接続と有効な認証情報（APIキーまたはサービスアカウント）が必要であり、それらの発行・管理は利用者の責任とする。

---

## Requirements

### Requirement 1: googleCalendar プロバイダーの生成と認証方式の選択

**Objective:** As a ライブラリ利用者, I want APIキーか OAuth2 サービスアカウントかを選んでプロバイダーを生成できる, so that 自分の Google Cloud 環境に合った認証方式で祝日データを利用できる

#### Acceptance Criteria

1. The googleCalendar プロバイダー shall provide a `New(ctx, opts)` factory function that accepts an `Options` value containing an `APIKey` field and a `CredentialsFile` field, and returns a value implementing the `HolidayProvider` interface together with an error.
2. When `New` is called with a non-empty `CredentialsFile`, the googleCalendar プロバイダー shall authenticate using the OAuth2 service account regardless of whether `APIKey` is also provided.
3. When `New` is called with a non-empty `APIKey` and an empty `CredentialsFile`, the googleCalendar プロバイダー shall authenticate using the API key.
4. If `New` is called with both `APIKey` and `CredentialsFile` empty, then the googleCalendar プロバイダー shall return an error without accessing the network.
5. When `New` returns successfully, the googleCalendar プロバイダー shall be ready to serve `HolidayProvider` method calls against the Google Calendar holiday data source.

---

### Requirement 2: APIキー認証

**Objective:** As a ライブラリ利用者, I want APIキーだけで手軽にプロバイダーを生成できる, so that サービスアカウントを用意せずに最小設定で祝日データへアクセスできる

#### Acceptance Criteria

1. While operating in API key authentication mode, the googleCalendar プロバイダー shall use the supplied `APIKey` to access the Google Calendar API.
2. If the supplied `APIKey` is rejected by the Google Calendar API during a holiday data request, then the googleCalendar プロバイダー shall return the resulting error to the caller without suppressing it.

---

### Requirement 3: OAuth2 サービスアカウント認証

**Objective:** As a ライブラリ利用者, I want サービスアカウントの認証情報ファイルを指定して認証できる, so that APIキーを使えない環境でも安全に祝日データへアクセスできる

#### Acceptance Criteria

1. While operating in OAuth2 service account authentication mode, the googleCalendar プロバイダー shall load the service account credentials from the file located at `CredentialsFile` and use them to access the Google Calendar API.
2. If the file at `CredentialsFile` does not exist, cannot be read, or cannot be parsed as valid service account credentials, then the googleCalendar プロバイダー shall return an error from `New` describing the failure, without suppressing it.

---

### Requirement 4: Google Calendar からの祝日データ取得

**Objective:** As a ライブラリ利用者, I want Google が管理する日本の祝日カレンダーから祝日を取得してもらえる, so that 常に最新の祝日データに基づいて営業日を計算できる

#### Acceptance Criteria

1. The googleCalendar プロバイダー shall retrieve holiday data from the fixed Calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com`.
2. The googleCalendar プロバイダー shall map each retrieved calendar event to a holiday date and its corresponding holiday name.
3. The googleCalendar プロバイダー shall require a network connection to access the Google Calendar API.
4. If a request to the Google Calendar API fails, then the googleCalendar プロバイダー shall return the resulting error to the caller without suppressing it.

---

### Requirement 5: HolidayProvider インターフェースの実装

**Objective:** As a ライブラリ利用者, I want googleCalendar プロバイダーを既存の BusinessCalendar にそのまま渡して全 API を使える, so that データソースを Google Calendar へ差し替えても営業日計算の挙動が一貫している

#### Acceptance Criteria

1. When `IsHoliday` is called on the googleCalendar プロバイダー for a date present in the Google Calendar holiday data, the googleCalendar プロバイダー shall return `true` and `nil` error.
2. When `IsHoliday` is called on the googleCalendar プロバイダー for a date not present in the holiday data, the googleCalendar プロバイダー shall return `false` and `nil` error.
3. When `HolidayName` is called on the googleCalendar プロバイダー for a date present in the holiday data, the googleCalendar プロバイダー shall return that date's holiday name as a non-empty string and `nil` error.
4. When `HolidayName` is called on the googleCalendar プロバイダー for a date not present in the holiday data, the googleCalendar プロバイダー shall return an empty string and `nil` error.
5. When `HolidaysBetween(ctx, from, to)` is called on the googleCalendar プロバイダー, the googleCalendar プロバイダー shall return all holidays within that range, with both `from` and `to` inclusive, in ascending date order.

> **テスト方針（受け入れ基準ではない）**: Google Calendar API への実呼び出しを伴うテスト（Requirement 2・3・4・5 のデータ取得挙動）は、ネットワークと有効な認証情報に依存するため `//go:build integration` タグで分離し、通常の `go test ./...` では実行しない。通常テストでは、認証情報が両方空のときに `New` がネットワークアクセスなしでエラーを返すこと（Requirement 1 の AC 4）など、ネットワーク非依存の契約を検証する。
