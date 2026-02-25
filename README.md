# Go Cron Telegram Reminder Service

A lightweight, cron-driven reminder daemon written in Go that fetches
quotes from an external API and sends them via Telegram on a configurable
schedule.

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
-   Uses per-execution timeouts via `context`
-   Gracefully shuts down on `SIGINT` / `SIGTERM`
-   Tracks successful and failed executions
-   Waits for in-flight jobs before exiting

It is designed to run as a long-lived background service.

------------------------------------------------------------------------

## Features

-   Cron-based scheduling (via `robfig/cron`)
-   Configurable schedule using flags
-   Encapsulated `telegram.Client` abstraction
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
    │       └── main.go          # Entry point — wiring, flags, cron lifecycle
    ├── internal/
    │   ├── quote/
    │   │   ├── client.go        # QuoteClient struct and constructor
    │   │   └── get.go           # GetQuote(ctx) method
    │   └── telegram/
    │       ├── client.go        # TelegramClient struct and constructor
    │       └── send.go          # Send(ctx, message) method
    ├── .env                     # Local secrets — never committed
    ├── .env.example             # Safe template to commit
    ├── DEPLOY.md                # VPS deployment guide
    ├── go.mod
    └── go.sum

------------------------------------------------------------------------

## Environment Variables

Create a `.env` file (excluded via `.gitignore`):

    BOT_TOKEN=123456:ABCDEF...
    CHAT_ID=123456789
    TELEGRAM_API_BASE_URL=https://api.telegram.org/bot
    QUOTE_API_URL=https://your-quote-api.com/api/random
    DEFAULT_QUOTE=Keep pushing forward, no matter what challenges you face.

Or export them manually:

    export BOT_TOKEN=...
    export CHAT_ID=...
    export TELEGRAM_API_BASE_URL=https://api.telegram.org/bot
    export QUOTE_API_URL=...
    export DEFAULT_QUOTE=...

`BOT_TOKEN`, `CHAT_ID`, and `QUOTE_API_URL` are required. The service
exits on startup if any are missing.

------------------------------------------------------------------------

## Build

    go build -o go-scheduler ./cmd/scheduler/

------------------------------------------------------------------------

## Run

    ./go-scheduler --schedule "@every 1m"

### Flags

  ------------------------------------------------------------------------
  Flag                 Alias       Default             Description
  -------------------- ----------- ------------------- -------------------
  `--schedule`         `-s`        `@every 2m`         Cron expression
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

Powered by `robfig/cron`.

------------------------------------------------------------------------

## Runtime Behavior

On startup:

-   Validates required environment variables
-   Initializes Telegram client
-   Starts cron scheduler

On each scheduled execution:

-   Creates a 5-second timeout context
-   Fetches a quote from the configured API
-   Falls back to `DEFAULT_QUOTE` if the fetch fails or returns empty
-   Sends the quote via Telegram (formatted with author attribution)
-   Tracks success and failure counts

On shutdown (Ctrl + C):

-   Stops accepting new scheduled executions
-   Waits for in-progress jobs to complete
-   Prints execution summary

Example:

    Cron reminder service shutting down.
    Runs: 8 successful, 0 failed.

------------------------------------------------------------------------

## Architecture

### `internal/telegram` — `telegram.Client`

Encapsulates:

-   Endpoint
-   Chat ID
-   HTTP client
-   `Send(ctx, message)` method

### `internal/quote` — `quote.Client`

Encapsulates:

-   Quote API endpoint
-   HTTP client
-   `GetQuote(ctx)` method — returns `"quote text\n\n- Author\n"` or an
    error

### Execution Model

-   Cron triggers execution
-   Each run receives its own timeout-bound context
-   `quote.Client` fetches the quote; falls back to `DEFAULT_QUOTE` on
    any error
-   `telegram.Client` delivers the message
-   No global state leakage inside transport layer
-   Graceful lifecycle handling using `c.Stop().Done()`

------------------------------------------------------------------------

## Deployment

The compiled binary is deployed and running on a Hostinger VPS as a
long-lived background service. Environment variables are configured
directly in the service file on the host.

------------------------------------------------------------------------

## Current Scope (v0.3)

-   Single reminder
-   Single user
-   External quote API with fallback
-   No persistence
-   No retry policy
-   No multi-job configuration

------------------------------------------------------------------------

## Next Steps

-   Retry logic for transient failures
-   Multi-reminder support
-   Dockerization
-   Observability improvements
