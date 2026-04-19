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
-   Handles `/start`, `/subscribe`, `/unsubscribe`, and `/about` commands
-   Routes callback queries (Subscribe / Unsubscribe)
-   Registers new users in PostgreSQL on `/start` (upsert — safe to repeat)
-   Persists subscription state in the database
-   Broadcasts quotes only to subscribed users
-   Persists Telegram update offset to avoid message replay on restart
-   Uses per-execution timeouts via `context`
-   Gracefully shuts down on `SIGINT` / `SIGTERM`
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
-   PostgreSQL-backed user registration and subscription management
-   Broadcast targets fetched from DB — no hardcoded chat IDs
-   Telegram update offset persisted to DB — no stale message replay on restart
-   Context-aware HTTP requests with timeout
-   Graceful shutdown with execution draining
-   Clean service lifecycle design

------------------------------------------------------------------------

## Project Structure

    go-scheduler/
    ├── cmd/
    │   └── scheduler/
    │       └── main.go               # Entry point — wires config, starts app, handles shutdown
    ├── deprecated/
    │   └── tickerCron.go             # Legacy ticker-based scheduler (unused)
    ├── internal/
    │   ├── app/
    │   │   └── app.go                # Application orchestrator — wires all services
    │   ├── broadcast/
    │   │   └── broadcast.go          # Quote broadcast coordinator with success/failure tracking
    │   ├── config/
    │   │   └── config.go             # Config loader — env vars and CLI flags
    │   ├── db/
    │   │   ├── db.go                 # PostgreSQL connection setup
    │   │   ├── users.go              # User registration and subscription queries
    │   │   └── config.go             # Bot config queries (telegram offset)
    │   ├── quote/
    │   │   ├── client.go             # QuoteClient struct and constructor
    │   │   └── get.go                # GetQuote(ctx) method
    │   ├── scheduler/
    │   │   └── scheduler.go          # Cron scheduler wrapper
    │   └── telegram/
    │       ├── client.go             # TelegramClient struct and constructor
    │       ├── handlers.go           # Message and callback query handlers
    │       ├── polling.go            # Long-polling implementation
    │       └── types.go              # Telegram API type definitions
    ├── .github/
    │   └── workflows/
    │       └── deploy.yml            # GitHub Actions — fly deploy on push to main
    ├── Dockerfile                    # Two-stage build: golang builder → alpine runtime
    ├── fly.toml                      # Fly.io app configuration
    ├── .env.development              # Local dev secrets — never committed
    ├── .env.production               # Production secrets template — never committed
    ├── .env.example                  # Safe template to commit
    ├── .gitignore
    ├── .dockerignore
    ├── DEPLOY.md                     # Fly.io deployment guide
    ├── go.mod
    └── go.sum

------------------------------------------------------------------------

## Environment Variables

Create a `.env` file (excluded via `.gitignore`):

    TG_BOT_TOKEN=123456:ABCDEF...
    TG_API_BASE_URL=https://api.telegram.org/bot
    QUOTE_API_URL=https://your-quote-api.com/api/random
    DEFAULT_QUOTE=Keep pushing forward, no matter what challenges you face.
    DATABASE_URL=postgres://user:password@host:5432/dbname

Or export them manually:

    export TG_BOT_TOKEN=...
    export TG_API_BASE_URL=https://api.telegram.org/bot
    export QUOTE_API_URL=...
    export DEFAULT_QUOTE=...
    export DATABASE_URL=...

`TG_BOT_TOKEN`, `QUOTE_API_URL`, and `DATABASE_URL` are required. The service
exits on startup if any are missing.

------------------------------------------------------------------------

## Database Schema

Before running the service, run the following in your PostgreSQL database:

    CREATE TABLE users (
        chat_id    BIGINT  PRIMARY KEY,
        first_name TEXT    NOT NULL,
        username   TEXT,
        subscribed BOOLEAN NOT NULL DEFAULT false
    );

    CREATE TABLE bot_config (
        key   TEXT PRIMARY KEY,
        value TEXT NOT NULL
    );

    INSERT INTO bot_config (key, value) VALUES ('telegram_offset', '0');

- `chat_id` is the Telegram chat ID — used as the primary key and the conflict target for upserts.
- `username` is nullable — not all Telegram users have a username set.
- `subscribed` defaults to `false` on insert; updated via the Subscribe / Unsubscribe inline buttons.
- `bot_config` stores runtime state that must survive restarts. The `telegram_offset` row tracks the last processed Telegram update ID to prevent message replay.

