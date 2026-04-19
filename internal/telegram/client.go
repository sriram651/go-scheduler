package telegram

import (
	"database/sql"
	"net/http"
	"time"
)

type Client struct {
	baseUrl  string
	token    string
	client   *http.Client
	offset   int
	sendHour int
	Database *sql.DB
}

func NewClient(baseUrl string, token string, httpClientTimeout time.Duration, database *sql.DB) *Client {
	return &Client{
		baseUrl:  baseUrl,
		token:    token,
		client:   &http.Client{Timeout: httpClientTimeout},
		offset:   0,
		Database: database,
	}
}

func (c *Client) UpdateOffset(newOffset int) {
	c.offset = newOffset
}

func (c *Client) UpdateSendHour(newSendHour int) {
	c.sendHour = newSendHour
}

func (c *Client) endpoint(path string, params string) string {
	return c.baseUrl + c.token + path + params
}
