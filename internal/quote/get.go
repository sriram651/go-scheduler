package quote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	httpRequest.Header.Set("x-api-key", c.QuotesApiKey)

	response, responseErr := c.Client.Do(httpRequest)

	if responseErr != nil {
		return "", responseErr
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("Error getting quotes %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	type QuoteStruct struct {
		// Id     string `json:"id"`
		Quote  string `json:"quote"`
		Author string `json:"author"`
	}

	var quoteResponse []QuoteStruct

	responseDecoder := json.NewDecoder(response.Body)

	decodeErr := responseDecoder.Decode(&quoteResponse)

	if decodeErr != nil {
		return "", decodeErr
	}

	if len(quoteResponse) == 0 {
		return "", fmt.Errorf("No quotes returned from the Quotes Ninja API")
	}

	randomQuote := quoteResponse[len(quoteResponse)-1]

	quoteWithAuthor := randomQuote.Quote + "\n\n" + "- " + randomQuote.Author + "\n"

	log.Println(quoteWithAuthor)

	return quoteWithAuthor, nil
}
