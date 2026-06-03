/kiro-spec-requirements
docs/planningにプロジェクトの開発計画の資料が置いてあります。
まずその資料を読み込み、
次に、開発計画ステップ4の開発を進めていきます。
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
ステップ4の要件定義書を作成してください。

step4-googlecalendar

-------------

/kiro-validate-gap step4-googlecalendar
いくつかの選択がある場合は、推奨の方法を選択。

/kiro-approve-req step4-googlecalendar
jj new

-------------

/kiro-spec-design step4-googlecalendar
/kiro-validate-design step4-googlecalendar
/kiro-approve-design step4-googlecalendar
jj new

-------------

/kiro-spec-tasks step4-googlecalendar
/kiro-approve-task step4-googlecalendar
jj new

-------------

/kiro-review-spec step4-googlecalendar

軽微な問題、注意事項、対応推奨アクションにあげられている項目は修正してください。

この違いはなに？
WithCredentialsFile か WithAuthCredentialsFile

WithAuthCredentialsFileを採用

軽微な問題、注意事項、対応推奨アクションにあげられている項目は修正してください。

jj new
!jj-merge feature/step4-googlecalendar
/commit-commands:commit-push-pr

/review 5

-------------

APIキーや認証は必要？
どのタイミングで必要になる？
どうやって取得すれば良い？

  1. Google Cloud Console でプロジェクトを作成（または既存を選択）
  2. 「APIとサービス」→「ライブラリ」で Google Calendar API を有効化
  3. 「APIとサービス」→「認証情報」→「認証情報を作成」→「APIキー」
  4. 作成された鍵を、推奨として Calendar API のみに制限（キーの制限 → API制限）
  5. 使うとき: export GOOGLE_CALENDAR_API_KEY=（取得した鍵） → go test -tags integration
  ./providers/googleCalendar/...

-------------



