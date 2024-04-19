package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/angelorc/go-curve-notifier/internal/types"
	"io"
	"log"
	"net/http"
)

type message struct {
	ChatID                int64  `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
	MessageThreadID       int64  `json:"message_thread_id"`
}

func SendMessage(cfg *types.Config, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.Telegram.BotToken)

	payload, err := json.Marshal(message{
		ChatID:                cfg.Telegram.ChatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		MessageThreadID:       cfg.Telegram.ThreadID,
	})
	if err != nil {
		return err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Println("failed to close response body")
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send successful request. Status was %q", response.Status)
	}
	return nil
}
