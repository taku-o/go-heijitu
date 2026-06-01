/kiro-spec-requirements
docs/planningにプロジェクトの開発計画の資料が置いてあります。
まずその資料を読み込み、
次に、開発計画ステップ2の開発を進めていきます。
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
ステップ2の要件定義書を作成してください。

step2-holidayjp-api

/kiro-validate-gap step2-holidayjp-api
/kiro-approve-req step2-holidayjp-api
jj new

これは変換ロジックさえあれば問題ない？
  - 外部ライブラリとの API 差異が 1 つある — holiday_jp-go の HolidayName
  は非祝日時にエラーを返すが、HolidayProvider 契約では空文字を要求する。変換ロジックが必要


-------------

/kiro-spec-design step2-holidayjp-api
/kiro-validate-design step2-holidayjp-api
/kiro-approve-design step2-holidayjp-api

-------------

/kiro-spec-tasks step2-holidayjp-api
/kiro-approve-task step2-holidayjp-api
/kiro-review-spec step2-holidayjp-api




