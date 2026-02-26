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

type CallbackQuery struct {
	ID      string   `json:"id"`
	Data    string   `json:"data"`
	Message *Message `json:"message"`
}

type UpdateStruct struct {
	UpdateId      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type GetUpdatesResponse struct {
	Ok     bool           `json:"ok"`
	Result []UpdateStruct `json:"result"`
}
