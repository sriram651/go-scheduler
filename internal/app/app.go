package app

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/sriram651/go-scheduler/internal/broadcast"
	"github.com/sriram651/go-scheduler/internal/config"
	"github.com/sriram651/go-scheduler/internal/db"
	"github.com/sriram651/go-scheduler/internal/quote"
	"github.com/sriram651/go-scheduler/internal/scheduler"
	"github.com/sriram651/go-scheduler/internal/telegram"
)

type App struct {
	Telegram  *telegram.Client
	Scheduler *scheduler.Scheduler
	Broadcast *broadcast.Broadcast
	Database  *sql.DB
}

func New(cfg config.Config) *App {
	databaseClient := db.Connect(cfg.DatabaseURL)

	telegramClient := telegram.NewClient(cfg.TelegramBaseURL, cfg.TelegramToken, cfg.TelegramPollTimeout, databaseClient)
	schedulerClient := scheduler.New(cfg.Schedule)

	quoteClient := quote.NewClient(cfg.QuotesBaseURL, cfg.DefaultQuote)
	broadcastClient := broadcast.NewClient(quoteClient, telegramClient, databaseClient)

	return &App{
		Telegram:  telegramClient,
		Scheduler: schedulerClient,
		Broadcast: broadcastClient,
		Database:  databaseClient,
	}
}

func (a *App) Start(ctx context.Context) {
	// Before starting, fetch the stored telegram offset
	telegramOffset, getOffsetErr := db.GetTelegramOffset(ctx, a.Database)
	sendHour, getSendHourErr := db.GetSendHour(ctx, a.Database)

	if getOffsetErr != nil {
		log.Println("Error getting `telegram_offset` from `bot_config` table:", getOffsetErr)
	}

	if getSendHourErr != nil {
		log.Fatalln("Error getting `send_hour` from `bot_config` table:", getSendHourErr)
	}

	// To avoid old messages replays, we store and set the offset if the scheduler restarts for some reason.
	a.Telegram.UpdateOffset(telegramOffset)

	// Update the send_hour retrieved from the DB into the broadcast's client
	a.Broadcast.UpdateSendHour(int(sendHour))
	a.Telegram.UpdateSendHour(int(sendHour))

	// Start the telegram long polling
	go a.Telegram.StartPolling(ctx)

	// Start scheduler service and send in the broadcast run
	go a.Scheduler.Start(ctx, func() { a.Broadcast.Run(ctx, time.Now().UTC()) })

	<-ctx.Done()
}
