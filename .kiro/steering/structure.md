# プロジェクト構成

## 構成方針

ライブラリのルートパッケージ（コア）と、祝日データソースごとのプロバイダーサブパッケージで構成する。1ファイル1責務を基本とし、コアは標準ライブラリ中心、外部依存は各プロバイダーパッケージに閉じ込める。

## ディレクトリパターン

### コア（ルートパッケージ）
**場所**: `/`（`package heijitu`）
**責務**: 営業日計算の公開 API・型・インターフェース。1つの型/関心ごとに1ファイル。
**例**: `calendar.go`（`BusinessCalendar` と公開 API）、`provider.go`（`HolidayProvider` インターフェース）、`monthday.go` / `holiday.go`（値オブジェクト）、`option.go`（関数オプション）、`config.go`（設定ファイル読み込み）

### プロバイダー
**場所**: `/providers/<name>/`（`package <name>`）
**責務**: `HolidayProvider` の具体実装。1プロバイダー＝1パッケージで、その外部依存をこのパッケージに閉じ込める。
**例**: `providers/holidayjp/provider.go`。親モジュールは `heijitu "github.com/taku-o/go-heijitu"` のエイリアスで参照する。

### サンプル・テストデータ
**場所**: `/example/`（利用サンプル、整備予定）、各パッケージの `testdata/`（テスト用フィクスチャ）

## 命名規則

- **ファイル**: snake_case（Go 標準）。テストは `対象_test.go`。規模の大きい API はファイル分割（例: `calendar_first_business_day_test.go`）
- **パッケージ**: lowercase の短い名前（コアは `heijitu`、プロバイダーはディレクトリ名）
- **公開シンボル**: CamelCase（公開）/ camelCase（非公開）

## インポート整理

```go
import (
    // 標準ライブラリ
    "context"
    "time"

    // 親モジュール（プロバイダーパッケージから）
    heijitu "github.com/taku-o/go-heijitu"

    // 外部ライブラリ
    holiday "github.com/holiday-jp/holiday_jp-go"
)
```

## コード構成の原則

- **依存方向**: プロバイダーがコア（親）を参照する。コアはプロバイダーを import しない（利用者が呼び出し時に注入する）
- **インターフェースは利用側（コア）で定義**する（Go の慣習）
- パッケージは単一の責務を持ち、循環インポートを避ける
- 外部ライブラリへの依存はプロバイダーと設定ファイル読み込みに限定し、コアの計算ロジックは標準ライブラリに保つ

---
_ファイルツリーではなくパターンを文書化する。パターンに従う新規ファイルはステアリング更新を不要にする。最終同期: 2026-06-03（コードベースから再構築）_
