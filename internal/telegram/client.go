package telegram

import "net/http"

type Client struct {
	chatId   string
	endpoint string
	client   *http.Client
}

func NewClient(chatId, endpoint string, httpClient *http.Client) *Client {
	return &Client{
		chatId:   chatId,
		endpoint: endpoint,
		client:   httpClient,
	}
}
