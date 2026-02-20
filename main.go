package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

// Limiting default concurrent goroutines to 5
var MAX_CONCURRENT_WORKERS = 5
var maxConcurrentWorkers int
var interval int
var running bool
var workerMutex sync.Mutex

func main() {
	// Accept --interval flag
	flag.IntVar(&interval, "interval", 5, "The time interval between each schedule in seconds")
	flag.IntVar(&maxConcurrentWorkers, "workers", MAX_CONCURRENT_WORKERS, "The maximum number of allowed concurrent goroutines")

	flag.Parse()

	if interval <= 1 {
		fmt.Println("Interval should atleast be 2s")
		os.Exit(2)
	}

	if maxConcurrentWorkers < 1 || maxConcurrentWorkers > 10 {
		fmt.Println("Workers should be in the range of 1-10")
		os.Exit(2)
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for range ticker.C {
		// Lock the access to running while checking
		workerMutex.Lock()
		if running {
			workerMutex.Unlock()
			fmt.Println(time.Now().Format("15:04:05"), "Skipping")
			continue
		}

		running = true

		// Access to next loop should be granted, so unlocking here
		workerMutex.Unlock()

		go func() {
			doSomething()

			// Since we access/modify this again, we do the lock & unlock again inside goroutine
			workerMutex.Lock()
			running = false
			workerMutex.Unlock()
		}()
	}
}

func doSomething() {
	fmt.Println(time.Now().Format("15:04:05"), "Doing")

	time.Sleep(time.Duration(3) * time.Second)
	fmt.Println("Finished")
}
