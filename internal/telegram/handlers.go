package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sriram651/go-scheduler/internal/db"
)

func (c *Client) handleMessage(ctx context.Context, m *Message) {
	switch m.Text {
	case "/start":
		c.handleStart(ctx, m)
	case "/subscribe":
		if err := c.handleSubscribe(ctx, m, true); err != nil {
			log.Println("error handling subscribe:", err)
		}
	case "/unsubscribe":
		if err := c.handleSubscribe(ctx, m, false); err != nil {
			log.Println("error handling unsubscribe:", err)
		}
	case "/about":
		c.handleAbout(ctx, m)
	case "/timezone":
		if err := c.handleTimezone(ctx, m); err != nil {
			log.Println("error handling timezone:", err)
		}
	}
}

func (c *Client) handleAbout(ctx context.Context, m *Message) {
	about := "✨ *Daemon Bot* ✨\n\n" +
		"Your daily companion for wisdom and inspiration.\n\n" +
		"📖 *What I do*\n" +
		"Once a day, I deliver a handpicked life quote straight to your chat — _no noise, just words worth reading_.\n\n" +
		"⚡ *Commands*\n" +
		"/subscribe — Start receiving quotes\n" +
		"/unsubscribe — Pause anytime, no hard feelings\n" +
		"/timezone — Set your local timezone, _no 3 AM pings_\n" +
		"/about — You're here!\n\n" +
		"Built with ☕ and Go."

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	if err := c.HandleSend(sendCtx, m.Chat.ID, about, nil); err != nil {
		log.Println(err)
	}

	sendCancel()
}

func (c *Client) handleStart(ctx context.Context, m *Message) {
	log.Println("User started -", m.Chat.ID)

	addNewUserErr := db.AddNewUser(ctx, c.Database, db.User{
		ChatId:    m.Chat.ID,
		FirstName: m.Chat.FirstName,
		UserName:  m.Chat.UserName,
	})

	if addNewUserErr != nil {
		log.Println("Error adding new user:", addNewUserErr)
	}

	welcomeMessage := "Hey there!\n\nI am *Daemon Bot*. I send a handpicked life quote once a day. If you'd love that, tap *Subscribe* below — then run /timezone so I can hit the right hour for you."

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

func (c *Client) handleSubscribe(ctx context.Context, m *Message, subscribed bool) error {
	addNewUserErr := db.AddNewUser(ctx, c.Database, db.User{
		ChatId:    m.Chat.ID,
		FirstName: m.Chat.FirstName,
		UserName:  m.Chat.UserName,
	})

	if addNewUserErr != nil {
		log.Println("Error adding new user:", addNewUserErr)

		c.replySubscriptionChangeErr(ctx, subscribed, m.Chat.ID)
		return addNewUserErr
	}

	if err := db.UpdateSubscription(ctx, c.Database, m.Chat.ID, subscribed); err != nil {
		targetState := "subscribed"
		if !subscribed {
			targetState = "unsubscribed"
		}
		log.Println("error updating subscription status to", targetState+":", err)

		c.replySubscriptionChangeErr(ctx, subscribed, m.Chat.ID)
		return err
	}

	c.replySubscription(ctx, subscribed, m.Chat.ID)
	return nil
}

func (c *Client) handleCallback(ctx context.Context, cb *CallbackQuery) {
	c.answerCallback(ctx, cb.ID)

	if cb.Message == nil {
		log.Println("received callback without message, id:", cb.ID)
		return
	}

	isTimezoneContPresent := strings.HasPrefix(cb.Data, "tz-cont:")
	isTimezonePresent := strings.HasPrefix(cb.Data, "tz:")

	if isTimezoneContPresent {
		continent, _ := strings.CutPrefix(cb.Data, "tz-cont:")
		if err := c.handleTimezoneContinent(ctx, continent, cb.Message); err != nil {
			return
		}

		return
	} else if isTimezonePresent {
		timezone, _ := strings.CutPrefix(cb.Data, "tz:")
		if err := c.handleTimezoneSelect(ctx, timezone, cb.Message); err != nil {
			return
		}

		return
	}

	switch cb.Data {
	case "subscribe":
		if err := c.handleSubscribe(ctx, cb.Message, true); err != nil {
			return
		}

	case "unsubscribe":
		if err := c.handleSubscribe(ctx, cb.Message, false); err != nil {
			return
		}
	}
}

func (c *Client) handleTimezone(ctx context.Context, m *Message) error {
	addNewUserErr := db.AddNewUser(ctx, c.Database, db.User{
		ChatId:    m.Chat.ID,
		FirstName: m.Chat.FirstName,
		UserName:  m.Chat.UserName,
	})

	if addNewUserErr != nil {
		log.Println("Error adding new user:", addNewUserErr)

		c.replyTimezoneUpdateErr(ctx, m.Chat.ID)
		return addNewUserErr
	}

	timezoneHandlerMessage := "🌍 *Let's set your timezone*\n\nSo I can send quotes at a sensible hour wherever you are — _no 3 AM pings_.\n\nPick your region to get started:"

	timezonesReplyMarkup := &ReplyMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "Asia", CallbackData: "tz-cont:asia"}, {Text: "Europe", CallbackData: "tz-cont:europe"}},
			{{Text: "Americas", CallbackData: "tz-cont:americas"}, {Text: "Africa", CallbackData: "tz-cont:africa"}},
			{{Text: "Oceania", CallbackData: "tz-cont:oceania"}, {Text: "Other", CallbackData: "tz-cont:other"}},
		},
	}

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, m.Chat.ID, timezoneHandlerMessage, timezonesReplyMarkup)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
		return sendErr
	}

	return nil
}

