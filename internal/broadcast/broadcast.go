package broadcast

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/sriram651/go-scheduler/internal/db"
	"github.com/sriram651/go-scheduler/internal/quote"
	"github.com/sriram651/go-scheduler/internal/telegram"
)

type Broadcast struct {
	Quote    *quote.Client
	Telegram *telegram.Client
	Database *sql.DB

	successCount           int
	failureCount           int
	broadcastTrackingMutex sync.Mutex
}

func NewClient(qc *quote.Client, tc *telegram.Client, database *sql.DB) *Broadcast {

	return &Broadcast{
		Quote:    qc,
		Telegram: tc,
		Database: database,
	}
}

func (b *Broadcast) finishBroadcastRun(success bool) {
	b.broadcastTrackingMutex.Lock()

	if success {
		b.successCount++
	} else {
		b.failureCount++
	}

	b.broadcastTrackingMutex.Unlock()
}

func (b *Broadcast) Run(ctx context.Context) {
	cronCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	var broadcastMessage string

	broadcastMessage, quoteFetchErr := b.Quote.GetQuote(cronCtx)

	if quoteFetchErr != nil {
		log.Println(quoteFetchErr)
	}

	if broadcastMessage == "" || quoteFetchErr != nil {
		broadcastMessage = b.Quote.DefaultQuote
	}

	subscribedUsers, getSubscribedUsersErr := db.GetSubscribedUsers(b.Database)

	if getSubscribedUsersErr != nil {
		log.Println(getSubscribedUsersErr)
		return
	}

	for _, user := range subscribedUsers {
		sendMessageError := b.Telegram.HandleSend(cronCtx, user, broadcastMessage, nil)

		if sendMessageError != nil {
			// b.finishBroadcastRun(false)
			log.Println(sendMessageError)
			continue
		}

		// TODO: Re-visit this
		// b.finishBroadcastRun(true)
	}
}
