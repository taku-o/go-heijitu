# プロバイダーガイド

`HolidayProvider` は `BusinessCalendar` が使用する祝日データを提供します。`go-heijitu` には3つのプロバイダーが同梱されています。本ガイドでは選択基準・設定方法・注意点を説明します。

## プロバイダーの選択

| プロバイダー | データソース | ネットワーク | 認証 | オフライン可否 |
|------------|------------|------------|------|-------------|
| `holidayjp` | `holiday-jp/holiday_jp-go` の埋め込みデータ | 不要 | なし | 可 |
| `caoCsv` | 内閣府公式CSV（ローカルファイル／オンライン） | オンラインモードで必要 | なし | 可（ローカルCSVモード） |
| `googleCalendar` | Google Calendar の日本祝日カレンダー | 必要 | APIキー または OAuth2 サービスアカウント | 不可 |

- **holidayjp**: 最も手軽で依存が軽く、完全オフラインで使いたい場合。
- **caoCsv**: 内閣府公式データを使いたい場合。同梱CSV（オフライン）またはオンライン取得を選べます。
- **googleCalendar**: Google が管理する祝日カレンダーを使いたい場合。

## holidayjp

```go
import "github.com/taku-o/go-heijitu/providers/holidayjp"

provider := holidayjp.New()
cal := heijitu.New(provider)
```

オプションなし・ネットワークアクセスなし。祝日データは依存ライブラリに埋め込まれています。

## caoCsv

```go
import "github.com/taku-o/go-heijitu/providers/caoCsv"

// ローカルCSVモード
provider, err := caoCsv.New(ctx, caoCsv.Options{CSVPath: "syukujitsu.csv"})

// オンラインモード（公式データを取得）
provider, err := caoCsv.New(ctx, caoCsv.Options{})
```

- `CSVPath` を指定した場合はローカルCSVを読み込みます（オフラインで動作）。
- `CSVPath` が空の場合は内閣府公式データをオンライン取得します。ネットワーク接続が必要です。
- CSVの取得・Shift_JISデコード・パースは `github.com/mikan/syukujitsu-go` に委譲します。

**注意:** オンラインモードは HTTP リクエストを行うため、ネットワークの可用性に依存します。

## googleCalendar

```go
import "github.com/taku-o/go-heijitu/providers/googleCalendar"

// APIキー認証
provider, err := googleCalendar.New(ctx, googleCalendar.Options{APIKey: apiKey})

// OAuth2 サービスアカウント認証
provider, err := googleCalendar.New(ctx, googleCalendar.Options{CredentialsFile: "service-account.json"})
```

- `APIKey` と `CredentialsFile` の両方を指定した場合は `CredentialsFile` を優先します。
- 両方が空の場合、`New` はネットワークアクセスせずにエラーを返します。
- プロバイダーは固定の Calendar ID `ja.japanese.official#holiday@group.v.calendar.google.com` から取得します。

**注意（呼び出しコスト）:** 本プロバイダーは各メソッド呼び出しごとに Calendar API リクエストを発行します。`NextBusinessDay` / `FirstBusinessDayOfMonth` / `FirstBusinessDaysOfYear` は内部で `IsHoliday` を日単位で呼び出すため、探索した日数ぶんのAPI往復が発生します。holidayjp / caoCsv（メモリ内）と比べてレイテンシ・APIクォータ消費が大きくなります。

### Google Calendar APIキーの取得手順

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを作成（または既存を選択）します。
2. 「**APIとサービス → ライブラリ**」で **Google Calendar API** を有効化します。
3. 「**APIとサービス → 認証情報 → 認証情報を作成 → APIキー**」を選択します。
4. 作成された鍵を Calendar API のみに制限します（推奨）。キーの設定で「**APIの制限**」を開き、Calendar API のみを許可します。

認証情報の発行・保管・ローテーションは利用者の責任です。APIキーをソース管理にコミットしないでください。

### integration テストの実行

実 Calendar API を呼び出すテストは `//go:build integration` タグで分離されており、通常の `go test ./...` では実行されません。実行するには環境変数でAPIキーを渡します。

```bash
export GOOGLE_CALENDAR_API_KEY=<取得した鍵>
go test -tags integration ./providers/googleCalendar/...
```

`GOOGLE_CALENDAR_API_KEY` が未設定の場合、integration テストはスキップされるため、認証情報のない環境でも `go test -tags integration` は失敗しません。
