package telegram

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

var BOT_TOKEN string
var offset = 0

func GetUpdatesHandler() {
	telegramBaseUrl := os.Getenv("TELEGRAM_API_BASE_URL")
	telegramBotToken := os.Getenv("BOT_TOKEN")
	telegramGetUpdatesPath := "/getUpdates"

	httpClient := http.Client{Timeout: 70 * time.Second}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)

		endpoint := telegramBaseUrl + telegramBotToken + telegramGetUpdatesPath + "?timeout=60&offset=" + strconv.Itoa(offset) + ""

		httpRequest, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

		if requestErr != nil {
			log.Println(requestErr)
			continue
		}

		response, err := httpClient.Do(httpRequest)

		cancel()

		if err != nil {
			log.Println(err)
			continue
		}

		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(response.Body)
			log.Printf("Error getting quotes %d: %s", response.StatusCode, strings.TrimSpace(string(body)))

			response.Body.Close()

			continue
		}

		responseDecoder := json.NewDecoder(response.Body)

		var getUpdatesResponse GetUpdatesResponse

		decodeErr := responseDecoder.Decode(&getUpdatesResponse)

		response.Body.Close()

		if decodeErr != nil {
			log.Println(decodeErr)
			continue
		}

		for _, update := range getUpdatesResponse.Result {
			if update.Message != nil {
				message := update.Message.Text
				userId := "tg-user-" + strconv.FormatInt(update.Message.Chat.ID, 10)

				// Subscribing to a new user
				if message == "/start" {
					log.Println("New user subscribed -", userId)
				}
			}

			offset = update.UpdateId + 1
		}
	}
}
