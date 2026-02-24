package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type EnvVars struct {
	BotToken           string `json:"BOT_TOKEN"`
	ChatId             string `json:"CHAT_ID"`
	TelegramApiBaseUrl string `json:"TELEGRAM_API_BASE_URL"`
}

var BOT_TOKEN string
var CHAT_ID string
var TELEGRAM_API_BASE_URL string
var sendMessagePath = "/sendMessage"

var message string

func init() {
	godotenv.Load()

	BOT_TOKEN = os.Getenv("BOT_TOKEN")
	CHAT_ID = os.Getenv("CHAT_ID")
	TELEGRAM_API_BASE_URL = os.Getenv("TELEGRAM_API_BASE_URL")
}

type SendMessage struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}

func SendTelegramMessage() {
	flag.StringVar(&message, "message", "", "The message to be sent by the bot to the user.")
	flag.StringVar(&message, "m", "", "The message to be sent by the bot to the user.")

	flag.Parse()

	if len(message) < 2 {
		fmt.Println("Message needs to be atleast 2 characters...")
		os.Exit(2)
	}

	endpoint := TELEGRAM_API_BASE_URL + BOT_TOKEN + sendMessagePath

	message := SendMessage{
		ChatId: CHAT_ID,
		Text:   message,
	}

	messageJson, marshalErr := json.Marshal(message)

	if marshalErr != nil {
		fmt.Println(marshalErr)
		os.Exit(2)
	}

	sendMessageBody := bytes.NewBuffer(messageJson)

	res, err := http.Post(endpoint, "application/json", sendMessageBody)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	body, readBodyErr := io.ReadAll(res.Body)
	defer res.Body.Close()

	if readBodyErr != nil {
		fmt.Println(readBodyErr)
		os.Exit(2)
	}

	fmt.Println("Here is the response:", string(body))
}
