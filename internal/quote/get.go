package quote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) GetQuote(ctx context.Context) (string, error) {
	var requestBody io.Reader

	httpRequest, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, c.QuoteBaseURL, requestBody)

	if requestErr != nil {
		return "", requestErr
	}

	httpRequest.Header.Set("Content-Type", "application/json")

	response, responseErr := c.Client.Do(httpRequest)

	if responseErr != nil {
		return "", responseErr
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("Error getting quotes %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw struct {
		Id     string `json:"id"`
		Quote  string `json:"text"`
		Author string `json:"byName"`
	}

	responseDecoder := json.NewDecoder(response.Body)

	decodeErr := responseDecoder.Decode(&raw)

	if decodeErr != nil {
		return "", decodeErr
	}

	quoteWithAuthor := raw.Quote + "\n\n" + "- " + raw.Author + "\n"

	return quoteWithAuthor, nil
}
