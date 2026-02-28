# Go Cron Telegram Reminder Service

A lightweight, cron-driven reminder daemon written in Go that fetches
quotes from an external API and sends them via Telegram on a configurable
schedule. Also listens for interactive Telegram commands via long-polling.

This project evolved from an interval-based scheduler into a
production-shaped, deployable reminder service with proper lifecycle
management.

------------------------------------------------------------------------

## Overview

This service:

-   Runs on a configurable cron schedule
-   Fetches a quote from an external API on each execution
-   Falls back to a configurable default quote on fetch failure
-   Sends messages using the Telegram Bot API
-   Listens for Telegram updates via long-polling
-   Handles `/start` command with inline keyboard buttons
-   Routes callback queries (Subscribe / Unsubscribe)
-   Uses per-execution timeouts via `context`
-   Gracefully shuts down on `SIGINT` / `SIGTERM`
-   Tracks successful and failed executions
-   Waits for in-flight jobs before exiting

It is designed to run as a long-lived background service.

------------------------------------------------------------------------

## Features

-   Cron-based scheduling (via `robfig/cron`)
-   Configurable schedule using flags
-   Centralized config loading via `internal/config`
-   Application orchestrator wiring all services via `internal/app`
-   Encapsulated `telegram.Client` with long-polling, message routing, and
    callback handling
-   Encapsulated `quote.Client` abstraction with fallback support
-   Context-aware HTTP requests with timeout
-   Graceful shutdown with execution draining
-   Mutex-protected success/failure tracking
-   Clean service lifecycle design

------------------------------------------------------------------------

## Project Structure

    go-scheduler/
    ├── cmd/
    │   └── scheduler/
    │       └── main.go          # Entry point — wires config, starts app, handles shutdown
    ├── deprecated/
    │   └── tickerCron.go        # Legacy ticker-based scheduler (unused)
    ├── internal/
    │   ├── app/
    │   │   └── app.go           # Application orchestrator — wires all services
    │   ├── broadcast/
    │   │   └── broadcast.go     # Quote broadcast coordinator with success/failure tracking
    │   ├── config/
    │   │   └── config.go        # Config loader — env vars and CLI flags
    │   ├── quote/
    │   │   ├── client.go        # QuoteClient struct and constructor
    │   │   └── get.go           # GetQuote(ctx) method
    │   ├── scheduler/
    │   │   └── scheduler.go     # Cron scheduler wrapper
    │   └── telegram/
    │       ├── client.go        # TelegramClient struct and constructor
    │       ├── handlers.go      # Message and callback query handlers
    │       ├── polling.go       # Long-polling implementation
    │       └── types.go         # Telegram API type definitions
    ├── .env                     # Local secrets — never committed
    ├── .env.example             # Safe template to commit
    ├── DEPLOY.md                # VPS deployment guide
    ├── go.mod
    └── go.sum

------------------------------------------------------------------------

## Environment Variables

Create a `.env` file (excluded via `.gitignore`):

    TG_BOT_TOKEN=123456:ABCDEF...
    TG_CHAT_ID=123456789
    TG_API_BASE_URL=https://api.telegram.org/bot
    QUOTE_API_URL=https://your-quote-api.com/api/random
    DEFAULT_QUOTE=Keep pushing forward, no matter what challenges you face.

Or export them manually:

    export TG_BOT_TOKEN=...
    export TG_CHAT_ID=...
    export TG_API_BASE_URL=https://api.telegram.org/bot
    export QUOTE_API_URL=...
    export DEFAULT_QUOTE=...

`TG_BOT_TOKEN`, `TG_CHAT_ID`, and `QUOTE_API_URL` are required. The service
exits on startup if any are missing.

------------------------------------------------------------------------

## Build

    go build -o go-scheduler ./cmd/scheduler/

------------------------------------------------------------------------

## Run

    ./go-scheduler --schedule "0 9 * * *"

