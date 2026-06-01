package heijitu

import "time"

// MonthDay は年をまたいで有効な月日を表す値オブジェクト。
// 会社独自の休業日など、毎年繰り返す日付の指定に使用する。
// Month と Day のバリデーションは行わない。
// 存在しない組み合わせ（例: 2月30日）はどの time.Time とも一致しない。
type MonthDay struct {
	Month time.Month `yaml:"month" json:"month"`
	Day   int        `yaml:"day" json:"day"`
}

// Matches は t の月と日が MonthDay と一致する場合 true を返す（年は無視）。
func (md MonthDay) Matches(t time.Time) bool {
	return t.Month() == md.Month && t.Day() == md.Day
}
