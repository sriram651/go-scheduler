package telegram

import (
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseUrl string
	token   string
	client  *http.Client
	offset  int
}

func NewClient(httpClientTimeout time.Duration) *Client {
	return &Client{
		baseUrl: os.Getenv("TG_API_BASE_URL"),
		token:   os.Getenv("TG_BOT_TOKEN"),
		client:  &http.Client{Timeout: httpClientTimeout},
		offset:  0,
	}
}

func (c *Client) endpoint(path string, params string) string {
	return c.baseUrl + c.token + path + params
}