### Flags

  ------------------------------------------------------------------------
  Flag                 Alias       Default             Description
  -------------------- ----------- ------------------- -------------------
  `--schedule`         `-s`        `0 * * * *`         Cron expression
                                                        defining when the
                                                        reminder runs
  ------------------------------------------------------------------------

------------------------------------------------------------------------

## Cron Syntax

Supports:

-   Standard 5-field cron expressions\
    `0 9 * * *`
-   Interval syntax\
    `@every 30s`, `@every 5m`

Powered by `robfig/cron`. Schedules are evaluated in local time.

------------------------------------------------------------------------

## Runtime Behavior

On startup:

-   Loads environment variables and CLI flags via `internal/config`
-   Initializes Telegram, Quote, Scheduler, and Broadcast clients
-   Starts Telegram long-polling concurrently (65-second poll timeout)
-   Starts cron scheduler concurrently

On each scheduled execution:

-   Creates a 5-second timeout context
-   Fetches a quote from the configured API
-   Falls back to `DEFAULT_QUOTE` if the fetch fails or returns empty
-   Sends the quote via Telegram (formatted with author attribution)
-   Tracks success and failure counts

On Telegram message received:

-   Routes to the appropriate handler
-   `/start` replies with a welcome message and inline keyboard
-   Callback queries (Subscribe / Unsubscribe) are acknowledged and handled

On shutdown (Ctrl + C):

-   Cancels the root context, stopping polling and scheduler
-   Waits for in-progress cron jobs to complete
-   Prints execution summary

Example:

    Cron reminder service shutting down.
    Runs: 8 successful, 0 failed.

------------------------------------------------------------------------

## Architecture

### `internal/config` — `Config`

Loads all configuration from environment variables and CLI flags into a
single `Config` struct passed to all other packages.

### `internal/app` — `App`

Orchestrates service startup. Constructs all clients and runs Telegram
polling and the cron scheduler concurrently.

### `internal/scheduler` — `Scheduler`

Wraps `robfig/cron`. Starts the cron job and blocks until the context is
cancelled, then drains in-flight executions before returning.

### `internal/broadcast` — `Broadcast`

Coordinates a single scheduled execution:

-   Fetches a quote (with timeout)
-   Falls back to `DEFAULT_QUOTE` on any error
-   Sends the message via `telegram.Client`
-   Tracks success/failure counts under a mutex

### `internal/quote` — `quote.Client`

Encapsulates:

-   Quote API endpoint
-   HTTP client
-   `GetQuote(ctx)` method — returns `"quote text\n\n- Author\n"` or an
    error

### `internal/telegram` — `telegram.Client`

Encapsulates:

-   Base URL, token, HTTP client
-   Long-polling via `StartPolling(ctx)` — routes updates to handlers
-   `HandleSend(ctx, chatId, text, replyMarkup)` — sends messages
-   `handleMessage` / `handleCallback` — command and button routing

### Execution Model

-   Cron triggers `broadcast.Run`, each with its own timeout-bound context
-   Telegram polling runs independently in a separate goroutine
-   Root context cancellation stops both subsystems cleanly
-   No global state leakage inside transport layers

------------------------------------------------------------------------

## Deployment

The compiled binary is deployed and running on a Hostinger VPS as a
long-lived background service. Environment variables are configured
directly in the service file on the host.

See [DEPLOY.md](DEPLOY.md) for the full deployment guide.

------------------------------------------------------------------------

## Current Scope

-   Single scheduled reminder
-   Single broadcast target (configured chat ID)
-   Interactive Telegram commands via long-polling (`/start`, callbacks)
-   External quote API with fallback
-   No persistence
-   No retry policy
-   No multi-job configuration

------------------------------------------------------------------------

## Next Steps

-   Persist subscriptions across restarts
-   Retry logic for transient failures
-   Multi-reminder support
-   Dockerization
-   Observability improvements
