package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var BOT_TOKEN string
var offset = 0
var telegramBaseUrl string
var telegramBotToken string
var telegramSendMessagePath = "/sendMessage"

func init() {
	godotenv.Load()

	telegramBaseUrl = os.Getenv("TELEGRAM_API_BASE_URL")
	telegramBotToken = os.Getenv("BOT_TOKEN")
}

func GetUpdatesHandler() {
	telegramGetUpdatesPath := "/getUpdates"

	telegramSendMessageEndpoint := telegramBaseUrl + telegramBotToken + telegramSendMessagePath

	httpClient := &http.Client{Timeout: 70 * time.Second}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)

		endpoint := telegramBaseUrl + telegramBotToken + telegramGetUpdatesPath + "?timeout=60&offset=" + strconv.Itoa(offset) + ""

		httpRequest, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

		if requestErr != nil {
			log.Println(requestErr)
			continue
		}

		response, err := httpClient.Do(httpRequest)

		cancel()

		if err != nil {
			log.Println(err)
			continue
		}

		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(response.Body)
			log.Printf("Error getting quotes %d: %s", response.StatusCode, strings.TrimSpace(string(body)))

			response.Body.Close()

			continue
		}

		responseDecoder := json.NewDecoder(response.Body)

		var getUpdatesResponse GetUpdatesResponse

		decodeErr := responseDecoder.Decode(&getUpdatesResponse)

		response.Body.Close()

		if decodeErr != nil {
			log.Println(decodeErr)
			continue
		}

		for _, update := range getUpdatesResponse.Result {
			if update.Message != nil {
				message := update.Message.Text
				chatId := strconv.FormatInt(update.Message.Chat.ID, 10)
				userId := "tg-user-" + chatId

				// Subscribing to a new user
				if message == "/start" {
					log.Println("New user subscribed -", userId)

					welcomeMessage := "Hey " + update.Message.Chat.FirstName + "!\n\nI am Daemon Bot. I send life quotes every hour. If you would love that, feel free to subscribe to me to get started."

					// The Inline keyboard buttons to subscribe and unsubscribe
					welcomeReplyMarkup := &ReplyMarkup{
						InlineKeyboard: [][]InlineKeyboardButton{
							{
								{Text: "Subscribe", CallbackData: "subscribe"},
								{Text: "Unsubscribe", CallbackData: "unsubscribe"},
							},
						},
					}

					sendCtx, sendCancel := context.WithTimeout(context.Background(), 5*time.Second)

					// Initiate a Welcome message
					tc := NewClient(update.Message.Chat.ID, telegramSendMessageEndpoint, httpClient)

					sendErr := tc.Send(sendCtx, welcomeMessage, welcomeReplyMarkup)

					sendCancel()

					if sendErr != nil {
						log.Println(sendErr)
					}
				}
			}

			if update.CallbackQuery != nil {
				switch update.CallbackQuery.Data {
				case "subscribe":
					AnswerCallback(update.CallbackQuery.ID, true)
					ReplySubOrUnsub(true, update.CallbackQuery.Message.Chat.ID)
				case "unsubscribe":
					AnswerCallback(update.CallbackQuery.ID, false)
					ReplySubOrUnsub(false, update.CallbackQuery.Message.Chat.ID)
				}
			}

			offset = update.UpdateId + 1
		}
	}
}

func AnswerCallback(callbackId string, subscribed bool) {
	httpClient := http.Client{Timeout: 5 * time.Second}

	callbackContext, callbackCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer callbackCancel()

	telegramAnswerCallbackPath := "/answerCallbackQuery"

	type AnswerCallbackBody struct {
		CallbackQueryId string `json:"callback_query_id"`
		Text            string `json:"text"`
		ShowAlert       bool   `json:"show_alert"`
	}

	answerCallbackBody := AnswerCallbackBody{
		CallbackQueryId: callbackId,
	}

	answerCallbackBodyJson, marshalErr := json.Marshal(answerCallbackBody)

	if marshalErr != nil {
		log.Println(marshalErr)
		return
	}

	requestBody := bytes.NewBuffer(answerCallbackBodyJson)

	telegramAnswerCallbackEndpoint := telegramBaseUrl + telegramBotToken + telegramAnswerCallbackPath

	answerCallbackReq, answerCallbackReqErr := http.NewRequestWithContext(callbackContext, http.MethodPost, telegramAnswerCallbackEndpoint, requestBody)

	if answerCallbackReqErr != nil {
		log.Println(answerCallbackReqErr)
		return
	}

	res, answerCallbackResErr := httpClient.Do(answerCallbackReq)

	if answerCallbackResErr != nil {
		log.Println(answerCallbackResErr)
		return
	}

	defer res.Body.Close()

	log.Println("Callback sent")
}

func ReplySubOrUnsub(subscribed bool, chatId int64) {
	httpClient := &http.Client{Timeout: 5 * time.Second}

	telegramSendMessageEndpoint := telegramBaseUrl + telegramBotToken + telegramSendMessagePath

	var answerCallbackText string

	if subscribed {
		answerCallbackText = "Thank you for subscribing to my hourly quotes. You will start receiving quotes from the start of next hour UTC. \n\nI hope you enjoy the journey."
	} else {
		answerCallbackText = "No problem, you can come back to subscribe whenever. \n\nI hope you have a good day!"
	}

	sendCtx, sendCancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Initiate a Welcome message
	tc := NewClient(chatId, telegramSendMessageEndpoint, httpClient)

	sendErr := tc.Send(sendCtx, answerCallbackText, nil)

	sendCancel()

	if sendErr != nil {
		log.Println(sendErr)
	}
}
