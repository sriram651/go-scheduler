package telegram

import "net/http"

type Client struct {
	chatId   int64
	endpoint string
	client   *http.Client
}

func NewClient(chatId int64, endpoint string, httpClient *http.Client) *Client {
	return &Client{
		chatId:   chatId,
		endpoint: endpoint,
		client:   httpClient,
	}
}

// Updates from telegram
type Message struct {
	Chat struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
	} `json:"chat"`
	Text string `json:"text"`
}

type UpdateStruct struct {
	UpdateId int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type GetUpdatesResponse struct {
	Ok     bool           `json:"ok"`
	Result []UpdateStruct `json:"result"`
}
