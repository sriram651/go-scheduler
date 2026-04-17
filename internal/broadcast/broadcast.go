package broadcast

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/sriram651/go-scheduler/internal/db"
	"github.com/sriram651/go-scheduler/internal/quote"
	"github.com/sriram651/go-scheduler/internal/telegram"
)

type Broadcast struct {
	Quote    *quote.Client
	Telegram *telegram.Client
	sendHour int
	Database *sql.DB
}

func NewClient(qc *quote.Client, tc *telegram.Client, database *sql.DB) *Broadcast {

	return &Broadcast{
		Quote:    qc,
		Telegram: tc,
		Database: database,
	}
}

func (b *Broadcast) UpdateSendHour(newSendHour int) {
	b.sendHour = newSendHour
}

func (b *Broadcast) Run(ctx context.Context) {
	log.Println("🚀 Cron run started")

	var success, failure int

	var broadcastMessage string

	quoteCtx, quoteCancel := context.WithTimeout(ctx, 5*time.Second)

	broadcastMessage, quoteFetchErr := b.Quote.GetQuote(quoteCtx)

	if quoteFetchErr != nil {
		log.Println(quoteFetchErr)
	}

	quoteCancel()

	if broadcastMessage == "" || quoteFetchErr != nil {
		broadcastMessage = b.Quote.DefaultQuote
	}

	subscribedUsers, getSubscribedUsersErr := db.GetSubscribedUsers(b.Database)

	if getSubscribedUsersErr != nil {
		log.Println("❌ Cron failed — could not fetch subscribed users:", getSubscribedUsersErr)
		return
	}

	for _, user := range subscribedUsers {
		sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)

		sendMessageError := b.Telegram.HandleSend(sendCtx, user, broadcastMessage, nil)

		sendCancel()

		if sendMessageError != nil {
			log.Printf("⚠️ Send to user %d failed: %s", user, sendMessageError)
			failure++
			continue
		}

		success++
	}

	if failure > 0 {
		if success == 0 {
			log.Println("❌ Cron run failed for all users - Failed:", failure)
		} else {
			log.Printf("⚠️ Partial success — Successful: %d, Failed: %d", success, failure)
		}
	} else {
		log.Printf("✅ Cron successful — %d messages sent", success)
	}
}
