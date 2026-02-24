# Go Cron Telegram Reminder Service

A lightweight, cron-driven reminder daemon written in Go that sends
scheduled messages via Telegram.

This project evolved from an interval-based scheduler into a
production-shaped, deployable reminder service with proper lifecycle
management.

------------------------------------------------------------------------

## Overview

This service:

-   Runs on a configurable cron schedule
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
-   Encapsulated `TelegramClient` abstraction
-   Context-aware HTTP requests with timeout
-   Graceful shutdown with execution draining
-   Mutex-protected success/failure tracking
-   Clean service lifecycle design

------------------------------------------------------------------------

## Environment Variables

Create a `.env` file (excluded via `.gitignore`):

    BOT_TOKEN=123456:ABCDEF...
    CHAT_ID=123456789
    TELEGRAM_API_BASE_URL=https://api.telegram.org/bot

Or export them manually:

    export BOT_TOKEN=...
    export CHAT_ID=...
    export TELEGRAM_API_BASE_URL=https://api.telegram.org/bot

------------------------------------------------------------------------

## Build

    go build -o go-scheduler

------------------------------------------------------------------------

## Run

    ./go-scheduler --message "Pay rent" --schedule "@every 1m"

### Flags

  ------------------------------------------------------------------------
  Flag                 Alias       Default             Description
  -------------------- ----------- ------------------- -------------------
  `--message`          `-m`        (required, min 2    Message text to
                                  chars)              send

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
-   Sends message via Telegram
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

### TelegramClient

Encapsulates:

-   Endpoint
-   Chat ID
-   HTTP client
-   `Send(ctx, message)` method

### Execution Model

-   Cron triggers execution
-   Each run receives its own timeout-bound context
-   No global state leakage inside transport layer
-   Graceful lifecycle handling using `c.Stop().Done()`

------------------------------------------------------------------------

## Current Scope (v0.2)

-   Single reminder
-   Single user
-   No persistence
-   No retry policy
-   No multi-job configuration

Designed intentionally minimal before deployment.

------------------------------------------------------------------------

## Next Steps

-   VPS deployment
-   Retry logic for transient failures
-   Multi-reminder support
-   Dockerization
-   Observability improvements
