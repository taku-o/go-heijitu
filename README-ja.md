# go-heijitu

[English](./README.md)

`go-heijitu` は日本の営業日を計算する Go ライブラリです。指定日が営業日か（土日・祝日・除外日付でないか）を判定し、次の営業日・指定年月の最初の営業日・指定年の各月初営業日・期間内の祝日一覧を求めます。祝日データソースは `HolidayProvider` インターフェースで差し替え可能です。

## インストール

```bash
go get github.com/taku-o/go-heijitu
```

## クイックスタート

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
    fmt.Println("今日は営業日:", ok)
}
```

## プロバイダー

祝日データソースは `heijitu.New` に `HolidayProvider` を渡して選択します。

- **holidayjp** — 埋め込み祝日データ。ネットワークアクセス不要。
- **caoCsv** — 内閣府公式CSV。ローカルファイル／オンライン。
- **googleCalendar** — Google Calendar の祝日カレンダー。APIキー／サービスアカウント。

`googleCalendar` プロバイダーには Google Calendar APIキーが必要です。要点: Google Cloud Console でプロジェクトを作成し、Google Calendar API を有効化、「**APIとサービス → 認証情報**」でAPIキーを作成し、（推奨）キーを Calendar API のみに制限します。詳細な手順と integration テストの実行方法は[プロバイダーガイド](./docs/ja/providers.md#google-calendar-apiキーの取得手順)を参照してください。

## ドキュメント

- [API仕様](./docs/ja/api-spec.md)
- [使い方ガイド](./docs/ja/usage.md)
- [プロバイダーガイド](./docs/ja/providers.md)

## ライセンス

MIT ライセンス。[LICENSE](./LICENSE) を参照してください。
