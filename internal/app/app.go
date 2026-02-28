package app

import (
	"context"
	"database/sql"

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

	telegramClient := telegram.NewClient(cfg.TelegramBaseURL, cfg.TelegramToken, cfg.TelegramPollTimeout)
	schedulerClient := scheduler.New(cfg.Schedule)

	quoteClient := quote.NewClient(cfg.QuotesBaseURL, cfg.DefaultQuote)
	broadcastClient := broadcast.NewClient(cfg.QuotesChatId, quoteClient, telegramClient, databaseClient)

	return &App{
		Telegram:  telegramClient,
		Scheduler: schedulerClient,
		Broadcast: broadcastClient,
		Database:  databaseClient,
	}
}

func (a *App) Start(ctx context.Context) {
	// Start the telegram long polling
	go a.Telegram.StartPolling(ctx)

	// Start scheduler service and send in the broadcast run
	go a.Scheduler.Start(ctx, func() { a.Broadcast.Run(ctx) })

	<-ctx.Done()
}
