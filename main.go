package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

var interval int
var running bool
var workerMutex sync.Mutex

func main() {
	// Accept --interval flag
	flag.IntVar(&interval, "interval", 5, "The time interval between each schedule in seconds")

	flag.Parse()

	if interval <= 1 {
		fmt.Println("Interval should atleast be 2s")
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