> **The `INSERT INTO bot_config` line is required.** If the `telegram_offset` row is missing, the service will log a warning at startup and continue running, but offset persistence will be silently broken — `UPDATE` with no matching row affects 0 rows. The symptom: Telegram messages may replay on every restart.

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
-   Connects to PostgreSQL
-   Initializes Telegram, Quote, Scheduler, and Broadcast clients
-   Loads last saved Telegram update offset from DB
-   Starts Telegram long-polling concurrently (65-second poll timeout)
-   Starts cron scheduler concurrently

On each scheduled execution:

-   Creates a 5-second timeout context
-   Fetches a quote from the configured API
-   Falls back to `DEFAULT_QUOTE` if the fetch fails or returns empty
-   Fetches subscribed users from the database
-   Sends the quote to each subscribed user via Telegram
-   Tracks success and failure counts

On Telegram message received:

-   Routes to the appropriate handler
-   `/start` registers the user and replies with a welcome message and inline keyboard
-   `/subscribe` and `/unsubscribe` update subscription state directly — no need to go through `/start`
-   `/about` replies with a description of the bot and available commands
-   Subscribe / Unsubscribe inline button callbacks also update subscription state
-   Each processed update saves the new offset to DB

On shutdown (Ctrl + C):

-   Cancels the root context, stopping polling and scheduler
-   Waits for in-progress cron jobs to complete

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

### `internal/db` — database

Manages PostgreSQL connection and queries:

-   `Connect(dbURL)` — opens and pings the connection
-   `AddNewUser(db, user)` — upserts a user row; updates name fields without touching subscription state
-   `UpdateSubscription(db, chatId, subscribed)` — sets subscribed flag for a user
-   `GetSubscribedUsersForHour(ctx, db, nowUTC, sendHour)` — returns chat IDs of subscribed users whose local hour (per their stored IANA timezone, UTC fallback) matches `sendHour`
-   `GetTelegramOffset(db)` — reads the last saved update offset from `bot_config`
-   `UpdateBotConfig(db, key, value)` — upserts a key-value row in `bot_config`

### `internal/broadcast` — `Broadcast`

Coordinates a single scheduled execution:

-   Fetches a quote (with timeout)
-   Falls back to `DEFAULT_QUOTE` on any error
-   Fetches subscribed users from the database
-   Sends the message to each user via `telegram.Client`

### `internal/quote` — `quote.Client`

Encapsulates:

-   Quote API endpoint
-   HTTP client
-   `GetQuote(ctx)` method — returns `"quote text\n\n- Author\n"` or an
    error

### `internal/telegram` — `telegram.Client`

Encapsulates:

-   Base URL, token, HTTP client, PostgreSQL reference
-   Long-polling via `StartPolling(ctx)` — routes updates to handlers
-   `HandleSend(ctx, chatId, text, replyMarkup)` — sends messages
-   `handleMessage` / `handleCallback` — command and button routing
-   `/start` triggers user upsert; `/subscribe`, `/unsubscribe` update subscription directly; `/about` describes the bot
-   Saves update offset to DB after each processed update

### Execution Model

-   Cron triggers `broadcast.Run`, each with its own timeout-bound context
-   Telegram polling runs independently in a separate goroutine
-   Root context cancellation stops both subsystems cleanly
-   No global state leakage inside transport layers

------------------------------------------------------------------------

## Deployment

The service runs as a Docker container on Fly.io. Every push to `main`
triggers GitHub Actions, which uses `flyctl deploy --remote-only` to
build and deploy the image directly on Fly's infrastructure.

Secrets are managed via `flyctl secrets set` and injected at runtime by
Fly.io. No secrets are committed to this repository.

See [DEPLOY.md](DEPLOY.md) for the full deployment guide.

------------------------------------------------------------------------

## Current Scope

-   Single scheduled reminder
-   PostgreSQL-backed user registration and subscription management
-   Broadcast targets fetched from the database at runtime
-   Telegram update offset persisted — no stale replays on restart
-   Interactive Telegram commands via long-polling (`/start`, `/subscribe`, `/unsubscribe`, `/about`, callbacks)
-   External quote API with fallback
-   No retry policy
-   No multi-job configuration

------------------------------------------------------------------------

## Next Steps

-   Retry logic for transient failures
-   Multi-reminder support
-   Observability improvements
