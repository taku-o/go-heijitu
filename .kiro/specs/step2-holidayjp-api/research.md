# ギャップ分析: Step 2 — holidayjp プロバイダー + 残りAPI実装

## 1. 現状調査

### 既存コードベースの構成

```
（ルートパッケージ: github.com/taku-o/go-heijitu）
calendar.go       — BusinessCalendar 構造体・New()・IsBusinessDay()・isExcluded()
option.go         — Option 型・WithExcludedDates()・WithConfig()
provider.go       — HolidayProvider インターフェース定義
holiday.go        — Holiday 型
monthday.go       — MonthDay 型・Matches()
config.go         — loadConfig()（YAML/JSON 読み込み）
calendar_test.go  — IsBusinessDay・WithExcludedDates・WithConfig のテスト（heijitu_test パッケージ）
provider_test.go  — testProvider モック・HolidayProvider 動作テスト（heijitu_test パッケージ）
setup_test.go     — YAML ライブラリのスモークテスト
```

**providers/ ディレクトリは存在しない** — Step 2 で新規作成が必要。

### 既存コードの再利用可能な資産

| 資産 | 場所 | Step 2 での利用方針 |
|------|------|---------------------|
| `HolidayProvider` インターフェース | `provider.go` | holidayjp プロバイダーが実装すべきインターフェース |
| `isExcluded(t, dates)` | `calendar.go` | `NextBusinessDay` / `FirstBusinessDayOfMonth` の除外日付チェックに再利用可 |
| `IsBusinessDay()` | `calendar.go` | 新規 API 内部でループ処理に直接呼び出し可 |
| `Holiday` 型 | `holiday.go` | `HolidaysBetween` の戻り値に使用 |
| `testProvider` モック | `provider_test.go` (heijitu_test) | 既存テストパターンを参考にできる |
| `dateLayout` 定数 | `provider_test.go` (heijitu_test) | 新規テストファイルに同等定数を定義する必要あり（別パッケージのため） |

### テスト規約の確認

- テストパッケージ: `heijitu_test`（ブラックボックステスト）
- テストスタイル: テーブル駆動テスト（複数ケースは `tests := []struct{...}` パターン）
- テストヘルパー: `t.TempDir()` で一時ファイル作成
- 新規プロバイダーテスト（`providers/holidayjp/`）は独立パッケージで上記規約を踏襲する

---

## 2. 要件のフィージビリティ分析

### 技術的必要事項マップ

| 要件 | 必要事項 | 現状 | ギャップ |
|------|----------|------|---------|
| Req 1: holidayjp プロバイダー | `providers/holidayjp/` パッケージ新規作成 | なし | **Missing** — ディレクトリ・ファイルとも未作成 |
| Req 1: holidayjp プロバイダー | `github.com/holiday-jp/holiday_jp-go` 依存追加 | go.mod に未記載 | **Missing** — `go get` が必要 |
| Req 1: holidayjp プロバイダー | `HolidayProvider` インターフェースの実装 | インターフェース定義あり | **Missing** — ラッパー実装が必要 |
| Req 2: NextBusinessDay | `calendar.go` にメソッド追加 | `IsBusinessDay` と `isExcluded` あり | **Missing** — ループロジック未実装 |
| Req 3: FirstBusinessDayOfMonth | `calendar.go` にメソッド追加 | 上記と同じ | **Missing** — 月内ループロジック未実装 |
| Req 4: FirstBusinessDaysOfYear | `calendar.go` にメソッド追加 | Req 3 が前提 | **Missing** — 12回ループ未実装 |
| Req 5: Holidays | `calendar.go` にメソッド追加 | `HolidayProvider.HolidaysBetween` あり | **Missing** — 委譲メソッド未実装 |

### 外部ライブラリ APIとの差異（重要）

`github.com/holiday-jp/holiday_jp-go` の実際の API は `go-heijitu` の `HolidayProvider` インターフェースと以下の差異がある:

| go-heijitu インターフェース | holiday_jp-go の対応 API | 差異 |
|-----------------------------|--------------------------|------|
| `IsHoliday(ctx, t) (bool, error)` | `holiday.IsHoliday(t) bool` | エラー返却なし → ラッパーで `nil` error を補完するだけで対応可 |
| `HolidayName(ctx, t) (string, error)` | `holiday.HolidayName(t) (string, error)` | **非祝日時にエラーを返す**（空文字ではない）→ 設計フェーズで変換ロジックの検討が必要 |
| `HolidaysBetween(ctx, from, to) ([]Holiday, error)` | `holiday.Between(t0, t1) holiday.Holidays` | 型が異なる・エラー返却なし → `heijitu.Holiday` スライスへの変換が必要 |

