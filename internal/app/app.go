package app

import (
	"context"

	"github.com/sriram651/go-scheduler/internal/broadcast"
	"github.com/sriram651/go-scheduler/internal/config"
	"github.com/sriram651/go-scheduler/internal/quote"
	"github.com/sriram651/go-scheduler/internal/scheduler"
	"github.com/sriram651/go-scheduler/internal/telegram"
)

type App struct {
	Telegram  *telegram.Client
	Scheduler *scheduler.Scheduler
	Broadcast *broadcast.Broadcast
}

func New(cfg config.Config) *App {
	telegramClient := telegram.NewClient(cfg.TelegramBaseURL, cfg.TelegramToken, cfg.TelegramPollTimeout)
	schedulerClient := scheduler.New(cfg.Schedule)

	quoteClient := quote.NewClient(cfg.QuotesBaseURL, cfg.DefaultQuote)
	broadcastClient := broadcast.NewClient(cfg.QuotesChatId, quoteClient, telegramClient)

	return &App{
		Telegram:  telegramClient,
		Scheduler: schedulerClient,
		Broadcast: broadcastClient,
	}
}

func (a *App) Start(ctx context.Context) {
	// Start the telegram long polling
	go a.Telegram.StartPolling(ctx)

	// Start scheduler service and send in the broadcast run
	go a.Scheduler.Start(ctx, func() { a.Broadcast.Run(ctx) })

	<-ctx.Done()
}
