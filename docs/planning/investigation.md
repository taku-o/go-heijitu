# 祝日ライブラリ候補調査

## 調査対象

日本の営業日計算ライブラリを構築するにあたり、祝日データの取得・判定手段として以下3種を調査した。

---

## 1. github.com/holiday-jp/holiday_jp-go

### 概要
日本の祝日データをGoコード内に埋め込んだ静的ライブラリ。

### 提供API
```go
holiday.IsHoliday(t time.Time) bool
holiday.Between(from, to time.Time) map[string]string  // 日付文字列 → 祝日名
```

### メリット
- 外部ネットワーク不要
- 依存が少なくシンプル
- 高速（メモリ内データ参照）

### デメリット
- ライブラリ更新なしには最新の祝日に追随できない
- 将来の祝日追加・変更への対応が遅延する可能性

### 判断
シンプルな用途のデフォルト実装として採用候補。外部接続なしで動作する点が強み。

---

## 2. 内閣府CSVデータ

### データソース
`https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv`

### フォーマット
- エンコード: Shift_JIS
- 内容: 年/月/日, 祝日名 の2列
- 範囲: 1955年〜（毎年更新）
- ライセンス: CC-BY（クリエイティブ・コモンズ表示）

### 関連ライブラリ
- `github.com/mikan/syukujitsu-go`: FetchAndParse()（URL取得）、LoadAndParse()（ローカルファイル）
- `github.com/soh335/shukujitsu`: 同CSVを使用、週次自動更新対応

### メリット
- 公式データのため信頼性が高い
- ローカルCSVファイルとして保持すればオフライン動作可能
- URL指定で常に最新データを取得可能

### デメリット
- Shift_JIS変換処理が必要
- URL取得の場合はネットワーク依存
- CSVファイルの更新タイミングに左右される

### 判断
公式データへの直接アクセスが必要なユースケースに有効。ローカルファイルモードとURLモードの両方をサポートすべき。

---

## 3. Google Calendar API

### データソース
Google Calendar の日本の祝日カレンダー
- Calendar ID: `ja.japanese.official#holiday@group.v.calendar.google.com`

### 関連ライブラリ
- `google.golang.org/api/calendar/v3`: Google公式APIクライアント
- `github.com/haruotsu/go-jpholiday`: Google Calendar APIを使った日本祝日判定ライブラリ

### メリット
- Googleが管理・更新するため常に最新
- 祝日名など豊富なメタデータが取得可能

### デメリット
- APIキーまたはOAuth認証が必要
- ネットワーク接続必須
- APIクォータ制限あり（無料枠内でも管理が必要）
- サービス障害時に動作不能になるリスク

### 判断
認証コストが高く、ライブラリに組み込む場合の利用者負担が大きい。オプション実装として提供する形が現実的。

---

## vulsio/go-holiday-jp について

ユーザー提示の候補だが、調査時点でGitHub上に該当リポジトリが確認できなかった。上記3候補に置き換えて検討を進める。

---

## 比較まとめ

| 項目 | holiday_jp-go | 内閣府CSV | Google Calendar API |
|------|--------------|-----------|---------------------|
| ネットワーク不要 | ✓ | ローカルモード時✓ | ✗ |
| 認証不要 | ✓ | ✓ | ✗ |
| 公式データ | ✗（独自管理） | ✓ | △（Google管理） |
| 最新性 | ライブラリ更新依存 | CSV更新依存 | 高い |
| 実装コスト | 低 | 中 | 高 |
| 外部依存 | ライブラリ | なし（ローカル時） | Google APIクライアント |

## 推奨プロバイダー

- **デフォルト**: `holiday_jp-go`（手軽・外部不要）
- **本番推奨**: 内閣府CSV（公式データ・ローカルキャッシュ対応）
- **オプション**: Google Calendar API（常時最新が必要な場合）
