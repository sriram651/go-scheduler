package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Limiting default concurrent goroutines to 5
var MAX_CONCURRENT_WORKERS = 5
var currentWorkersCount = 0

var workerMutex sync.Mutex

// Flags
var interval int
var timeForWork int
var maxConcurrentWorkers int

func main() {
	parseFlags()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	// We introduce a interruption channel to observe "Ctrl + C" to initiate completion.
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)

	var workerWaitGroup sync.WaitGroup

Outer:
	for {
		select {
		case <-ticker.C:
			// Lock the access to running while checking
			workerMutex.Lock()

			// Check if the current number of workers being used is under the maximum allowed
			if currentWorkersCount < maxConcurrentWorkers {
				// We still have workers to assign
				currentWorkersCount++
				workerWaitGroup.Add(1)

				go func() {
					doSomething()

					// Since we access/modify this again, we do the lock & unlock again inside goroutine
					workerMutex.Lock()
					currentWorkersCount--
					workerWaitGroup.Done()
					workerMutex.Unlock()
				}()

				// Access to next loop should be granted, so unlocking here
				workerMutex.Unlock()
			} else {
				workerMutex.Unlock()
				fmt.Println(time.Now().Format("15:04:05"), "Skipping")
				continue
			}

		case <-interruptChannel:
			// Stop accepting inputs from interrupt channel also
			signal.Stop(interruptChannel)

			fmt.Print("\nPlease wait till I close existing tasks...\n")

			// Stop accepting new tickers
			ticker.Stop()

			// Wait for the exisiting jobs to finish and exit the scheduler loop
			workerWaitGroup.Wait()
			break Outer
		}
	}
}

func doSomething() {
	fmt.Println(time.Now().Format("15:04:05"), "Doing")

	time.Sleep(time.Duration(timeForWork) * time.Second)
	fmt.Println("Finished")
}

func parseFlags() {
	// Accept --interval flag
	flag.IntVar(&interval, "interval", 5, "The time interval between each schedule in seconds")
	flag.IntVar(&interval, "i", 5, "The time interval between each schedule in seconds")

	flag.IntVar(&timeForWork, "time", 10, "The time (in seconds) taken for each job to reach completion")
	flag.IntVar(&timeForWork, "t", 10, "The time (in seconds) taken for each job to reach completion")

	flag.IntVar(&maxConcurrentWorkers, "workers", MAX_CONCURRENT_WORKERS, "The maximum number of allowed concurrent goroutines")
	flag.IntVar(&maxConcurrentWorkers, "w", MAX_CONCURRENT_WORKERS, "The maximum number of allowed concurrent goroutines")

	flag.Parse()

	if interval < 1 {
		fmt.Println("Interval should atleast be 1s")
		os.Exit(2)
	}

	if maxConcurrentWorkers < 1 || maxConcurrentWorkers > 10 {
		fmt.Println("Workers should be in the range of 1-10")
		os.Exit(2)
	}
}
