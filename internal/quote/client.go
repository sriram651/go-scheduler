package quote

import "net/http"

type Client struct {
	endpoint string
	client   *http.Client
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		endpoint: endpoint,
		client:   httpClient,
	}
}
