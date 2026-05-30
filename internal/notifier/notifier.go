//sends alert via telegram
//receives signals from detector and sends formatted messages to telegram

package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// telegram send message to a bot api
type Telegram struct {
	token  string
	chatID string
	client *http.Client
}

type telegramRequest struct {
	ChatId    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type telegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

//newtelegram create a new telegram notifier
//get your token from @botfather and chat_id from /getUpdates

func NewTelegram(token, chatID string) *Telegram {
	return &Telegram{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 10 * time.Second},
	}

}

// send posts a message to your telegram chat
// returns an erro if the request fails or telegram rejects it
func (t *Telegram) Send(message string) error {
	if t.token == "" {
		return fmt.Errorf("telegram token is empty")
	}
	if t.chatID == "" {
		return fmt.Errorf("telegram chat ID is empty")
	}
	if message == "" {
		return fmt.Errorf("message is empty")
	}

	payload := telegramRequest{
		ChatId:    t.chatID,
		Text:      message,
		ParseMode: "",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %v", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var tgResp telegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("failed to decode Telegram response: %w", err)
	}

	if !tgResp.OK {
		return fmt.Errorf("Telegram API error: %s", tgResp.Description)
	}

	return nil
}
