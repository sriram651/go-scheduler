package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type SendMessage struct {
	ChatId      int64        `json:"chat_id"`
	Text        string       `json:"text"`
	ReplyMarkup *ReplyMarkup `json:"reply_markup,omitempty"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func (c *Client) Send(ctx context.Context, text string, replyMarkup *ReplyMarkup) error {
	message := SendMessage{
		ChatId:      c.chatId,
		Text:        text,
		ReplyMarkup: replyMarkup,
	}

	messageJson, marshalErr := json.Marshal(message)

	if marshalErr != nil {
		return marshalErr
	}

	sendMessageBody := bytes.NewBuffer(messageJson)

	httpRequest, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, sendMessageBody)

	if requestErr != nil {
		return requestErr
	}

	httpRequest.Header.Set("Content-Type", "application/json")

	response, responseErr := c.client.Do(httpRequest)

	if responseErr != nil {
		return responseErr
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("upstream error sending message %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
			Chat      struct {
				ID        int64  `json:"id"`
				FirstName string `json:"first_name"`
			} `json:"chat"`
			Text string `json:"text"`
		} `json:"result"`
	}

	responseDecoder := json.NewDecoder(response.Body)

	decodeErr := responseDecoder.Decode(&raw)

	if decodeErr != nil {
		return decodeErr
	}

	result := "You sent " + raw.Result.Chat.FirstName + " a message: \"" + raw.Result.Text + "\"\n"

	log.Println(result)

	return nil
}
