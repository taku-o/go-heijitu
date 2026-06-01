package heijitu

// Option は BusinessCalendar の構築オプション。
type Option func(*BusinessCalendar)

// WithExcludedDates は指定した月日を除外日付として追加するオプションを返す。
// 複数回呼び出した場合、除外日付は追記（マージ）される。
func WithExcludedDates(dates []MonthDay) Option {
	return func(bc *BusinessCalendar) {
		bc.excludedDates = append(bc.excludedDates, dates...)
	}
}

// WithConfig は設定ファイルを読み込み、excluded_dates を除外日付として登録するオプションを返す。
// ファイルが存在しない場合や解析エラーが発生した場合はエラーを返す。
func WithConfig(path string) (Option, error) {
	cfg, err := loadConfig(path)
	if err != nil {
		return nil, err
	}
	return WithExcludedDates(cfg.ExcludedDates), nil
}
