// Package heijitu は日本の営業日を計算するライブラリである。
//
// 土日・祝日・会社独自の休業日を考慮して営業日を判定し、次の営業日・
// 指定年月の最初の営業日・指定年の各月初営業日・期間内の祝日一覧を求める
// API を BusinessCalendar が提供する。
//
// 祝日データソースは HolidayProvider インターフェースで抽象化されており、
// 利用者が New に実装を注入して差し替える。標準の実装は providers 配下に
// 用意している（holidayjp: 埋め込みデータ、caoCsv: 内閣府CSV、
// googleCalendar: Google Calendar API）。
//
// 会社独自の休業日は WithExcludedDates（パラメータ指定）と WithConfig
// （設定ファイル指定）で登録でき、IsBusinessDay には呼び出し限定の追加除外
// 日付を渡せる。
package heijitu
