/kiro-spec-requirements
docs/planningにプロジェクトの開発計画の資料が置いてあります。
まずその資料を読み込み、
次に、開発計画ステップ1の開発を進めていきます。
>  ┌──────────┬────────────────────────────────────────────────┐
>  │ ステップ │                      内容                      │
>  ├──────────┼────────────────────────────────────────────────┤
>  │ Step 1   │ プロジェクト初期化・コア型・IsBusinessDay まで │
>  ├──────────┼────────────────────────────────────────────────┤
>  │ Step 2   │ holidayjp プロバイダー + 残り全APIの実装       │
>  ├──────────┼────────────────────────────────────────────────┤
>  │ Step 3   │ 内閣府CSVプロバイダー                          │
>  ├──────────┼────────────────────────────────────────────────┤
>  │ Step 4   │ Google Calendar APIプロバイダー                │
>  ├──────────┼────────────────────────────────────────────────┤
>  │ Step 5   │ example・GoDoc・README（ドキュメント整備）     │
>  └──────────┴────────────────────────────────────────────────┘

作業用のgitブランチを作成後、
ステップ1の要件定義書を作成してください。

-------------

/kiro-validate-gap step1-core

このプロジェクトではエラーが発生したら即座に投げる。
エラーを握りつぶすようなことはしない。フォールバックなども不要。

/kiro-approve-req step1-core

jj new

-------------

/kiro-spec-design step1-core

/kiro-validate-design step1-core

  Suggestion: 依存方向図を以下に修正する：
  Holiday, MonthDay（値オブジェクト）
    ↑ (HolidayProvider, config がそれぞれ依存)
  HolidayProvider（インターフェース） ← calendar.go が直接依存
  config（内部設定型）             ← option.go が依存
    ↑
  option.go
    ↑
  calendar.go（BusinessCalendar）

Suggestion: File Structure Plan に package heijitu の記載を追加する。

/kiro-approve-design step1-core

jj new

-------------

/kiro-spec-tasks step1-core

/kiro-approve-task step1-core

/kiro-review-spec step1-core

見つかった問題、簡単に直せるものは直して。

仕様周り。

2月29日など境界値の扱い、は適当に仕様を決めて。

これも対応。適当に仕様を決める。
> HolidayName が祝日の場合の返り値が未記載。

jj new

jj bookmark set feature/step1-core-implementation -r @-

git switch feature/step1-core-implementation
pull request作成

-------------





