# Deployment Guide

The service runs as a Docker container on Fly.io. Deployments are fully
automated via GitHub Actions — every push to `main` builds and deploys
the image directly to Fly.io using `flyctl`.

No secrets are committed to this repository.

------------------------------------------------------------------------

## How Deployments Work

1. Push to `main` triggers the GitHub Actions workflow
2. Actions sets up `flyctl` using the official `superfly/flyctl-actions` action
3. `flyctl deploy --remote-only` builds the image on Fly's remote builder and deploys it

------------------------------------------------------------------------

## Prerequisites

- A Fly.io account with the app created (`flyctl launch`)
- GitHub Actions secret configured: `FLY_API_TOKEN` (get it via `flyctl auth token`)

------------------------------------------------------------------------

## First-Time Fly.io Setup

### 1. Install flyctl and log in

    brew install flyctl
    flyctl auth login

### 2. Launch the app (first time only)

    flyctl launch

This generates `fly.toml`. Since this is a background worker with no HTTP
server, ensure `fly.toml` has no `[http_service]` block.

### 3. Set up the database schema

Connect to your PostgreSQL instance and run:

```sql
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
```

> **The `INSERT` above is required.** If the `telegram_offset` row is missing, the service will log a warning at startup and continue running, but offset persistence will be silently broken — messages may replay on every restart.

### 4. Set secrets on Fly.io

    flyctl secrets set TG_BOT_TOKEN=your_tg_bot_token_here
    flyctl secrets set TG_API_BASE_URL=https://api.telegram.org/bot
    flyctl secrets set QUOTE_API_URL=https://your-quote-api.com/api/random
    flyctl secrets set DEFAULT_QUOTE="Your fallback quote here."
    flyctl secrets set DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=require"

> Secrets live only in Fly.io and are never committed to git.

### 5. Deploy

    flyctl deploy

------------------------------------------------------------------------

## Checking Logs

    flyctl logs

------------------------------------------------------------------------

## Updating Secrets

    flyctl secrets set KEY=new_value

Fly automatically redeploys after secrets are updated.

------------------------------------------------------------------------

## GitHub Actions Secrets Required

| Secret          | Description                                      |
| --------------- | ------------------------------------------------ |
| `FLY_API_TOKEN` | Fly.io API token — get it via `flyctl auth token` |

------------------------------------------------------------------------

## Summary

| Action          | Command                           |
| --------------- | --------------------------------- |
| Deploy          | Push to `main` — Actions handles the rest |
| Check logs      | `flyctl logs`                     |
| Update secrets  | `flyctl secrets set KEY=value`    |
| Check status    | `flyctl status`                   |
| SSH into machine| `flyctl ssh console`              |
