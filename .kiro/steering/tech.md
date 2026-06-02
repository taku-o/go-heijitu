# 技術スタック

## アーキテクチャ

公開 Go ライブラリ（実行バイナリを持たない）。ルートパッケージ `heijitu` が営業日計算のコアと公開 API・型・`HolidayProvider` インターフェースを提供し、`providers/<name>/` 配下の各パッケージが祝日データソースごとの `HolidayProvider` 実装を担う。コアは `HolidayProvider` インターフェースにのみ依存し、具体的なプロバイダー実装は利用者が呼び出し時に注入する。

## コア技術

- **言語**: Go 1.23.4
- **モジュール**: `github.com/taku-o/go-heijitu`
- **配布形態**: ライブラリ（`main` パッケージなし）

## 主要ライブラリ

依存は「コア」と「各プロバイダー」で分離する。コアは標準ライブラリ + YAML のみに依存し、外部ライブラリへの依存はプロバイダーパッケージに閉じ込める。

- `gopkg.in/yaml.v3`: 設定ファイル読み込み（コア）
- `github.com/holiday-jp/holiday_jp-go`: holidayjp プロバイダー（埋め込み祝日データ）
- `github.com/mikan/syukujitsu-go`（caoCsv プロバイダーで採用予定）: 内閣府CSVの取得・パース・Shift_JISデコード
- `google.golang.org/api`・`golang.org/x/oauth2`（googleCalendar プロバイダーで採用予定）

## 開発標準

### コードパターン
- **関数オプションパターン**: `type Option func(*BusinessCalendar)`、`WithExcludedDates` / `WithConfig` で構築をカスタマイズ
- **`context.Context` 第一引数**: `HolidayProvider` と `BusinessCalendar` の各メソッドは `ctx context.Context` を第一引数に取る
- **インターフェースは利用側で定義**: `HolidayProvider` はコア（利用側）で定義し、各プロバイダーが実装する

### エラーハンドリング
- エラーは握りつぶさず呼び出し元へ伝播する
- `HolidayName` は非祝日のとき `("", nil)` を返す（エラーにしない）
- プログラマエラー（`New` への nil プロバイダー）のみ `panic`。実行時の入力エラーは error で返す

### テスト
- 標準 `testing` パッケージ、テーブル駆動 + Given-When-Then コメント
- 外部テストパッケージ（`package <name>_test`）で公開 API 越しに検証
- `var _ heijitu.HolidayProvider = (*Provider)(nil)` でインターフェース充足をコンパイル時に保証
- フィクスチャは `testdata/` に配置
- 規模の大きい API はファイルを分割（例: `calendar_next_business_day_test.go`）

### コード品質
- `gofmt` 準拠、Go 標準のコーディング規約に従う

## 開発環境

### 必要ツール
- Go 1.23.4

### 主要コマンド
```bash
go build ./...   # ビルド
go test ./...    # テスト
go vet ./...     # 静的解析
```

## 主要な技術的決定事項

- **プロバイダー抽象**: 祝日データソースを `HolidayProvider` で差し替え可能にし、外部依存をプロバイダーごとに分離
- **設定ファイル**: YAML / JSON を拡張子（`.yaml` / `.yml` / `.json`）で自動判別
- **キャッシュなし**: オンライン取得型プロバイダーは取得結果を永続キャッシュしない
- **振替休日**: ライブラリ側では関与せず、各プロバイダーの祝日データに委ねる

---
_標準とパターンを文書化し、全ての依存関係は列挙しない。最終同期: 2026-06-03（コードベースから再構築）_
