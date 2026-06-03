# ソースコード構成

## ディレクトリ構成

```
go-heijitu/
├── go.mod
├── go.sum
│
├── calendar.go          # BusinessCalendar 本体・公開API
├── holiday.go           # Holiday 型定義
├── monthday.go          # MonthDay 型定義
├── provider.go          # HolidayProvider インターフェース定義
├── option.go            # Option 型・WithExcludedDates・WithConfig
├── config.go            # 設定ファイル読み込み（YAML/JSON判別・パース）
│
├── providers/
│   ├── holidayjp/
│   │   └── provider.go  # holiday-jp/holiday_jp-go を使った実装
│   ├── caoCsv/
│   │   └── provider.go  # 内閣府CSVを使った実装（ローカル/URL）
│   └── googleCalendar/
│       └── provider.go  # Google Calendar APIを使った実装
│
└── example/
    └── main.go          # 利用サンプル
```

---

## 各ファイルの責務

### calendar.go

`BusinessCalendar` 構造体と、ユーザーが呼び出す全APIメソッドを実装する。  
`HolidayProvider` に依存し、土日・祝日・除外日付の判定ロジックをここで統合する。

```
依存: provider.go, monthday.go, holiday.go
```

### holiday.go

祝日1件を表す `Holiday` 型を定義する。

```go
type Holiday struct {
    Date time.Time
    Name string
}
```

### monthday.go

年をまたいで有効な月日を表す `MonthDay` 型を定義する。  
除外日付（会社独自の休業日）の指定に使用。

```go
type MonthDay struct {
    Month time.Month
    Day   int
}
```

一致判定メソッドも持つ。

```go
func (md MonthDay) Matches(t time.Time) bool
```

### provider.go

祝日判定の実装を差し替えるための `HolidayProvider` インターフェースを定義する。  
3つのメソッドを持つ。

```
依存: holiday.go
```

### option.go

`BusinessCalendar` の構築オプションを定義する。  
`WithExcludedDates`（パラメータ指定）と `WithConfig`（設定ファイル指定）を提供。

```
依存: monthday.go, config.go
```

### config.go

設定ファイル（YAML/JSON）を読み込み、`Config` 型に変換する。  
ファイル拡張子（`.yaml` / `.yml` / `.json`）から形式を自動判別する。

```go
type Config struct {
    ExcludedDates []MonthDay `yaml:"excluded_dates" json:"excluded_dates"`
}
```

```
外部依存: gopkg.in/yaml.v3（YAML読み込み時）
```

---

## providers/ の各ファイル

### providers/holidayjp/provider.go

`github.com/holiday-jp/holiday_jp-go` を内部で使用する `HolidayProvider` 実装。  
外部ネットワーク不要。データはライブラリに埋め込まれている。

```go
type Provider struct{}

func New() *Provider
// HolidayProvider インターフェースを実装
```

```
外部依存: github.com/holiday-jp/holiday_jp-go
```

### providers/caoCsv/provider.go

内閣府CSV（`syukujitsu.csv`）を読み込む `HolidayProvider` 実装。  
ローカルCSVファイルパスを `Options.CSVPath` で指定する。空の場合は内閣府公式データをオンライン取得する。  
内部で mikan のパース結果（`[]syukujitsu.Entry`）を保持し、点照合は `Find` に委譲する。

```go
type Provider struct {
    entries []syukujitsu.Entry
}

type Options struct {
    CSVPath string // ローカルファイルパス。空の場合は内閣府公式データをオンライン取得する
}

func New(ctx context.Context, opts Options) (*Provider, error)
// HolidayProvider インターフェースを実装
```

```
外部依存: github.com/mikan/syukujitsu-go（CSV取得・パーサー、Shift_JISデコードを内部処理）
         golang.org/x/text（mikan/syukujitsu-go 経由の推移的依存）
```

### providers/googleCalendar/provider.go

Google Calendar APIから日本の祝日カレンダーを取得する `HolidayProvider` 実装。  
Calendar ID: `ja.japanese.official#holiday@group.v.calendar.google.com`

```go
type Provider struct{}

type Options struct {
    APIKey          string // APIキー認証
    CredentialsFile string // OAuth2サービスアカウントJSONファイルパス
}

func New(ctx context.Context, opts Options) (*Provider, error)
// HolidayProvider インターフェースを実装
```

```
外部依存: google.golang.org/api/calendar/v3
         golang.org/x/oauth2/google
```

---

## 依存関係の全体図

```
calendar.go
  └── provider.go (interface)
        └── providers/holidayjp/provider.go  ← holiday-jp/holiday_jp-go
        └── providers/caoCsv/provider.go     ← mikan/syukujitsu-go (→ x/text)
        └── providers/googleCalendar/provider.go ← google.golang.org/api
  └── monthday.go
  └── holiday.go

option.go
  └── config.go
        └── gopkg.in/yaml.v3
  └── monthday.go
```

コア（`calendar.go`, `provider.go`, `monthday.go`, `holiday.go`）は標準ライブラリのみに依存する。  
外部ライブラリへの依存はプロバイダーと設定ファイル読み込みに閉じ込める。

---

## go.mod の想定

```
module github.com/taku-o/go-heijitu

go 1.23

require (
    github.com/holiday-jp/holiday_jp-go v0.0.0-...  // holidayjpプロバイダー
    github.com/mikan/syukujitsu-go v...              // caoCsvプロバイダー（CSV取得・パース・Shift_JISデコード）
    golang.org/x/text v...                           // mikan/syukujitsu-go 経由の推移的依存
    google.golang.org/api v...                       // googleCalendarプロバイダー
    golang.org/x/oauth2 v...                         // googleCalendarプロバイダー
    gopkg.in/yaml.v3 v...                            // 設定ファイル読み込み
)
```
