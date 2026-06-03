# 要件定義書: Step 3 — caoCsv（内閣府CSV）プロバイダー実装

## Introduction

日本の営業日計算ライブラリ `go-heijitu` において、内閣府が公開する公式祝日CSV（`syukujitsu.csv`）を祝日データソースとして利用する `caoCsv` プロバイダーを実装する。

`caoCsv` プロバイダーは Step 1・Step 2 で確立済みの `HolidayProvider` インターフェースを満たす実装であり、ローカルCSVファイルから読み込むモードと、内閣府公式データをオンラインで取得するモードの2モードに対応する。データソースは `Options.CSVPath` で切り替え、`CSVPath` を指定した場合はそのローカルファイルを、`CSVPath` が空の場合は内閣府公式データをオンライン取得する。本ステップ完了後、利用者は `holidayjp` プロバイダーの代わりに公式データに基づく `caoCsv` プロバイダーを `BusinessCalendar` に渡して、すべての既存 API を利用できる状態になる。

Step 4（googleCalendar）プロバイダーの実装、および Step 1・Step 2 で実装済みのコア機能（`BusinessCalendar` の各 API・型定義・インターフェース定義・設定ファイル読み込み・`holidayjp` プロバイダー）の変更は、本ステップのスコープ外とする。

## Boundary Context

- **In scope**: `caoCsv` プロバイダーの実装（コンストラクタ `New`・ローカルCSVファイルモード・内閣府公式データのオンライン取得モード・Shift_JIS デコード・CSVパース・`HolidayProvider` の3メソッド実装）・本プロバイダーのテスト
- **Out of scope**: `googleCalendar` プロバイダー・`holidayjp` プロバイダーの変更・`BusinessCalendar` の各 API や型定義・インターフェース定義・設定ファイル読み込みの変更・取得済みCSVの永続キャッシュ機構・内閣府公式データソース以外の任意URLからの取得
- **Adjacent expectations**: `HolidayProvider` インターフェース（`IsHoliday` / `HolidayName` / `HolidaysBetween`）と `BusinessCalendar` は Step 1・Step 2 で実装済みであることを前提とする。`caoCsv` プロバイダーは既存のインターフェース契約（祝日でない日には空文字を返す、`HolidaysBetween` は両端を含み日付昇順、エラーは握りつぶさず伝播する）に従う。

---

## Requirements

### Requirement 1: caoCsv プロバイダーの生成とデータソース選択

**Objective:** As a ライブラリ利用者, I want ローカルCSVファイルか内閣府公式オンラインデータかを選んでプロバイダーを生成できる, so that 公式データに基づいた営業日計算を最小設定で始められる

#### Acceptance Criteria

1. The caoCsv プロバイダー shall provide a `New(ctx, opts)` factory function that accepts an `Options` value containing a `CSVPath` field, and returns a value implementing the `HolidayProvider` interface together with an error.
2. When `New` is called with a non-empty `CSVPath`, the caoCsv プロバイダー shall load holiday data from the local CSV file at that path.
3. When `New` is called with an empty `CSVPath`, the caoCsv プロバイダー shall load holiday data by fetching the CSV from the official 内閣府 (Cabinet Office) data source online.
4. When `New` returns successfully, the caoCsv プロバイダー shall have already loaded and parsed all holiday data so that subsequent `HolidayProvider` method calls require no additional file read or network access.

---

### Requirement 2: ローカルCSVファイルモード

**Objective:** As a ライブラリ利用者, I want 手元に保存した内閣府CSVファイルから祝日データを読み込める, so that ネットワーク接続なしのオフライン環境でも公式データを利用できる

#### Acceptance Criteria

1. When `New` loads data in local file mode, the caoCsv プロバイダー shall read the CSV content from the file located at `CSVPath`.
2. If the file at `CSVPath` does not exist or cannot be read, the caoCsv プロバイダー shall return an error from `New` describing the read failure, without suppressing it.
3. The caoCsv プロバイダー shall not require any network connection when operating in local file mode.

---

### Requirement 3: オンライン取得モード（内閣府公式データ）

**Objective:** As a ライブラリ利用者, I want CSVPath を指定せずに内閣府公式の最新祝日データを取得できる, so that 手元にファイルを用意せず常に最新の公式データで営業日を計算できる

#### Acceptance Criteria

1. When `New` loads data in online mode (empty `CSVPath`), the caoCsv プロバイダー shall fetch the holiday CSV from the official 内閣府 (Cabinet Office) data source.
2. If the online fetch fails or returns content that cannot be obtained, the caoCsv プロバイダー shall return an error from `New` describing the fetch failure, without suppressing it.
3. The caoCsv プロバイダー shall fetch the CSV at the time `New` is called and shall not persist the fetched data to any cache beyond the in-memory data of the returned provider.

---

### Requirement 4: CSVデータのデコードとパース

**Objective:** As a ライブラリ利用者, I want 内閣府CSVのShift_JIS形式と列構成を正しく解釈してもらえる, so that 文字化けや解釈誤りなく祝日名と祝日日付を扱える

#### Acceptance Criteria

1. The caoCsv プロバイダー shall decode the CSV content from Shift_JIS encoding to UTF-8 before parsing the rows.
2. The caoCsv プロバイダー shall parse each holiday row of the 内閣府CSV into a date and its corresponding holiday name.
3. While parsing the CSV, the caoCsv プロバイダー shall exclude the header row from the holiday data.
4. If the CSV content cannot be decoded or parsed, the caoCsv プロバイダー shall return an error from `New` describing the failure, without suppressing it.

---

### Requirement 5: HolidayProvider インターフェースの実装

**Objective:** As a ライブラリ利用者, I want caoCsv プロバイダーを既存の BusinessCalendar にそのまま渡して全 API を使える, so that データソースを公式CSVへ差し替えても営業日計算の挙動が一貫している

#### Acceptance Criteria

1. When `IsHoliday` is called on the caoCsv プロバイダー for a date present in the loaded 内閣府CSV holiday data, the caoCsv プロバイダー shall return `true` and `nil` error.
2. When `IsHoliday` is called on the caoCsv プロバイダー for a date not present in the loaded holiday data, the caoCsv プロバイダー shall return `false` and `nil` error.
3. When `HolidayName` is called on the caoCsv プロバイダー for a date present in the loaded holiday data, the caoCsv プロバイダー shall return that date's holiday name as a non-empty string and `nil` error.
4. When `HolidayName` is called on the caoCsv プロバイダー for a date not present in the loaded holiday data, the caoCsv プロバイダー shall return an empty string and `nil` error.
5. When `HolidaysBetween(ctx, from, to)` is called on the caoCsv プロバイダー, the caoCsv プロバイダー shall return all holidays in the loaded data within that range, with both `from` and `to` inclusive, in ascending date order.

> **テスト方針（受け入れ基準ではない）**: ローカルCSVモードの祝日判定の妥当性は、同一日付について holidayjp プロバイダーの結果と突き合わせて検証する。両者は内閣府公式祝日についておおむね一致する想定だが、対象年の差や振替休日の扱いで差異が生じうるため、holidayjp との完全一致は契約上の受け入れ基準とはしない。受け入れ基準は本要件の AC 1〜5（読み込んだ内閣府CSVデータに対する挙動）で判定する。
