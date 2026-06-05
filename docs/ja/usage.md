# 使い方ガイド

本ガイドでは `go-heijitu` のインストールと、よくあるユースケースの使い方を説明します。

## インストール

```bash
go get github.com/taku-o/go-heijitu
```

コアパッケージは標準ライブラリと YAML にのみ依存します。各祝日プロバイダーはサブパッケージに分かれており、その外部依存はパッケージ内に閉じています。

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

## ユースケース

### 営業日かどうかを判定する

```go
ok, err := cal.IsBusinessDay(ctx, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
// → false（1月1日は祝日のため）
```

### 次の営業日を求める

```go
next, err := cal.NextBusinessDay(ctx, time.Now())
fmt.Printf("%s (%s)\n", next.Format("2006-01-02"), next.Weekday())
```

`from` 当日は除外され、その翌日以降で最初の営業日が返ります。

### 指定月の最初の営業日

```go
first, err := cal.FirstBusinessDayOfMonth(ctx, 2026, time.April)
```

### 指定年の各月の最初の営業日

```go
days, err := cal.FirstBusinessDaysOfYear(ctx, 2026)
for i, d := range days {
    fmt.Printf("%2d: %s\n", i+1, d.Format("2006-01-02"))
}
```

スライスは必ず12件です（インデックス0が1月）。

### 期間内の祝日を一覧する

```go
from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.Local)
holidays, err := cal.Holidays(ctx, from, to)
for _, h := range holidays {
    fmt.Printf("%s: %s\n", h.Date.Format("2006-01-02"), h.Name)
}
```

### 会社の休業日をパラメータで除外する

```go
cal := heijitu.New(holidayjp.New(),
    heijitu.WithExcludedDates([]heijitu.MonthDay{
        {Month: time.December, Day: 31},
        {Month: time.January, Day: 2},
    }),
)
```

### 会社の休業日を設定ファイルで除外する

`heijitu.yaml` を用意します。

```yaml
excluded_dates:
  - month: 12
    day: 31
  - month: 1
    day: 2
```

```go
opt, err := heijitu.WithConfig("heijitu.yaml")
if err != nil {
    log.Fatal(err)
}
cal := heijitu.New(holidayjp.New(), opt)
```

`WithExcludedDates` と `WithConfig` は併用でき、除外日付はマージされます。

### 1回の呼び出しだけ日付を除外する

```go
ok, err := cal.IsBusinessDay(ctx,
    time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
    heijitu.MonthDay{Month: time.May, Day: 1},
)
// → false（5/1 をこの呼び出し限りで除外日として扱うため）
```

### 祝日データソースを切り替える

任意の `HolidayProvider` を `heijitu.New` に渡せます。埋め込みデータの代わりに内閣府CSVを使う場合:

```go
provider, err := caoCsv.New(ctx, caoCsv.Options{}) // 公式データをオンライン取得
if err != nil {
    log.Fatal(err)
}
cal := heijitu.New(provider)
```

各プロバイダーの選択基準と設定方法は[プロバイダーガイド](./providers.md)を参照してください。
