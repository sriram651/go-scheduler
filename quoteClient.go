package main

import (
	"net/http"
)

type QuoteClient struct {
	endpoint string
	client   *http.Client
}
