package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBaseURL     string
	TelegramToken       string
	TelegramPollTimeout time.Duration
	QuotesBaseURL       string
	DefaultQuote        string
	Schedule            string

	DatabaseURL string
}

// TODO: ENV Vars validation
func LoadConfig() Config {
	godotenv.Load()

	log.Println("✅Env loaded")

	var schedule string

	flag.StringVar(&schedule, "schedule", "0 * * * *", "Cron schedule that controls when the reminder is sent (supports standard cron syntax and @every intervals)")
	flag.StringVar(&schedule, "s", "0 * * * *", "Cron schedule that controls when the reminder is sent (supports standard cron syntax and @every intervals)")

	flag.Parse()

	return Config{
		TelegramToken:       os.Getenv("TG_BOT_TOKEN"),
		TelegramBaseURL:     os.Getenv("TG_API_BASE_URL"),
		TelegramPollTimeout: time.Second * 65,
		QuotesBaseURL:       os.Getenv("QUOTE_API_URL"),
		Schedule:            schedule,
		DefaultQuote:        os.Getenv("DEFAULT_QUOTE"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
	}
}
