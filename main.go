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
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func init() {
	godotenv.Load()
}

// Environment variables
var BOT_TOKEN string
var CHAT_ID string
var TELEGRAM_API_BASE_URL string

var sendMessagePath = "/sendMessage"
var endpoint string

// Flag variables
var message string
var schedule string

var successCronCount int
var failureCronCount int

var httpClient = &http.Client{Timeout: 5 * time.Second}
var cronTrackingMutex sync.Mutex

func main() {
	checkAndAssignEnvVars()

	flag.StringVar(&message, "message", "", "The message to be sent by the bot to the user.")
	flag.StringVar(&message, "m", "", "The message to be sent by the bot to the user.")

	flag.StringVar(&schedule, "schedule", "@every 2m", "Cron schedule that controls when the reminder is sent (supports standard cron syntax and @every intervals)")
	flag.StringVar(&schedule, "s", "@every 2m", "Cron schedule that controls when the reminder is sent (supports standard cron syntax and @every intervals)")

	flag.Parse()

	if len(message) < 2 {
		log.Println("Message needs to be atleast 2 characters...")
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

	c.AddFunc(schedule, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		sendMessageError := tc.Send(ctx, message)

		if sendMessageError != nil {
			cronTrackingMutex.Lock()
			failureCronCount++
			cronTrackingMutex.Unlock()
			log.Println(sendMessageError)

			return
		}

		cronTrackingMutex.Lock()
		successCronCount++
		cronTrackingMutex.Unlock()
	})

	log.Println("✅ Starting cron service...")
	c.Start()

	<-interruptChannel

	<-c.Stop().Done()

	fmt.Printf("\nCron reminder service shutting down. \nRuns: %d successful, %d failed.\n", successCronCount, failureCronCount)
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
