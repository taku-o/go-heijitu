package googleCalendar_test

import (
	"context"
	"testing"

	heijitu "github.com/taku-o/go-heijitu"
	googleCalendar "github.com/taku-o/go-heijitu/providers/googleCalendar"
)

// コンパイル時のインターフェース充足チェック。
// このファイルがコンパイルされることで Provider が HolidayProvider を満たすことを保証する。
var _ heijitu.HolidayProvider = (*googleCalendar.Provider)(nil)

// --- New ---

// TestNew_EmptyOptions は APIKey・CredentialsFile 両方空の場合にエラーが返ることを確認する。
// ネットワークアクセスは発生しない（要件 1.4）。
func TestNew_EmptyOptions(t *testing.T) {
	// Given: APIKey と CredentialsFile の両方が空の Options
	ctx := context.Background()
	opts := googleCalendar.Options{}

	// When: New を呼ぶ
	p, err := googleCalendar.New(ctx, opts)

	// Then: エラーが返り、プロバイダーは nil
	if err == nil {
		t.Fatal("New(empty options) returned nil error, want error")
	}
	if p != nil {
		t.Errorf("New(empty options) returned non-nil provider")
	}
}

// TestNew_NonExistentCredentialsFile は存在しないパスを CredentialsFile に指定した場合に
// calendar.NewService のファイル読込失敗としてエラーが返ることを確認する（要件 3.2）。
// ネットワークアクセスは発生しない。
func TestNew_NonExistentCredentialsFile(t *testing.T) {
	// Given: 存在しないパスを CredentialsFile に指定した Options
	ctx := context.Background()
	opts := googleCalendar.Options{CredentialsFile: "/non/existent/credentials.json"}

	// When: New を呼ぶ
	p, err := googleCalendar.New(ctx, opts)

	// Then: エラーが返り、プロバイダーは nil
	if err == nil {
		t.Fatal("New(non-existent credentials file) returned nil error, want error")
	}
	if p != nil {
		t.Errorf("New(non-existent credentials file) returned non-nil provider")
	}
}
