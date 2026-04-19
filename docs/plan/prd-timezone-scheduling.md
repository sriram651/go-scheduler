# PRD: Timezone-Aware Scheduling

## Problem Statement

Users in non-UTC timezones receive quotes at inconvenient local hours. The bot currently fires on a fixed UTC schedule (every 6 hours), meaning a user in Singapore receives quotes at 1AM, 7AM, 1PM, and 7PM local time with no control over timing. There is no way for a user to configure when they receive their daily quote relative to their local time.

## Solution

Allow each user to set their timezone once via a `/timezone` command. A single globally-configured send hour (e.g. 9AM) is stored in the bot config. The cron fires every hour, and on each fire, only users whose local time matches the configured send hour receive a quote. Users who have not set a timezone fall back to UTC.

## User Stories

1. As a subscribed user in a non-UTC timezone, I want to receive my quote at a reasonable local hour, so that it does not arrive at 1AM while I am asleep.
2. As a subscribed user, I want to set my timezone through the Telegram bot, so that the bot knows when to send me quotes.
3. As a subscribed user, I want to choose my timezone from a list of options presented as buttons, so that I do not have to type a timezone name manually.
4. As a subscribed user, I want my timezone preference to be saved permanently, so that I do not have to set it again after every restart.
5. As a subscribed user who has not set a timezone, I want to continue receiving quotes on the UTC schedule, so that the bot still works for me without any action required.
6. As a subscribed user, I want to be able to change my timezone at any time, so that I can update it if I move or travel.
7. As a subscribed user, I want confirmation after setting my timezone, so that I know the change was saved successfully.
8. As the bot owner, I want to configure the global send hour once in the bot config, so that all users receive their quote at that hour in their respective local times.
9. As the bot owner, I want the send hour to be loaded at startup, so that it does not require a DB read on every cron fire.
10. As the bot owner, I want the cron to fire every hour instead of every 6 hours, so that users across all timezones are served at the correct local hour.
11. As the bot owner, I want users who have not set a timezone to fall back to UTC automatically, so that existing users are unaffected by this change.
12. As the bot owner, I want the list of supported timezones to be hardcoded in the application, so that no additional infrastructure is needed to manage them.
13. As the bot owner, I want timezone names stored as IANA names (e.g. `Asia/Singapore`), so that daylight saving time is handled correctly at runtime without manual offset updates.
14. As the bot owner, I want the user filtering logic to run in SQL, so that only relevant users are fetched per cron fire and no unnecessary data is transferred.
15. As the bot owner, I want the current UTC time to be passed into the broadcast function from the scheduler, so that the filtering logic is testable in isolation.

## Implementation Decisions

### DB Schema Changes
- Add a nullable `timezone TEXT` column to the `users` table. Null means UTC fallback — no migration needed for existing users.
- Add a `send_hour` key to the `bot_config` table with the desired send hour as a string integer (e.g. `"9"` for 9AM). Loaded once at startup.

### Config / `sendHour` placement
- `send_hour` is loaded from `bot_config` at startup and lives on the consuming component.
- **Amended during implementation:** originally specified as a new `Config.SendHour int` field. That conflicts with `Config`'s "loaded from env + flags at startup, read-only afterwards" invariant (the DB isn't up when `LoadConfig` runs). Resolved by placing `sendHour` on `Broadcast` as a private field, set via `Broadcast.UpdateSendHour(int)` from `app.Start` after `GetSendHour`. Matches the real `telegram_offset` pattern (offset lives on the Telegram client, not on `Config`), and keeps the value on the component that actually uses it.

### Cron Schedule
- Default schedule expression changes from `0 */6 * * *` to `0 * * * *` (every hour).

### New DB Function — User Filtering
- A new DB function replaces `GetSubscribedUsers` for the broadcast path.
- Accepts the current UTC time and the configured send hour as parameters (not computed internally — required for testability).
- Uses PostgreSQL `AT TIME ZONE` and `EXTRACT(HOUR FROM ...)` to convert each user's stored IANA timezone to local time and compare the local hour against the send hour.
- Users with a null timezone are treated as UTC.
- Returns only the chat IDs of matching subscribed users.

### Broadcast
- `broadcast.Run` receives the current UTC time from the scheduler (passed in, not computed internally).
- Uses the new DB filtering function instead of `GetSubscribedUsers`.
- All other broadcast logic (quote fetch, send loop, counters) is unchanged.

### Scheduler
- Captures `time.Now().UTC()` at the moment the cron fires and passes it into `broadcast.Run`.

### `/timezone` Command
- New command added to the message handler alongside `/start`, `/subscribe`, `/unsubscribe`, `/about`.
- Responds with an inline keyboard built from a hardcoded slice of supported timezones.
- Each button displays a human-readable label (e.g. `India (IST)`) and carries the IANA name as its callback data (e.g. `Asia/Kolkata`).
- The callback handler reads the IANA name from the callback data, updates `users.timezone` for that user, and replies with a confirmation message.

### Supported Timezones
- ~15–20 entries covering major populated regions worldwide.
- Hardcoded as a slice of structs (IANA name + display name) in the application code.
- Future option: migrate to a `supported_timezones` DB table to allow additions without deployment. Not needed at current scale.
- IANA names sourced from the standard tz database (Wikipedia "List of tz database time zones").
- **Amended during implementation:** shipped 67 zones grouped into 6 continents (`[]continentGroup{Name, Zones}` in `internal/telegram/timezones.go`), not a flat 15–20 list. A flat keyboard of 67 buttons is unusable; the two-level continent → zone picker handles the expanded list cleanly. See `timezone-command-session.md` for the curated zone breakdown.

