package quote

import (
	"net/http"
	"time"
)

type Client struct {
	Client       *http.Client
	QuoteBaseURL string
	DefaultQuote string
}

func NewClient(quotesBaseURL string, defaultQuote string) *Client {
	return &Client{
		Client:       &http.Client{Timeout: 5 * time.Second},
		QuoteBaseURL: quotesBaseURL,
		DefaultQuote: defaultQuote,
	}
}
