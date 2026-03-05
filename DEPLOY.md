# Deployment Guide

The service runs as a Docker container on a Hostinger VPS. Deployments
are fully automated via GitHub Actions — every push to `main` builds a
new image, pushes it to GitHub Container Registry, and restarts the
container on the VPS.

No secrets are committed to this repository.

------------------------------------------------------------------------

## How Deployments Work

1. Push to `main` triggers the GitHub Actions workflow
2. Actions builds the Docker image and pushes it to `ghcr.io/sriram651/go-scheduler:latest`
3. Actions SSHes into the VPS and runs:
   - `docker pull` — fetches the new image
   - `docker stop` + `docker rm` — removes the old container
   - `docker run` — starts a fresh container from the new image

------------------------------------------------------------------------

## Prerequisites

- SSH access to the VPS
- Docker installed on the VPS
- GitHub Actions secrets configured: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`

------------------------------------------------------------------------

## First-Time VPS Setup

### 1. Install Docker

    curl -fsSL https://get.docker.com | sh

### 2. Set up the database schema

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

### 3. Create the env file

Create `/etc/go-scheduler.env` on the VPS:

    sudo nano /etc/go-scheduler.env

Paste the following, filling in real values:

    TG_BOT_TOKEN=your_tg_bot_token_here
    TG_API_BASE_URL=https://api.telegram.org/bot
    QUOTE_API_URL=https://your-quote-api.com/api/random
    DEFAULT_QUOTE=Your fallback quote here.
    DATABASE_URL=postgres://user:password@host:5432/dbname

> This file lives only on the server and is never committed to git.

### 4. Run the container manually (first time)

    docker run -d \
      --name go-scheduler \
      --restart=unless-stopped \
      --env-file /etc/go-scheduler.env \
      ghcr.io/sriram651/go-scheduler:latest

### 5. Verify it is running

    docker ps
    docker logs go-scheduler

------------------------------------------------------------------------

## Checking Logs

Follow live logs:

    docker logs -f go-scheduler

Last 50 lines:

    docker logs --tail 50 go-scheduler

------------------------------------------------------------------------

## Updating Environment Variables

To change a secret or config value:

1. Edit the env file on the VPS:

        sudo nano /etc/go-scheduler.env

2. Restart the container to pick up the new values:

        docker stop go-scheduler
        docker rm go-scheduler
        docker run -d --name go-scheduler --restart=unless-stopped --env-file /etc/go-scheduler.env ghcr.io/sriram651/go-scheduler:latest

------------------------------------------------------------------------

## GitHub Actions Secrets Required

| Secret        | Description                        |
| ------------- | ---------------------------------- |
| `VPS_HOST`    | IP address of the VPS              |
| `VPS_USER`    | SSH username (e.g. `root`)         |
| `VPS_SSH_KEY` | Ed25519 private key for SSH access |

------------------------------------------------------------------------

## Summary

| Action            | Command                                                      |
| ----------------- | ------------------------------------------------------------ |
| Deploy            | Push to `main` — Actions handles the rest                    |
| Check status      | `docker ps`                                                  |
| Tail logs         | `docker logs -f go-scheduler`                                |
| Restart container | `docker restart go-scheduler`                                |
| Update env vars   | Edit `/etc/go-scheduler.env`, then stop/rm/run the container |
