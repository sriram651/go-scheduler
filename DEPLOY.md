# Deployment Guide

How to build and push a new binary to the Hostinger VPS, with secrets
managed entirely on the server via a systemd service file.

No secrets are committed to this repository.

------------------------------------------------------------------------

## Prerequisites

- SSH access to the VPS
- Go installed locally (for building)
- The systemd service already set up on the server (see First-Time Setup)

------------------------------------------------------------------------

## Building the Binary

Cross-compile for Linux (the VPS target) from your local machine:

    GOOS=linux GOARCH=amd64 go build -o go-scheduler ./cmd/scheduler/

This produces a `go-scheduler` binary in the project root.

> If your VPS uses a different architecture (e.g. ARM), set
> `GOARCH=arm64` instead.

------------------------------------------------------------------------

## Pushing a New Build

Copy the binary to the VPS:

    scp go-scheduler user@your-vps-ip:/path/to/app/go-scheduler

Then restart the service to pick up the new binary:

    ssh user@your-vps-ip "sudo systemctl restart go-scheduler"

Verify it came back up cleanly:

    ssh user@your-vps-ip "sudo systemctl status go-scheduler"

------------------------------------------------------------------------

## Checking Logs

    ssh user@your-vps-ip "sudo journalctl -u go-scheduler -f"

Or to see the last 50 lines without following:

    ssh user@your-vps-ip "sudo journalctl -u go-scheduler -n 50 --no-pager"

------------------------------------------------------------------------

## First-Time Setup on the VPS

### 1. Copy the binary

    scp go-scheduler user@your-vps-ip:/path/to/app/go-scheduler
    ssh user@your-vps-ip "chmod +x /path/to/app/go-scheduler"

### 2. Create the systemd service file

On the VPS, create `/etc/systemd/system/go-scheduler.service`:

    sudo nano /etc/systemd/system/go-scheduler.service

Paste the following, filling in real values under `[Service]`:

```ini
[Unit]
Description=Go Cron Telegram Reminder Service
After=network.target

[Service]
Type=simple
ExecStart=/path/to/app/go-scheduler --schedule "0 * * * *"
Restart=on-failure
RestartSec=10

Environment="TG_BOT_TOKEN=your_tg_bot_token_here"
Environment="TG_API_BASE_URL=https://api.telegram.org/bot"
Environment="QUOTE_API_URL=https://your-quote-api.com/api/random"
Environment="DEFAULT_QUOTE=Your fallback quote here."
Environment="DATABASE_URL=postgres://user:password@host:5432/dbname"

[Install]
WantedBy=multi-user.target
```

> The `Environment=` lines live only on the server and are never
> committed to git. This is the only place secrets should exist.

### 3. Enable and start the service

    sudo systemctl daemon-reload
    sudo systemctl enable go-scheduler
    sudo systemctl start go-scheduler

### 4. Confirm it is running

    sudo systemctl status go-scheduler

------------------------------------------------------------------------

## Updating Environment Variables

To change a secret or config value:

1. Edit the service file on the VPS:

        sudo nano /etc/systemd/system/go-scheduler.service

2. Update the relevant `Environment=` line.

3. Reload and restart:

        sudo systemctl daemon-reload
        sudo systemctl restart go-scheduler

------------------------------------------------------------------------

## Summary

| Action          | Command                                                             |
| --------------- | ------------------------------------------------------------------- |
| Build for Linux | `GOOS=linux GOARCH=amd64 go build -o go-scheduler ./cmd/scheduler/` |
| Push binary     | `scp go-scheduler user@vps-ip:/path/to/app/`                        |
| Restart service | `sudo systemctl restart go-scheduler`                               |
| Check status    | `sudo systemctl status go-scheduler`                                |
| Tail logs       | `sudo journalctl -u go-scheduler -f`                                |
| Edit env vars   | `sudo nano /etc/systemd/system/go-scheduler.service`                |
