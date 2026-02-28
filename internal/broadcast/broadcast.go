package broadcast

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sriram651/go-scheduler/internal/quote"
	"github.com/sriram651/go-scheduler/internal/telegram"
)

type Broadcast struct {
	Quote    *quote.Client
	Telegram *telegram.Client
	chatId   int64

	successCount           int
	failureCount           int
	broadcastTrackingMutex sync.Mutex
}

func NewClient(quotesChatId int64, qc *quote.Client, tc *telegram.Client) *Broadcast {

	return &Broadcast{
		chatId:   quotesChatId,
		Quote:    qc,
		Telegram: tc,
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

	sendMessageError := b.Telegram.HandleSend(cronCtx, b.chatId, broadcastMessage, nil)

	if sendMessageError != nil {
		b.finishBroadcastRun(false)
		log.Println(sendMessageError)

		return
	}

	b.finishBroadcastRun(true)
}
