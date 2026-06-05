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

/kiro-spec-tasks step5-docs
/kiro-approve-task step5-docs
jj new

-------------

/kiro-review-spec step5-docs

こちら対応してください。
  1. 【推奨・軽微】tasks.md 2.2 の表現を確定的に:
  「不足があれば補正」→「全公開シンボルを点検し、コメントが無い/Go Doc
  慣習違反のものを補い、結果として全公開シンボルがコメントを持つ状態にする」へ。曖昧表現5件を一括解消。
  2. （任意）design.md に「ルートは単一パッケージ heijitu＝複数ファイル、doc.go
  が概要を担当」の一文を補足。

軽微な問題、注意事項、対応推奨アクションにあげられている項目は修正してください。

jj new
!jj-merge feature/step5-docs
/commit-commands:commit-push-pr

/review 6

推奨の方法にする
>  要件8.4「内容が対応」は構造ミラー（同一章構成）で担保する方針だが、実装時に「対応＝見出し構成一致＋コ
>  ード例セット一致」など客観基準を運用で固定すると検証が容易（実装段階で対応可）。

MIT / taku-o / 2026 で良い
>  - LICENSE 著作権者・年 / CHANGELOG 初版表記: design・research
>  で「実装時に確定」とされている未定値。実装着手時に確定（MIT / taku-o / 2026 想定）。

/commit

tasks.mdの最後にタスクを追加したい。
docs/planningのファイルは、今回の全てのタスクが完了したら、不要ファイルとなりますよね？
.kiro/specs/initial-planning/planning を作成して、そこにファイルを移動するタスクを追加してください。
ファイル移動後、docs/planningディレクトリは削除。

/kiro-review-spec step5-docs

/commit-push

-------------

takt --task "/kiro-impl step5-docs 1
必要ならテストを修正して良い。"

/kiro-review-feature step5-docs 1
jj new

-------------

takt --task "/kiro-impl step5-docs 2.1
必要ならテストを修正して良い。"

/kiro-review-feature step5-docs 2.1
jj new

-------------

/kiro-impl step5-docs 2
/kiro-review-feature step5-docs 2
jj new

-------------

/kiro-impl step5-docs 3
/kiro-review-feature step5-docs 3
jj new

タスク3.1はチェックついてないけど、終わってる？
jj new

-------------

/kiro-impl step5-docs 4
/kiro-review-feature step5-docs 4
jj new

-------------






