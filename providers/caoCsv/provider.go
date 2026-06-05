// Package caoCsv は内閣府が公開する祝日CSVを祝日データソースとする
// heijitu.HolidayProvider 実装を提供する。ローカルCSVファイルの読み込みと
// オンライン取得の両モードに対応する。
package caoCsv

import (
	"context"
	"slices"
	"time"

	heijitu "github.com/taku-o/go-heijitu"

	syukujitsu "github.com/mikan/syukujitsu-go"
)

// Options は caoCsv プロバイダーのデータソース設定。
type Options struct {
	CSVPath string // ローカルCSVファイルパス。空の場合は内閣府公式データをオンライン取得する
}

// Provider は内閣府CSVを保持する HolidayProvider 実装。
type Provider struct {
	entries []syukujitsu.Entry
}

// New はデータソースを読み込んで caoCsv プロバイダーを返す。
// CSVPath が非空ならローカルファイルを、空なら内閣府公式データをオンライン取得する。
func New(ctx context.Context, opts Options) (*Provider, error) {
	var entries []syukujitsu.Entry
	var err error
	if opts.CSVPath != "" {
		entries, err = syukujitsu.LoadAndParse(opts.CSVPath)
	} else {
		entries, err = syukujitsu.FetchAndParse(ctx)
	}
	if err != nil {
		return nil, err
	}
	return &Provider{entries: entries}, nil
}

// IsHoliday は指定日が祝日かどうかを返す。エラーは常に nil。
func (p *Provider) IsHoliday(_ context.Context, t time.Time) (bool, error) {
	_, found := syukujitsu.Find(p.entries, t)
	return found, nil
}

// HolidayName は指定日の祝日名を返す。非祝日の場合は ("", nil) を返す。
func (p *Provider) HolidayName(_ context.Context, t time.Time) (string, error) {
	name, found := syukujitsu.Find(p.entries, t)
	if !found {
		return "", nil
	}
	return name, nil
}

// HolidaysBetween は from〜to（両端含む）の祝日リストを日付昇順で返す。エラーは常に nil。
func (p *Provider) HolidaysBetween(_ context.Context, from, to time.Time) ([]heijitu.Holiday, error) {
	loc := from.Location()
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc)

	result := make([]heijitu.Holiday, 0)
	for _, e := range p.entries {
		entryDate := time.Date(e.Year, time.Month(e.Month), e.Day, 0, 0, 0, 0, loc)
		if !entryDate.Before(fromDate) && !entryDate.After(toDate) {
			result = append(result, heijitu.Holiday{Date: entryDate, Name: e.Name})
		}
	}
	slices.SortFunc(result, func(a, b heijitu.Holiday) int {
		return a.Date.Compare(b.Date)
	})
	return result, nil
}
