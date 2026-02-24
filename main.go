package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func init() {
	godotenv.Load()
}

var BOT_TOKEN string
var CHAT_ID string
var TELEGRAM_API_BASE_URL string
var sendMessagePath = "/sendMessage"
var endpoint string

var message string

var httpClient = &http.Client{Timeout: 5 * time.Second}

func main() {
	checkAndAssignEnvVars()

	flag.StringVar(&message, "message", "", "The message to be sent by the bot to the user.")
	flag.StringVar(&message, "m", "", "The message to be sent by the bot to the user.")

	flag.Parse()

	if len(message) < 2 {
		fmt.Println("Message needs to be atleast 2 characters...")
		os.Exit(2)
	}

	tc := &TelegramClient{
		chatId:   CHAT_ID,
		endpoint: endpoint,
		client:   httpClient,
	}

	c := cron.New()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)

	c.AddFunc("@every 2m", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		sendMessageError := tc.Send(ctx, message)

		if sendMessageError != nil {
			log.Println(sendMessageError)
		}
	})

	c.Start()

	notified := <-interruptChannel

	fmt.Print("\nStopping cron scheduling with ", notified, " signal...\n")

	<-c.Stop().Done()
}

func checkAndAssignEnvVars() {
	BOT_TOKEN = os.Getenv("BOT_TOKEN")
	CHAT_ID = os.Getenv("CHAT_ID")

	if BOT_TOKEN == "" || CHAT_ID == "" {
		log.Println("Env vars BOT_TOKEN & CHAT_ID are required!")
		os.Exit(2)
	}
	TELEGRAM_API_BASE_URL = os.Getenv("TELEGRAM_API_BASE_URL")

	parsedUrl, urlParseErr := url.Parse(TELEGRAM_API_BASE_URL + BOT_TOKEN + sendMessagePath)

	if urlParseErr != nil {
		fmt.Println("Invalid URL: ", urlParseErr)
		os.Exit(2)
	}

	endpoint = parsedUrl.String()
}
