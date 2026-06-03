//go:build integration

package caoCsv_test

import (
	"context"
	"testing"
	"time"

	caoCsv "github.com/taku-o/go-heijitu/providers/caoCsv"
)

// TestNew_OnlineMode_FetchesOfficialData は CSVPath 空で内閣府公式データをオンライン取得し、
// 既知の祝日が正しく判定されることを確認する。
func TestNew_OnlineMode_FetchesOfficialData(t *testing.T) {
	// Given: CSVPath 空（オンラインモード）
	ctx := context.Background()
	opts := caoCsv.Options{}

	// When: New を呼ぶ
	p, err := caoCsv.New(ctx, opts)

	// Then: エラーなしでプロバイダーが得られる
	if err != nil {
		t.Fatalf("New(online mode) returned error: %v", err)
	}
	if p == nil {
		t.Fatal("New(online mode) returned nil provider")
	}

	// And: 既知の祝日（元日）が true と判定される
	target := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	got, err := p.IsHoliday(ctx, target)
	if err != nil {
		t.Fatalf("IsHoliday(%s) returned error: %v", target.Format(time.DateOnly), err)
	}
	if !got {
		t.Errorf("IsHoliday(%s) = false, want true", target.Format(time.DateOnly))
	}
}
