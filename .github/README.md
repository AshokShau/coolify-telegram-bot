# Coolify Manager Bot

A Telegram bot to manage your Coolify projects, primarily focused on scheduling application restarts.

## Features

- **Schedule Restarts**: Schedule your applications to restart automatically.
- **Support for various intervals**: Run tasks hourly, daily, weekly, monthly, or with custom intervals (e.g., every 3 days).
- **Time Zone**: The bot operates in **Asia/Kolkata** time zone.
- **Project Management**: List projects, view status, restart manually, etc.

## Commands

### `/start`
Start the bot and see the main menu.

### `/ping`
Check if the bot is running.

### `/jobs`
List all currently scheduled tasks. You can manage them interactively.

### `/schedule <app_name> <schedule_type> [time]`
Schedule a restart for an application.

**Arguments:**
- `<app_name>`: The name of the application in Coolify (case-insensitive).
- `<schedule_type>`: The frequency of the restart.
    - `one_time`: Run once at a specific date/time.
    - `hourly`: Run every hour at minute 0.
    - `daily`: Run every day at midnight (00:00).
    - `weekly`: Run every Sunday at midnight.
    - `monthly`: Run on the 1st of every month at midnight.
    - `yearly`: Run on January 1st at midnight.
    - `cron`: Use a custom cron expression.
    - `every_Xd`: Run every X days (e.g., `every_2d`).
    - `every_Xh`: Run every X hours (e.g., `every_6h`).
    - `every_Xm`: Run every X minutes (e.g., `every_30m`).

**Time Argument (Optional for `daily` and `every_Xd`):**
You can specify a time in `HH:MM` format (24-hour) for daily and day-interval schedules.

**Examples:**
- Restart "my-app" every day at 6:00 AM:
  ```
  /schedule my-app daily 06:00
  ```
- Restart "worker" every 3 days at 2:30 PM:
  ```
  /schedule worker every_3d_at_14:30
  ```
  *(Or shorthand: `/schedule worker 3d_at_14:30`)*
- Restart "api" once on specific date:
  ```
  /schedule api one_time 2023-12-31T23:59:00Z
  ```
- Restart "backup" using cron (every Monday at 3 AM):
  ```
  /schedule backup cron 0 3 * * 1
  ```

### `/unschedule <task_id>`
Remove a scheduled task by its ID (found in `/jobs`).
Alias: `/rmJob`

## Setup

1.  **Environment Variables**:
    Create a `.env` file based on `sample.env` and fill in your details:
    - `BOT_TOKEN`: Your Telegram Bot Token.
    - `API_ID` & `API_HASH`: Your Telegram API credentials.
    - `COOLIFY_URL`: URL of your Coolify instance.
    - `COOLIFY_TOKEN`: Your Coolify API Token.
    - `DB_URL`: MongoDB connection string.
    - `DEV_IDS`: Comma-separated list of Telegram User IDs authorized to use the bot.

2.  **Run with Docker**:
    ```bash
    docker build -t coolifymanager .
    docker run -d --env-file .env --name coolifymanager coolifymanager
    ```

## Time Zone

The default time zone is set to **Asia/Kolkata**. All times in schedules (like `06:00`) are interpreted in this time zone.


Built with [gotdbot](https://github.com/AshokShau/gotdbot), powered by Coolify's REST API.

---

### ⚙️ Features

* 📋 **List all Coolify projects**
* 🔄 **Restart**, 🚀 **Redeploy**, 🛑 **Stop**, ❌ **Delete** apps
* ℹ️ **Check project status** and 📜 **View logs**
* 🔒 **Developer-only features** (via `DEV_IDS`)
* ⚡ Inline button-based UI — no typing needed

---

### 🚀 Deploy Locally

#### 1. Clone the repo

```bash
git clone https://github.com/AshokShau/coolify-telegram-bot
cd coolify-telegram-bot
```

#### 2. Setup environment variables

Create a `.env` file using the template:

```bash
cp sample.env .env
```

Then edit `.env`:

```env
API_ID=
API_HASH=
API_URL=https://app.coolify.io
API_TOKEN=your_coolify_token
TOKEN=your_telegram_bot_token
DEV_IDS=123456789
```

#### 3. Run the bot

```bash
go generate
```

```bash
go run main.go
```

---

### 📄 Coolify API Endpoints Used

This bot integrates with Coolify using:

* `GET /applications`
* `GET /applications/:uuid`
* `GET /applications/:uuid/logs`
* `GET /applications/:uuid/envs`
* `GET /applications/:uuid/start`
* `GET /applications/:uuid/restart`
* `GET /applications/:uuid/stop`
* `DELETE /applications/:uuid`

All requests are authenticated via a `Bearer` token.

---

### 📦 Tech Stack

* Language: Go
* API: [Coolify REST API](https://github.com/coollabsio/coolify)

---

### 🛠️ TODO

> Future features and improvements planned:

1. [x] 🔁 Paginated project list with `< Prev | 1 | 2 | 3 | Next >` buttons
2. [x] 🧠 Cache project data to reduce API calls
3. [ ] Add support for more endpoints like Deployments, Environments, Databases and more.

---

### 🙋‍♂️ Support

* Telegram Support: [@GuardxSupport](https://t.me/GuardxSupport)
* Updates Channel: [@FallenProjects](https://t.me/FallenProjects)

---

### 📜 License

MIT — do what you want, just give credit.
© 2025 AshokShau