### Why IANA names over UTC offsets?
Many timezones shift their UTC offset seasonally for daylight saving time. Storing a fixed offset (e.g. `-05:00`) would silently break for affected users when clocks change. Storing the IANA name (e.g. `America/New_York`) allows the runtime to compute the correct current offset on every cron fire. India (`Asia/Kolkata`, `+05:30`) does not observe DST — the distinction only matters for affected regions, but using IANA names consistently avoids the problem entirely.

### Why SQL filtering over Go filtering?
Fetching all subscribed users every hour and discarding non-matching ones in Go would transfer wasted rows on every cron fire. PostgreSQL natively supports timezone-aware hour extraction. Filtering in SQL means only the relevant users are returned, keeping the broadcast path efficient as user count grows.

## Testing Decisions

No tests currently exist in this codebase. The following module is the strongest candidate for introducing the first test:

**New DB filtering function** — this is the deepest module introduced by this feature. It takes a UTC time and a send hour as inputs and returns a list of chat IDs. Given a known set of users with known timezones, the output is fully deterministic. A good test would:
- Insert users with known timezones into a test DB
- Call the function with a specific UTC time
- Assert that only the users whose local hour matches are returned
- Test the null timezone fallback (UTC) explicitly

A good test checks external behavior only — what goes in, what comes out — not implementation details like query structure or intermediate variables.

## Out of Scope

- Per-user send hour preference — all users share a single globally-configured send hour.
- Quote cache and deduplication — Phase 2, deferred.
- Category preference and per-category quote fetching — Phase 3, deferred.
- `supported_timezones` DB table — future migration, not needed for v1.
- Retry logic for failed sends — blocked on concurrent broadcast sends, separately deferred.
- Concurrent broadcast sends — current sequential model is acceptable at current user count.

## Further Notes

- The `bot_config` key-value table already exists and is used for `telegram_offset`. Adding `send_hour` follows the exact same pattern.
- The `/timezone` command callback flow should follow the existing callback pattern: call `answerCallbackQuery` first, then update the DB, then reply with confirmation.
- Users who set their timezone will receive their first timezone-adjusted quote on the next cron fire that matches their local send hour — there is no immediate send on setting the timezone.

## Progress

### Done (branch `feat/tz-aware-broadcast`)
- **DB schema:** `users.timezone TEXT` column added (nullable, no default); `send_hour=9` seeded in `bot_config`.
- **`db.GetSendHour`:** loads the configured hour at startup. `ErrNoRows` handling now lives on the `row.Scan` error path (previously on the `ParseInt` path where it could never match — fixed).
- **`db.GetSubscribedUsersForHour(ctx, db, nowUTC, sendHour)`:** filters subscribed users in SQL using `EXTRACT(HOUR FROM $1 AT TIME ZONE COALESCE(timezone, 'UTC')) = $2`. Ready for broadcast to call.
- **SendHour wiring:** `Broadcast.sendHour` (private field) + `Broadcast.UpdateSendHour(int)` setter. `app.Start` captures the DB value and calls the setter right after `GetSendHour`. See "Config / `sendHour` placement" above for why it's on `Broadcast`, not `Config`.
- **`/timezone` command:** two-level continent → zone picker covering 67 zones, callback prefix routing (`tz-cont:*`, `tz:*`), allowlist validation before DB write, defensive upsert both at command entry and at zone-select (stale-button case), confirmation reply. See `timezone-command-session.md` for details.
- **`/about` copy:** updated to list `/timezone` and describe the new once-a-day-at-local-hour behavior.

### Pending
- First test for `GetSubscribedUsersForHour` (PRD-flagged candidate).

### Done (continued)
- **GH #10 — broadcast wiring:** `broadcast.Run(ctx, nowUTC time.Time)` now takes the fire-time UTC timestamp and calls `db.GetSubscribedUsersForHour(ctx, b.Database, nowUTC, b.sendHour)` instead of the old unfiltered fetch. Old `GetSubscribedUsers` deleted.
- **GH #11 — scheduler wiring + cron flip:** `app.Start`'s closure captures `time.Now().UTC()` on every fire (`a.Scheduler.Start(ctx, func() { a.Broadcast.Run(ctx, time.Now().UTC()) })`), keeping `scheduler.Start` generic. Default `--schedule`/`-s` in `config/config.go` flipped from `0 */6 * * *` → `0 * * * *`. Landed together with #10 so the cron cadence change never existed without the filter.

### Known follow-ups in current code
- **`GetTelegramOffset` has the same dead `ErrNoRows` branch** we just fixed in `GetSendHour` — the `errors.Is(convErr, sql.ErrNoRows)` check lives on the `ParseInt` error path where it can never match. Missing row surfaces as an error caught by the `log.Println` + continue in `app.Start`, so no functional bug; just dead code to clean up.

### Deferred to pre-prod polish
- Back button on the zone-level keyboard. Needs `editMessageText` / `editMessageReplyMarkup` (swap keyboard in place) rather than new `sendMessage` bubbles. Worth doing before prod; not before feature completion. See `timezone-command-session.md` for the implementation sketch.
- BotFather registration for `/timezone` after PR merge.
