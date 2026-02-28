package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBaseURL        string
	TelegramToken          string
	TelegramPollTimeout    time.Duration
	TelegramGetPostTimeout time.Duration
	QuotesBaseURL          string
	DefaultQuote           string
	Schedule               string

	QuotesChatId int64
	// DBPath          string
	// Schedule        string
}

func LoadConfig() Config {
	godotenv.Load()

	chatId, err := strconv.ParseInt(os.Getenv("TG_CHAT_ID"), 10, 64)

	if err != nil {
		log.Println(err)
	}

	return Config{
		TelegramToken:          os.Getenv("TG_BOT_TOKEN"),
		TelegramBaseURL:        os.Getenv("TG_API_BASE_URL"),
		TelegramPollTimeout:    time.Second * 65,
		TelegramGetPostTimeout: time.Second * 5,
		QuotesBaseURL:          os.Getenv("QUOTE_API_URL"),
		Schedule:               "0 * * * *",
		DefaultQuote:           os.Getenv("DEFAULT_QUOTE"),
		QuotesChatId:           chatId,
	}
}
