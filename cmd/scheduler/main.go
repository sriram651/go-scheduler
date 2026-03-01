package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sriram651/go-scheduler/internal/app"
	"github.com/sriram651/go-scheduler/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newApp := app.New(cfg)

	defer func() {
		log.Println("❌ Closing the postgres pool...")

		err := newApp.Database.Close()

		if err != nil {
			log.Println("❌ Error closing the postgres pool:", err)
		}
	}()

	go newApp.Start(ctx)

	waitForShutdown(cancel)
}

// Handle graceful shutdown
func waitForShutdown(cancel context.CancelFunc) {
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)

	<-interruptChannel

	cancel()
}
