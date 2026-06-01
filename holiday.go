package heijitu

import "time"

// Holiday は祝日の日付と名称を保持する値オブジェクト。
type Holiday struct {
	// Date は祝日の日付。time.Time はタイムゾーン情報を含むため、
	// 呼び出し元で統一した Location（例: time.UTC または JST）を使用すること。
	Date time.Time
	// Name は祝日名（例: "元日"、"成人の日"）。
	Name string
}
