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
	"github.com/sriram651/go-scheduler/internal/quote"
	"github.com/sriram651/go-scheduler/internal/telegram"
)

func init() {
	godotenv.Load()
}

var BOT_TOKEN string
var CHAT_ID string
var TELEGRAM_API_BASE_URL string
var QUOTE_ENDPOINT string
var DEFAULT_QUOTE string

var sendMessagePath = "/sendMessage"
var endpoint string
var quoteUrl string

var schedule string

var successCronCount int
var failureCronCount int

var httpClient = &http.Client{Timeout: 5 * time.Second}
var cronTrackingMutex sync.Mutex

func main() {
	checkAndAssignEnvVars()

	flag.StringVar(&schedule, "schedule", "@every 2m", "Cron schedule that controls when the reminder is sent (supports standard cron syntax and @every intervals)")
	flag.StringVar(&schedule, "s", "@every 2m", "Cron schedule that controls when the reminder is sent (supports standard cron syntax and @every intervals)")

	flag.Parse()

	tc := telegram.NewClient(CHAT_ID, endpoint, httpClient)

	// Start a go-routine to handle the subscriptions to user
	go telegram.GetUpdatesHandler()

	c := cron.New()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)

	c.AddFunc(schedule, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		quoteClient := quote.NewClient(quoteUrl, httpClient)

		var quote string

		quote, quoteFetchErr := quoteClient.GetQuote(ctx)

		if quoteFetchErr != nil {
			quote = DEFAULT_QUOTE
			log.Println(quoteFetchErr)
		} else if quote == "" {
			quote = DEFAULT_QUOTE
		}

		sendMessageError := tc.Send(ctx, quote)

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

	QUOTE_ENDPOINT = os.Getenv("QUOTE_API_URL")

	DEFAULT_QUOTE = os.Getenv("DEFAULT_QUOTE")

	if BOT_TOKEN == "" || CHAT_ID == "" {
		log.Println("Env vars BOT_TOKEN & CHAT_ID are required!")
		os.Exit(2)
	}

	if QUOTE_ENDPOINT == "" {
		log.Println("Env var QUOTE_ENDPOINT is required!")
		os.Exit(2)
	}

	TELEGRAM_API_BASE_URL = os.Getenv("TELEGRAM_API_BASE_URL")

	parsedUrl, urlParseErr := url.Parse(TELEGRAM_API_BASE_URL + BOT_TOKEN + sendMessagePath)

	if urlParseErr != nil {
		fmt.Println("Invalid URL: ", urlParseErr)
		os.Exit(2)
	}

	endpoint = parsedUrl.String()

	parsedQuoteUrl, quoteUrlParseErr := url.Parse(QUOTE_ENDPOINT)

	if quoteUrlParseErr != nil {
		fmt.Println("Invalid URL: ", quoteUrlParseErr)
		os.Exit(2)
	}

	quoteUrl = parsedQuoteUrl.String()
}
