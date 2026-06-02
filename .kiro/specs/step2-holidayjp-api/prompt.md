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

Critical IssueはSuggestionを採用。
  🔴 Critical Issue 1: FirstBusinessDayOfMonth の初期 time.Time 構築のタイムゾーンが未指定
  Suggestion: Design.md の FirstBusinessDayOfMonth Implementation Notes に「候補日は time.Date(year,
  month, day, 0, 0, 0, 0, time.Local) で構築する」と明示する。（既存の IsBusinessDay テストは time.UTC
  使用だが、月初日を構築する際は time.Local が実用上の慣例に合う。どちらかを明示すればよい。）

  🔴 Critical Issue 2: HolidaysBetween 戻り値のソート順検証がテスト戦略に欠如
  Suggestion: テスト戦略 Unit Tests 5番に「かつ日付昇順で並んでいること」の条件を追記する。

/kiro-approve-design step2-holidayjp-api
jj new

-------------

/kiro-spec-tasks step2-holidayjp-api
/kiro-approve-task step2-holidayjp-api
jj new

/kiro-review-spec step2-holidayjp-api

軽微な問題、注意事項、対応推奨アクションにあげられている項目は修正してください。

jj new

!jj-merge feature/step2-holidayjp-api
/commit-commands:commit-push-pr

/review 3

-------------

takt --task "/kiro-impl step2-holidayjp-api 1
必要ならテストを修正して良い。"

/kiro-review-feature step2-holidayjp-api 1
jj new

-------------

takt --task "/kiro-impl step2-holidayjp-api 2
必要ならテストを修正して良い。"

/kiro-review-feature step2-holidayjp-api 2
jj new

-------------

takt --task "/kiro-impl step2-holidayjp-api 3
必要ならテストを修正して良い。"

/kiro-review-feature step2-holidayjp-api 3

設計を承認します。
テストの修正もOK。どのようにするかは判断を任せます。

jj new

-------------

takt --task "/kiro-impl step2-holidayjp-api 4
必要ならテストを修正して良い。"
/kiro-review-feature step2-holidayjp-api 4

/code-review

/simplify-loop

これはどうなった？
> diagnostics の line 243 range over int は私の変更箇所外（既存コード）の可能性が高いため確認します。

こちら対応してください
calendar.go:93（実装コード）の同パターンはスコープ外のため未対応（希望あれば対応可

jj new

-------------

!jj-merge feature/step2-holidayjp-api
/commit-push-pr-update

/review 3

これは対応して、取り込んで。
  - HolidaysBetween の round-trip（provider.go:40-48）: holiday.Between がキー文字列を返し、それを
  ParseInLocation で time.Time に戻している。ライブラリ API の制約で現状はやむを得ないが、将来
  holiday_jp-go が time.Time を直接返す API を提供すれば文字列パースを省ける

取り込まずでOK


