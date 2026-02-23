# Go Interval Scheduler (v1)

A lifecycle-aware, interval-based job scheduler written in Go.

This project demonstrates bounded concurrency, race-free shared state management, and graceful shutdown handling for long-running services.

---

## Features

* Interval-based execution using `time.Ticker`
* Configurable worker concurrency limit
* Skip-on-saturation policy (no backlog accumulation)
* Graceful shutdown on `SIGINT` / `SIGTERM`
* Waits for in-flight workers before exiting
* Race-free shared state (validated with `-race`)

---

## How It Works

On every interval tick:

* If `currentWorkersCount < maxConcurrentWorkers`

  * Launch a new worker goroutine
* Else

  * Skip execution

On shutdown:

1. Stop accepting new ticks
2. Wait for active workers to finish
3. Exit cleanly

The scheduler never queues jobs. When saturated, it skips.

---

## Usage

Build:

```
go build -o scheduler
```

Run:

```
./scheduler --interval 5 --time 10 --workers 3
```

### Flags

| Flag         | Alias | Description                         | Default |
| ------------ | ----- | ----------------------------------- | ------- |
| `--interval` | `-i`  | Interval between job runs (seconds) | 5       |
| `--time`     | `-t`  | Simulated job duration (seconds)    | 10      |
| `--workers`  | `-w`  | Maximum concurrent workers (1–10)   | 5       |

---

## Example

```
./scheduler -i 5 -t 10 -w 3
```

This runs a job every 5 seconds.
Each job takes 10 seconds.
Maximum 3 concurrent jobs.
Additional ticks are skipped when saturated.

---

## Concurrency Model

* Shared state protected via `sync.Mutex`
* Worker lifecycle coordinated using `sync.WaitGroup`
* No race conditions in shared state access

---

## Shutdown Behavior

* First `Ctrl + C`: graceful shutdown
* Stops scheduling new jobs
* Waits for active workers
* Exits cleanly

---

## Project Focus

This version focuses on:

* Correct scheduling semantics
* Bounded concurrency
* Predictable saturation behavior
* Proper service lifecycle management

It is intentionally simple and does not include:

* Cron support
* Persistent storage
* Context-based cancellation
* Observability features

---

## Status

Stable interval-based scheduler.

Next planned evolution: cron-driven execution and real-world job integration.