func (c *Client) handleTimezoneContinent(ctx context.Context, cbData string, m *Message) error {
	continentTitle := strings.ToUpper(cbData[:1]) + cbData[1:]
	timezoneContinentHandlerMessage := "Pick your timezone in *" + continentTitle + "*:"

	var zones []string

	for _, continentGroup := range timezonesByContinent {
		continentName := strings.ToLower(continentGroup.Name)

		if continentName == cbData {
			zones = continentGroup.Zones
			break
		}
	}

	if len(zones) == 0 {
		return fmt.Errorf("Invalid continent selection received - %s", cbData)
	}

	var keyboardMarkup [][]InlineKeyboardButton

	// Used as buffer
	var keyboardRow []InlineKeyboardButton

	for _, zone := range zones {
		if len(keyboardRow) == 2 {
			keyboardMarkup = append(keyboardMarkup, keyboardRow)

			// Empty out the buffer
			keyboardRow = []InlineKeyboardButton{}
		}

		zoneButton := InlineKeyboardButton{
			Text:         zone,
			CallbackData: "tz:" + zone,
		}

		keyboardRow = append(keyboardRow, zoneButton)
	}

	if len(keyboardRow) != 0 {
		keyboardMarkup = append(keyboardMarkup, keyboardRow)

		// Empty out the buffer
		keyboardRow = []InlineKeyboardButton{}
	}

	timezonesReplyMarkup := &ReplyMarkup{
		InlineKeyboard: keyboardMarkup,
	}

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, m.Chat.ID, timezoneContinentHandlerMessage, timezonesReplyMarkup)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
		return sendErr
	}

	return nil
}

func (c *Client) handleTimezoneSelect(ctx context.Context, cbData string, m *Message) error {
	isTimezoneValid := IsValidTimeZone(cbData)

	if !isTimezoneValid {
		return fmt.Errorf("Selected timezone is not valid: %s", cbData)
	}

	addNewUserErr := db.AddNewUser(ctx, c.Database, db.User{
		ChatId:    m.Chat.ID,
		FirstName: m.Chat.FirstName,
		UserName:  m.Chat.UserName,
	})

	if addNewUserErr != nil {
		log.Println("Error adding new user:", addNewUserErr)

		c.replyTimezoneUpdateErr(ctx, m.Chat.ID)
		return addNewUserErr
	}

	tzUpdateErr := db.UpdateUserTimezone(ctx, c.Database, m.Chat.ID, cbData)

	if tzUpdateErr != nil {
		log.Println("Error updating user's timezone. chat_id: ", m.Chat.ID, ", timezone: ", cbData)
		c.replyTimezoneUpdateErr(ctx, m.Chat.ID)

		return tzUpdateErr
	}

	c.replyUpdateTimezone(ctx, cbData, m.Chat.ID)

	return nil
}

func (c *Client) replyUpdateTimezone(ctx context.Context, tz string, chatId int64) {
	loc, _ := time.LoadLocation(tz)
	_, offset := time.Now().In(loc).Zone()

	sendHourOffsetInMinutes := (offset / 60) % 60
	offsetMinutesStr := strconv.Itoa(sendHourOffsetInMinutes)

	if sendHourOffsetInMinutes < 10 {
		offsetMinutesStr = "0" + offsetMinutesStr
	}

	userSendTime := strconv.Itoa(c.sendHour) + ":" + offsetMinutesStr

	answerCallbackText := "✅ *Timezone saved:* `" + tz + "`\n\nNo more 3 AM pings — quotes will land at `" + userSendTime + "hrs` from now on. Run /timezone again to change it."

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, chatId, answerCallbackText, nil)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
	}
}

func (c *Client) replyTimezoneUpdateErr(ctx context.Context, chatId int64) {
	answerCallbackText := "Couldn't save your timezone just now. Please try /timezone again in a moment."

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, chatId, answerCallbackText, nil)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
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
		answerCallbackText = "You're subscribed ✨\n\nYou'll get a life quote once a day. If you haven't already, run /timezone so I send it at a sensible local hour — _no 3 AM pings_."
	} else {
		answerCallbackText = "Unsubscribed. No hard feelings — /subscribe again anytime."
	}

	sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

	sendErr := c.HandleSend(sendCtx, chatId, answerCallbackText, nil)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
	}
}

func (c *Client) replySubscriptionChangeErr(ctx context.Context, subscribed bool, chatId int64) {
	var answerCallbackText string

	if subscribed {
		answerCallbackText = "Couldn't subscribe you right now. Please try again in a moment."
	} else {
		answerCallbackText = "Couldn't unsubscribe you right now. Please try again in a moment."
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
		ParseMode:   "Markdown",
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
