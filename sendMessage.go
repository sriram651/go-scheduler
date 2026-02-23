package main

import (
	"bytes"
	"encoding/json"
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

func init() {
	godotenv.Load()
	readEnv, readEnvErr := godotenv.Read(".env")

	if readEnvErr != nil {
		fmt.Println(readEnvErr)
		os.Exit(2)
	}

	jsonData, marshalErr := json.Marshal(readEnv)

	if marshalErr != nil {
		fmt.Println(marshalErr)
		os.Exit(2)
	}

	var envVars EnvVars

	unmarshalErr := json.Unmarshal(jsonData, &envVars)

	if unmarshalErr != nil {
		fmt.Println(unmarshalErr)
		os.Exit(2)
	}

	BOT_TOKEN = envVars.BotToken
	CHAT_ID = envVars.ChatId
	TELEGRAM_API_BASE_URL = envVars.TelegramApiBaseUrl
}

type SendMessage struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}

func SendTelegramMessage() {
	endpoint := TELEGRAM_API_BASE_URL + BOT_TOKEN + sendMessagePath

	message := SendMessage{
		ChatId: CHAT_ID,
		Text:   "I am the bot!",
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
