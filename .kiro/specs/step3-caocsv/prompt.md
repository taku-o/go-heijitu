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
ステップ3の要件定義書を作成してください。

step3-caocsv

/kiro-validate-gap step3-caocsv

要件3の任意 CSVURL、ってどんな要件？
  - 重要な制約（要研究）: 設計案で採用予定の mikan/syukujitsu-go の FetchAndParse
  は内閣府URLをハードコードしており、要件3の任意 CSVURL を満たせない。公開関数 Parse([]byte)
  を中核に、取得（ローカル=os.ReadFile / URL=http.Get）を自前で行う構成が要件に整合する（x/text は mikan
  経由の推移的依存になる）。

mikanがCSVのURLに対してハードコードで責任を持ってくれるなら、
mikanを使用するときは、任意CSVURLの仕様は不要じゃないかな？

「任意CSVURLを廃止し mikan に委譲」

何か他に問題見つかったかな？

> ステアリングが実態と乖離している
>
>  2. 【仕様】Requirement 5 AC6 の表現が弱い
>
>  「holidayjp と同等の結果」を requirements.md では "consistent with the official holiday entries
>  contained in that CSV" と書き換えました。これはやや循環的でテスト基準としては曖昧です。実際 holidayjp
>  と内閣府CSVは将来年や振替休日の扱いで完全一致しない可能性があります。AC
>  として残すより「テスト方針（ローカルCSVで holidayjp
>  と突き合わせ）」へ寄せる方が素直かもしれません。承認前に整理するか確認させてください
>
>  3. 【設計時に解消】research.md の申し送り（再掲・未解決）
>
>  - time.Time の Location/正規化（最重要）: 内部キー "2006-01-02" と照合する際、IsHoliday(t) の t
>  が時刻成分やタイムゾーンを持つケースの正規化方針が必要。holidayjp は from.Location()
>  を使用。ここを揃えないと突合ミスが出ます。
>  - mikan の未確認点: LoadAndParse/FetchAndParse がヘッダ行を内部除外するか、go 1.23.4
>  との互換バージョン、Find を使うか自前 map にするか。
>  - オンラインモードのテスト: FetchAndParse はネットワーク依存。通常 go test での扱い（スキップ/分離）を
>  design で決定。

/kiro-steering

推奨の方法で対応
>  2. 【仕様】Requirement 5 AC6 の表現が弱い
>
>  「holidayjp と同等の結果」を requirements.md では "consistent with the official holiday entries
>  contained in that CSV" と書き換えました。これはやや循環的でテスト基準としては曖昧です。実際 holidayjp
>  と内閣府CSVは将来年や振替休日の扱いで完全一致しない可能性があります。AC
>  として残すより「テスト方針（ローカルCSVで holidayjp
>  と突き合わせ）」へ寄せる方が素直かもしれません。承認前に整理するか確認させてください


mikanの未確認点以外は、適切な方法を提案して。
mikanの未確認点は実際に挙動を確認するまで分からないよね？
>  3. 【設計時に解消】research.md の申し送り（再掲・未解決）
>
>  - time.Time の Location/正規化（最重要）: 内部キー "2006-01-02" と照合する際、IsHoliday(t) の t
>  が時刻成分やタイムゾーンを持つケースの正規化方針が必要。holidayjp は from.Location()
>  を使用。ここを揃えないと突合ミスが出ます。
>  - mikan の未確認点: LoadAndParse/FetchAndParse がヘッダ行を内部除外するか、go 1.23.4
>  との互換バージョン、Find を使うか自前 map にするか。
>  - オンラインモードのテスト: FetchAndParse はネットワーク依存。通常 go test での扱い（スキップ/分離）を
>  design で決定。

採用
> 壁時計の暦日（Y/M/D）だけをキーにする方針を提案します。

設計が窮屈になったり、当初の想定と大きく変わらないなら、
> mikan の Find
を採用して。独自実装はなるべくしない。使えるならライブラリを使う。

こちらは判断をお任せする。
> C. オンラインモードのテスト

/kiro-validate-gap step3-caocsv
いろいろ変えたので、もう一度ひととおりチェックして欲しい。

/kiro-approve-req step3-caocsv
jj new


-------------

/kiro-spec-design step3-caocsv
/kiro-validate-design step3-caocsv

Issue 1・2 を design.md に反映する

/kiro-approve-design step3-caocsv
jj new

-------------

/kiro-spec-tasks step3-caocsv

実装時に調査する項目があるよね？
それはどのタイミングでやることになっている？
早い段階のタスクでやりたい。

/kiro-approve-task step3-caocsv
jj new

-------------

/kiro-review-spec step3-caocsv

軽微な問題、注意事項、対応推奨アクションにあげられている項目は修正してください。

jj new

!jj-merge feature/step3-caocsv
/commit-commands:commit-push-pr

/review 4





