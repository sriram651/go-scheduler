package telegram

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (c *Client) StartPolling(ctx context.Context) {
	// Pure polling logic only
	for {
		updates := c.getUpdates(ctx)

		for _, u := range updates {
			c.routeUpdate(u)
			c.offset = u.UpdateId + 1
		}
	}
}

func (c *Client) getUpdates(parentCtx context.Context) []Update {
	params := "?offset=" + strconv.Itoa(c.offset)
	getUpdatesEndpoint := c.endpoint("/getUpdates", params)

	ctx, cancel := context.WithTimeout(parentCtx, 65*time.Second)

	defer cancel()

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, getUpdatesEndpoint, nil)

	if reqErr != nil {
		log.Println(reqErr)
		return nil
	}

	res, resErr := c.client.Do(req)

	if resErr != nil {
		log.Println(resErr)
		return nil
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Println("Received status code:", res.StatusCode, " with body:", body)
		defer res.Body.Close()

		return nil
	}

	responseDecoder := json.NewDecoder(res.Body)

	var updatesRes GetUpdatesResponse

	decodeErr := responseDecoder.Decode(&updatesRes)

	res.Body.Close()

	if decodeErr != nil {
		log.Println(decodeErr)

		return nil
	}

	return updatesRes.Result
}

func (c *Client) routeUpdate(u Update) {
	if u.Message != nil {
		c.handleMessage(u.Message)
	} else if u.CallbackQuery != nil {
		c.handleCallback(u.CallbackQuery)
	}
}
