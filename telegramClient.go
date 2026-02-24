package main

import (
	"net/http"
)

type TelegramClient struct {
	chatId   string
	endpoint string
	client   *http.Client
}
