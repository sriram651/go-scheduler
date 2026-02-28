package telegram

import (
	"net/http"
	"time"
)

type Client struct {
	baseUrl string
	token   string
	client  *http.Client
	offset  int
}

func NewClient(baseUrl string, token string, httpClientTimeout time.Duration) *Client {
	return &Client{
		baseUrl: baseUrl,
		token:   token,
		client:  &http.Client{Timeout: httpClientTimeout},
		offset:  0,
	}
}

func (c *Client) endpoint(path string, params string) string {
	return c.baseUrl + c.token + path + params
}
