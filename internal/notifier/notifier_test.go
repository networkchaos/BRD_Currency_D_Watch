package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockTelegram sets up a fake Telegram server and returns a Telegram notifier
// pointed at it. Tests never hit the real Telegram API.
func mockTelegram(t *testing.T, responseOK bool, description string) (*Telegram, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := telegramResponse{
			OK:          responseOK,
			Description: description,
		}
		json.NewEncoder(w).Encode(resp)
	}))

	tg := &Telegram{
		token:  "fake-token",
		chatID: "123456",
		client: srv.Client(),
	}

	// Override the URL by wrapping — we'll refactor Send to accept a URL override
	// For now, test validation paths directly
	return tg, srv
}

func TestSend_EmptyToken(t *testing.T) {
	tg := &Telegram{token: "", chatID: "123", client: http.DefaultClient}
	err := tg.Send("hello")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestSend_EmptyChatID(t *testing.T) {
	tg := &Telegram{token: "tok", chatID: "", client: http.DefaultClient}
	err := tg.Send("hello")
	if err == nil {
		t.Fatal("expected error for empty chatID, got nil")
	}
}

func TestSend_EmptyMessage(t *testing.T) {
	tg := &Telegram{token: "tok", chatID: "123", client: http.DefaultClient}
	err := tg.Send("")
	if err == nil {
		t.Fatal("expected error for empty message, got nil")
	}
}

func TestNewTelegram(t *testing.T) {
	tg := NewTelegram("my-token", "my-chat")
	if tg.token != "my-token" {
		t.Errorf("expected token 'my-token', got '%s'", tg.token)
	}
	if tg.chatID != "my-chat" {
		t.Errorf("expected chatID 'my-chat', got '%s'", tg.chatID)
	}
	if tg.client == nil {
		t.Error("expected non-nil HTTP client")
	}
}
