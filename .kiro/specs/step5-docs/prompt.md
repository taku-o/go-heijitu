/kiro-spec-requirements
docs/planningにプロジェクトの開発計画の資料が置いてあります。
まずその資料を読み込み、
次に、開発計画ステップ5の開発を進めていきます。
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
ステップ5の要件定義書を作成してください。

APIキーの取得方法についても、ドキュメントに記載したい。
当初の計画で記載予定がなければ記載するように計画して。
  1. Google Cloud Console でプロジェクトを作成（または既存を選択）
  2. 「APIとサービス」→「ライブラリ」で Google Calendar API を有効化
  3. 「APIとサービス」→「認証情報」→「認証情報を作成」→「APIキー」
  4. 作成された鍵を、推奨として Calendar API のみに制限（キーの制限 → API制限）
  5. 使うとき: export GOOGLE_CALENDAR_API_KEY=（取得した鍵） → go test -tags integration
  ./providers/googleCalendar/...

step5-docs

-------------

/kiro-validate-gap step5-docs

こちら対応お願いします。
  1. 要件の軽微修正を推奨: 要件2の公開型一覧から Config を外し、「設定ファイル仕様（excluded_dates）」と
  して扱う形に調整（API仕様を実コードに整合させるため）。承認前に反映するのが望ましいです。

jj new

/kiro-approve-req step5-docs

-------------

/kiro-spec-design step5-docs
/kiro-validate-design step5-docs
/kiro-approve-design step5-docs
jj new

-------------

/kiro-spec-tasks step4-googlecalendar




/kiro-approve-task step4-googlecalendar
jj new










