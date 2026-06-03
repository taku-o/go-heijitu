package googleCalendar

import (
	"context"
	"errors"
	"time"

	heijitu "github.com/taku-o/go-heijitu"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// holidayCalendarID は日本の祝日カレンダーの固定 Calendar ID。
const holidayCalendarID = "ja.japanese.official#holiday@group.v.calendar.google.com"

// Options は googleCalendar プロバイダーの認証設定。
type Options struct {
	APIKey          string // APIキー認証
	CredentialsFile string // OAuth2 サービスアカウントJSONファイルのパス
}

// Provider は Google Calendar API を保持する HolidayProvider 実装。
type Provider struct {
	service *calendar.Service
}

// New は認証方式を選択して Calendar クライアントを構築し googleCalendar プロバイダーを返す。
// CredentialsFile 非空ならサービスアカウント、空かつ APIKey 非空なら APIキー、両方空ならエラー。
func New(ctx context.Context, opts Options) (*Provider, error) {
	var clientOpts []option.ClientOption

	switch {
	case opts.CredentialsFile != "":
		clientOpts = []option.ClientOption{
			option.WithAuthCredentialsFile(option.ServiceAccount, opts.CredentialsFile),
			option.WithScopes(calendar.CalendarReadonlyScope),
		}
	case opts.APIKey != "":
		clientOpts = []option.ClientOption{
			option.WithAPIKey(opts.APIKey),
		}
	default:
		return nil, errors.New("googleCalendar: either APIKey or CredentialsFile must be provided")
	}

	svc, err := calendar.NewService(ctx, clientOpts...)
	if err != nil {
		return nil, err
	}
	return &Provider{service: svc}, nil
}

// dayStart は t の暦日（Y/M/D）の 0 時を loc で返す。
func dayStart(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

// holidaysInWindow は timeMin〜timeMax（UTC・RFC3339）の範囲で Events.List を全ページ取得し、
// 終日イベントを heijitu.Holiday に変換して返す。
func (p *Provider) holidaysInWindow(ctx context.Context, timeMin, timeMax time.Time, loc *time.Location) ([]heijitu.Holiday, error) {
	call := p.service.Events.List(holidayCalendarID).
		Context(ctx).
		TimeMin(timeMin.UTC().Format(time.RFC3339)).
		TimeMax(timeMax.UTC().Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime")

	var holidays []heijitu.Holiday
	pageToken := ""
	for {
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		events, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, ev := range events.Items {
			if ev.Start == nil || ev.Start.Date == "" {
				continue
			}
			d, err := time.ParseInLocation(time.DateOnly, ev.Start.Date, loc)
			if err != nil {
				return nil, err
			}
			holidays = append(holidays, heijitu.Holiday{Date: d, Name: ev.Summary})
		}
		if events.NextPageToken == "" {
			break
		}
		pageToken = events.NextPageToken
	}
	return holidays, nil
}

// IsHoliday は指定日が祝日かどうかを返す。
func (p *Provider) IsHoliday(ctx context.Context, t time.Time) (bool, error) {
	holidays, err := p.HolidaysBetween(ctx, t, t)
	if err != nil {
		return false, err
	}
	return len(holidays) > 0, nil
}

// HolidayName は指定日の祝日名を返す。非祝日の場合は ("", nil) を返す。
func (p *Provider) HolidayName(ctx context.Context, t time.Time) (string, error) {
	holidays, err := p.HolidaysBetween(ctx, t, t)
	if err != nil {
		return "", err
	}
	if len(holidays) == 0 {
		return "", nil
	}
	return holidays[0].Name, nil
}

// HolidaysBetween は from〜to（両端含む）の祝日リストを日付昇順で返す。
func (p *Provider) HolidaysBetween(ctx context.Context, from, to time.Time) ([]heijitu.Holiday, error) {
	loc := from.Location()
	fromDate := dayStart(from, loc)
	toDate := dayStart(to, loc)

	if fromDate.After(toDate) {
		return []heijitu.Holiday{}, nil
	}

	// クエリ窓は UTC・RFC3339 で前後に余裕を取り、取得後に loc 壁時計 Y/M/D で厳密照合する。
	timeMin := dayStart(from, time.UTC).AddDate(0, 0, -1)
	timeMax := dayStart(to, time.UTC).AddDate(0, 0, 2)
	holidays, err := p.holidaysInWindow(ctx, timeMin, timeMax, loc)
	if err != nil {
		return nil, err
	}

	// Events.List は SingleEvents(true)+OrderBy("startTime") で日付昇順に返し、全ページを順次連結するため、
	// 範囲フィルタ後も昇順が保たれる（明示ソートは不要）。
	result := make([]heijitu.Holiday, 0, len(holidays))
	for _, h := range holidays {
		if !h.Date.Before(fromDate) && !h.Date.After(toDate) {
			result = append(result, h)
		}
	}
	return result, nil
}
