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
	"time"

	"github.com/sriram651/go-scheduler/internal/db"
)

func (c *Client) handleMessage(ctx context.Context, m *Message) {
	switch m.Text {
	case "/start":
		c.handleStart(ctx, m)
	}
}

func (c *Client) handleStart(ctx context.Context, m *Message) {
	log.Println("User started -", m.Chat.ID)

	addNewUserErr := db.AddNewUser(c.Database, db.User{
		ChatId:    m.Chat.ID,
		FirstName: m.Chat.FirstName,
		UserName:  m.Chat.UserName,
	})

	if addNewUserErr != nil {
		log.Println("Error adding new user:", addNewUserErr)
	}

	welcomeMessage := "Hey " + m.Chat.FirstName + "!\n\nI am Daemon Bot. I send life quotes every hour. If you would love that, feel free to subscribe to me to get started."

	// The Inline keyboard buttons to subscribe and unsubscribe
	welcomeReplyMarkup := &ReplyMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "Subscribe", CallbackData: "subscribe"},
				{Text: "Unsubscribe", CallbackData: "unsubscribe"},
			},
		},
	}

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, m.Chat.ID, welcomeMessage, welcomeReplyMarkup)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
	}
}

func (c *Client) handleCallback(ctx context.Context, cb *CallbackQuery) {
	c.answerCallback(ctx, cb.ID)

	if cb.Message == nil {
		log.Println("received callback without message, id:", cb.ID)
		return
	}

	switch cb.Data {
	case "subscribe":
		c.replySubscription(ctx, true, cb.Message.Chat.ID)
	case "unsubscribe":
		c.replySubscription(ctx, false, cb.Message.Chat.ID)
	}
}

func (c *Client) answerCallback(ctx context.Context, cbId string) {
	answerCallbackEndpoint := c.endpoint("/answerCallbackQuery", "")

	callbackContext, callbackCancel := context.WithTimeout(ctx, 5*time.Second)

	defer callbackCancel()

	type AnswerCallbackBody struct {
		CallbackQueryID string `json:"callback_query_id"`
	}

	answerCallbackBody := AnswerCallbackBody{
		CallbackQueryID: cbId,
	}

	answerCallbackBodyJson, marshalErr := json.Marshal(answerCallbackBody)

	if marshalErr != nil {
		log.Println(marshalErr)
		return
	}

	requestBody := bytes.NewBuffer(answerCallbackBodyJson)

	answerCallbackReq, answerCallbackReqErr := http.NewRequestWithContext(callbackContext, http.MethodPost, answerCallbackEndpoint, requestBody)

	if answerCallbackReqErr != nil {
		log.Println(answerCallbackReqErr)
		return
	}

	res, answerCallbackResErr := c.client.Do(answerCallbackReq)

	if answerCallbackResErr != nil {
		log.Println(answerCallbackResErr)
		return
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Println("[answerCallback] Callback ID:", cbId)
		log.Println("[answerCallback] Non-200 status code with body:", string(body))
	}

	defer res.Body.Close()
}

func (c *Client) replySubscription(ctx context.Context, subscribed bool, chatId int64) {
	var answerCallbackText string

	if subscribed {
		answerCallbackText = "Thank you for subscribing to my hourly quotes. You will start receiving quotes from the start of next hour(Local time). \n\nI hope you enjoy the journey."
	} else {
		answerCallbackText = "No problem, you can come back to subscribe whenever. \n\nI hope you have a good day!"
	}

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, chatId, answerCallbackText, nil)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
	}
}

func (c *Client) HandleSend(ctx context.Context, chatId int64, text string, replyMarkup *ReplyMarkup) error {
	message := SendMessage{
		ChatID:      chatId,
		Text:        text,
		ReplyMarkup: replyMarkup,
	}

	messageJson, marshalErr := json.Marshal(message)

	if marshalErr != nil {
		return marshalErr
	}

	sendMessageBody := bytes.NewBuffer(messageJson)

	httpRequest, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("/sendMessage", ""), sendMessageBody)

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

	return nil
}