**`HolidayName` の差異は特に注意:** holiday_jp-go では非祝日に対してエラーを返すが、go-heijitu の `HolidayProvider` 契約では空文字を返すことが要件（Req 1.5）。
ラッパーでは「非祝日エラーを検出して空文字に変換」するか、「先に `IsHoliday` で判定してから名前取得」するか、いずれかの戦略が必要。

---

## 3. 実装アプローチの選択肢

### Option A: calendar.go を拡張 + 新規 providers/holidayjp/ を作成

**対象ファイル:**
- `calendar.go` に `NextBusinessDay`, `FirstBusinessDayOfMonth`, `FirstBusinessDaysOfYear`, `Holidays` を追記
- `providers/holidayjp/provider.go` を新規作成
- `providers/holidayjp/provider_test.go` を新規作成
- `calendar_test.go` にテストケースを追記

**適合理由:** `calendar.go` に追記する 4 メソッドはすべて同一構造体のメソッドであり、既存の `isExcluded()` や `provider` フィールドに直接アクセスできる。ファイルを分割するほどの責務の分離はない。

**トレードオフ:**
- ✅ `isExcluded()` の再利用が自然（同ファイル・同パッケージ）
- ✅ 既存のテストパターンに完全に合致
- ✅ 新規ファイルは providers/ 以下のみで全体のファイル数を最小化
- ❌ `calendar.go` のファイルサイズが増加（現在 65行 → 推定 130〜150行）

### Option B: calendar_api.go など API 別に新規ファイルを作成

**対象ファイル:**
- `calendar_api.go` に 4 メソッドを独立して配置
- `providers/holidayjp/provider.go` を新規作成

**適合理由:** `calendar.go` の肥大化を防ぎつつ、同パッケージ内で `isExcluded()` へのアクセスを維持できる。

**トレードオフ:**
- ✅ `calendar.go` のサイズを維持
- ✅ 関心の分離がより明確
- ❌ Step 2 の範囲としては過剰な分割（追加メソッドは 4 つのみ）
- ❌ 既存の Step 1 パターンと一貫性がない（Step 1 は calendar.go 1 ファイル）

### Option C: IsBusinessDay を内部ロジックとして再利用

**新規 API 内部から直接 `IsBusinessDay()` を呼び出すことで**、除外日付・祝日・土日の判定ロジックを DRY に保つ選択肢。

**検討事項:** `IsBusinessDay()` が `context.Context` を受け取るため、ループ内で呼び出す際はコンテキスト伝播が自然に行われる。エラーも同様に伝播される。この方式であれば `isExcluded()` を直接呼ぶ代わりに `IsBusinessDay()` を使うだけでよく、実装がシンプルになる。

---

## 4. 推奨アプローチ

**Option A（calendar.go 拡張 + providers/holidayjp/ 新規作成）+ Option C（IsBusinessDay 内部再利用）の組み合わせ**

- `NextBusinessDay`, `FirstBusinessDayOfMonth`, `FirstBusinessDaysOfYear` の内部では `IsBusinessDay()` を呼び出してループする（除外日付・祝日・土日の条件をすべて一元管理）
- `Holidays` は `provider.HolidaysBetween()` に委譲するだけ
- `providers/holidayjp/` は独立パッケージとして新規作成
- `HolidayName` の非祝日エラー変換ロジックは設計フェーズで確定する

---

## 5. 実装の複雑度とリスク

| 分類 | 評価 | 根拠 |
|------|------|------|
| 工数 | S（1〜3日） | Step 1 のパターンが確立されており、4 メソッドはいずれも単純なループロジック + 1 プロバイダーのラッパー |
| リスク | Low | 唯一の注意点は `HolidayName` の API 差異（非祝日エラー変換）だが、対処パターンは明確 |

---

## 6. 設計フェーズへの引継ぎ事項

1. **`HolidayName` 非祝日エラーの変換戦略**: holiday_jp-go の `HolidayName` が非祝日時にエラーを返す仕様に対し、`go-heijitu` の契約（空文字返却）をどう満たすか確定する
2. **`HolidaysBetween` の型変換**: `holiday.Holidays` → `[]heijitu.Holiday` の変換ロジック（日付ゼロ値の扱い含む）
3. **`providers/holidayjp/` のパッケージ命名**: `package holidayjp` とするか `package holiday` とするかを決定する
4. **テストデータの選択**: 特定の祝日日付（元旦、成人の日 等）をハードコードするか、現在年度から動的に計算するかを決定する
