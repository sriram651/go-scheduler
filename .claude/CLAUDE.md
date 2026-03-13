# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Build:**
```sh
go build -o go-scheduler ./cmd/scheduler/
```

**Run:**
```sh
./go-scheduler --schedule "0 9 * * *"
# or with interval syntax:
./go-scheduler --schedule "@every 30s"
```

**Run (default schedule — every 6 hours):**
```sh
./go-scheduler
```

**Tests:** No tests currently exist in this codebase.

**Lint:**
```sh
go vet ./...
```

## Environment Setup

Copy `.env.example` to `.env` and fill in values. Required: `TG_BOT_TOKEN`, `QUOTE_API_URL`, `DATABASE_URL`. Optional: `TG_API_BASE_URL` (defaults to Telegram API), `DEFAULT_QUOTE` (fallback text).

## Database Setup

Before running, create these tables in PostgreSQL:

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

The `bot_config` insert is required — without it, the Telegram offset won't persist and messages will replay on restart.

## Architecture

The app is a cron-driven Telegram bot that fetches quotes and broadcasts them to subscribed users.

**Startup flow:** `main.go` loads config → constructs `App` via `internal/app` → starts Telegram polling and cron scheduler as concurrent goroutines → blocks on `SIGINT`/`SIGTERM`.

**Two concurrent subsystems:**
1. **Cron scheduler** (`internal/scheduler`) — wraps `robfig/cron`, fires `broadcast.Run` on schedule. On shutdown, drains in-flight jobs before returning.
2. **Telegram long-polling** (`internal/telegram`) — 65-second poll timeout, routes updates to handlers in `handlers.go`. Persists the update offset to DB after each processed update to prevent replay on restart.

**Scheduled broadcast flow** (`internal/broadcast`): creates a 5-second context → fetches quote via `quote.Client` (falls back to `DEFAULT_QUOTE` on error) → fetches subscribed user chat IDs from DB → sends to each via `telegram.Client`.

**Key design points:**
- All config flows through a single `Config` struct from `internal/config` — no package-level globals.
- `internal/app` is the only place that wires dependencies together.
- `internal/db` uses `database/sql` with `pgx/v5` driver (not pgxpool); the connection is a `*sql.DB` shared across the app.
- Telegram offset is stored in `bot_config` table as a key-value row (`telegram_offset`). It's read at startup and updated after every processed update.
- User upsert on `/start` uses `ON CONFLICT (chat_id) DO UPDATE` — safe to repeat without duplicating users.

## Deployment

Push to `main` triggers the GitHub Actions workflow (`.github/workflows/deploy.yml`), which uses `flyctl deploy --remote-only` to build and deploy the image directly on Fly.io. Secrets are managed via `flyctl secrets set` — no env file on a VPS. The app runs as a background worker on Fly.io (no HTTP service, no health check port).

The `deprecated/` directory contains the old ticker-based scheduler implementation — ignore it.

## How to Work in This Codebase

This project is a learning-oriented codebase. The owner values understanding over feature velocity.

- **Do not provide direct Go code changes.** Do not write or rewrite Go implementation. Exception: `.md` files are fine.
- Move step by step. Understand the current stage before expanding scope.
- Prefer reasoning and architectural guidance over direct code dumps.
- Identify risks and design decisions; ask clarifying questions if needed.
- Suggest structural direction first — guide the user to write the code themselves.
- Keep things incremental. No large refactors unless absolutely necessary.

## Known Pitfalls (Don't Repeat These)

- **`sql.ErrNoRows` is a sentinel error** — always check with `errors.Is(err, sql.ErrNoRows)`, never let it surface as a generic error to callers.
- **Plain `UPDATE` silently no-ops on missing rows** — always upsert (`INSERT ... ON CONFLICT DO UPDATE`) before updating when row existence isn't guaranteed. `/subscribe` and `/unsubscribe` call `AddNewUser` first, then `UpdateSubscription`, for exactly this reason.
- **Per-run counters belong as local variables, not struct fields** — struct fields accumulate across cron runs and produce misleading log summaries.
- **Context cancel: call directly, don't defer** — cancel the context directly after the operation it guards so resources are released promptly, not at function return.
- **Telegram `answerCallbackQuery` must be called before any other response** — stale callback IDs from restarts are harmless noise; just answer and move on.
- **`bot_config` seed row is required** — `UPDATE` with no matching `telegram_offset` row affects 0 rows silently. Symptom: messages replay on every restart.

## Decisions Already Made — Don't Revisit

- **No concurrent broadcast sends yet** — sequential is fine for the current user count, but concurrency must come before retry logic. Retrying inside a sequential loop blocks all subsequent users.
- **No retry logic yet** — blocked on concurrent broadcast sends. A plan is drafted in `retry-plan.md`.
- **Dev environment is local** — `go run ./cmd/scheduler/` with a dev bot token and a dev Neon DB branch. No dev container on the VPS.
- **Docker is prod-only** — single container on Fly.io, secrets managed via `flyctl secrets set`. No multi-environment Docker setup.
- **`bot_config` is a generic key-value table** — kept flexible for future runtime state without over-engineering.
- **Single category selection only** — the new quote API uses AND logic across categories, making multi-select unreliable. Verified against the API. Not a design choice — a hard constraint.
- **No per-user schedule preference** — fixed 1/day hardcoded cron for everyone. Keeps the scheduler model simple.

## Planned Work

See `custom-config-plan.md` for the full design. Work is divided into three phases:

- **Phase 1 — Cache + Deduplication:** Store fetched quotes and stop sending the same quote to the same user twice. Uses existing API with its stable ID as cache key. Start here.
- **Phase 2 — Hashing:** Replace the API's ID with a SHA-256 content hash as the cache key. Validates hashing logic in isolation before the API swap.
- **Phase 3 — New Endpoint + Category Preference:** Swap quote API, add per-user category selection via `/settings`, update broadcast to group users by category and fetch per unique category.

Retry logic is deferred until after concurrent broadcast sends are implemented. See `retry-plan.md`.
