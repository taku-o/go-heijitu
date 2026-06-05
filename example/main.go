package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	heijitu "github.com/taku-o/go-heijitu"
	"github.com/taku-o/go-heijitu/providers/caoCsv"
	"github.com/taku-o/go-heijitu/providers/googleCalendar"
	"github.com/taku-o/go-heijitu/providers/holidayjp"
)

func main() {
	ctx := context.Background()

	runHolidayjp(ctx)
	runExcludedDates(ctx)
	runExtraExcluded(ctx)
	runCaoCsvLocal(ctx)
	runCaoCsvURL(ctx)
	runGoogleCalendar(ctx)
}

func runHolidayjp(ctx context.Context) {
	fmt.Println("=== holidayjp ===")

	provider := holidayjp.New()
	cal := heijitu.New(provider)

	today := time.Now()

	// IsBusinessDay
	isBiz, err := cal.IsBusinessDay(ctx, today)
	if err != nil {
		log.Printf("holidayjp IsBusinessDay error: %v", err)
		return
	}
	fmt.Printf("IsBusinessDay(%s): %v\n", today.Format(time.DateOnly), isBiz)

	// NextBusinessDay
	next, err := cal.NextBusinessDay(ctx, today)
	if err != nil {
		log.Printf("holidayjp NextBusinessDay error: %v", err)
		return
	}
	fmt.Printf("NextBusinessDay(%s): %s\n", today.Format(time.DateOnly), next.Format(time.DateOnly))

	// FirstBusinessDayOfMonth
	first, err := cal.FirstBusinessDayOfMonth(ctx, today.Year(), today.Month())
	if err != nil {
		log.Printf("holidayjp FirstBusinessDayOfMonth error: %v", err)
		return
	}
	fmt.Printf("FirstBusinessDayOfMonth(%d, %s): %s\n", today.Year(), today.Month(), first.Format(time.DateOnly))

	// FirstBusinessDaysOfYear
	days, err := cal.FirstBusinessDaysOfYear(ctx, today.Year())
	if err != nil {
		log.Printf("holidayjp FirstBusinessDaysOfYear error: %v", err)
		return
	}
	fmt.Printf("FirstBusinessDaysOfYear(%d):\n", today.Year())
	for i, d := range days {
		fmt.Printf("  %s: %s\n", time.Month(i+1), d.Format(time.DateOnly))
	}

	// Holidays
	from := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	to := time.Date(today.Year(), 12, 31, 0, 0, 0, 0, time.Local)
	holidays, err := cal.Holidays(ctx, from, to)
	if err != nil {
		log.Printf("holidayjp Holidays error: %v", err)
		return
	}
	fmt.Printf("Holidays(%d):\n", today.Year())
	for _, h := range holidays {
		fmt.Printf("  %s: %s\n", h.Date.Format(time.DateOnly), h.Name)
	}
}

func runExcludedDates(ctx context.Context) {
	fmt.Println()
	fmt.Println("=== WithExcludedDates + WithConfig ===")

	provider := holidayjp.New()

	configOpt, err := heijitu.WithConfig("example/heijitu.yaml")
	if err != nil {
		log.Printf("WithConfig error: %v", err)
		return
	}

	extraDates := []heijitu.MonthDay{
		{Month: time.June, Day: 1},
	}

	cal := heijitu.New(provider, heijitu.WithExcludedDates(extraDates), configOpt)

	// WithExcludedDates で追加した 6/1 を確認
	target := time.Date(time.Now().Year(), time.June, 1, 0, 0, 0, 0, time.Local)
	isBiz, err := cal.IsBusinessDay(ctx, target)
	if err != nil {
		log.Printf("WithExcludedDates check error: %v", err)
		return
	}
	fmt.Printf("WithExcludedDates: IsBusinessDay(%s): %v\n", target.Format(time.DateOnly), isBiz)

	// WithConfig で追加した 12/29 を確認
	target2 := time.Date(time.Now().Year(), time.December, 29, 0, 0, 0, 0, time.Local)
	isBiz2, err := cal.IsBusinessDay(ctx, target2)
	if err != nil {
		log.Printf("WithConfig check error: %v", err)
		return
	}
	fmt.Printf("WithConfig: IsBusinessDay(%s): %v\n", target2.Format(time.DateOnly), isBiz2)
}

func runExtraExcluded(ctx context.Context) {
	fmt.Println()
	fmt.Println("=== extraExcluded ===")

	provider := holidayjp.New()
	cal := heijitu.New(provider)

	today := time.Now()
	extra := heijitu.MonthDay{Month: today.Month(), Day: today.Day()}
	isBiz, err := cal.IsBusinessDay(ctx, today, extra)
	if err != nil {
		log.Printf("extraExcluded error: %v", err)
		return
	}
	fmt.Printf("extraExcluded: IsBusinessDay(%s, extra=%s/%d): %v\n",
		today.Format(time.DateOnly), today.Month(), today.Day(), isBiz)
}

func runCaoCsvLocal(ctx context.Context) {
	fmt.Println()
	fmt.Println("=== caoCsv (local CSV) ===")

	provider, err := caoCsv.New(ctx, caoCsv.Options{
		CSVPath: "example/testdata/syukujitsu.csv",
	})
	if err != nil {
		log.Printf("caoCsv local CSV error: %v", err)
		return
	}

	cal := heijitu.New(provider)

	// 同梱の example/testdata/syukujitsu.csv は2025年分のみを含む最小サンプルのため、
	// 2025年の元日を照会して祝日（営業日でない）と判定されることを示す。
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)
	isBiz, err := cal.IsBusinessDay(ctx, target)
	if err != nil {
		log.Printf("caoCsv IsBusinessDay error: %v", err)
		return
	}
	fmt.Printf("caoCsv local: IsBusinessDay(%s): %v\n", target.Format(time.DateOnly), isBiz)
}

func runCaoCsvURL(ctx context.Context) {
	fmt.Println()
	fmt.Println("=== caoCsv (online URL) ===")

	provider, err := caoCsv.New(ctx, caoCsv.Options{})
	if err != nil {
		log.Printf("caoCsv online URL error: %v", err)
		return
	}

	cal := heijitu.New(provider)

	target := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)
	isBiz, err := cal.IsBusinessDay(ctx, target)
	if err != nil {
		log.Printf("caoCsv online IsBusinessDay error: %v", err)
		return
	}
	fmt.Printf("caoCsv online: IsBusinessDay(%s): %v\n", target.Format(time.DateOnly), isBiz)
}

func runGoogleCalendar(ctx context.Context) {
	fmt.Println()
	fmt.Println("=== googleCalendar ===")

	apiKey := os.Getenv("GOOGLE_CALENDAR_API_KEY")
	if apiKey == "" {
		fmt.Println("GOOGLE_CALENDAR_API_KEY is not set, skipping googleCalendar example")
		return
	}

	provider, err := googleCalendar.New(ctx, googleCalendar.Options{
		APIKey: apiKey,
	})
	if err != nil {
		log.Printf("googleCalendar New error: %v", err)
		return
	}

	cal := heijitu.New(provider)

	target := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)
	isBiz, err := cal.IsBusinessDay(ctx, target)
	if err != nil {
		log.Printf("googleCalendar IsBusinessDay error: %v", err)
		return
	}
	fmt.Printf("googleCalendar: IsBusinessDay(%s): %v\n", target.Format(time.DateOnly), isBiz)
}
