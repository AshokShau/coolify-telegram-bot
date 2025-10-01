## 📦 Coolify Manager Bot

A modern Telegram bot to **manage your Coolify applications** via buttons — no CLI, no dashboards.

Built with [gogram](https://github.com/AmarnathCJD/gogram), powered by Coolify's REST API.

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
API_URL=https://app.coolify.io
API_TOKEN=your_coolify_token
TOKEN=your_telegram_bot_token
PORT=8080
WEBHOOK_URL=https://yourdomain.com/webhook
ENV=dev
DEV_IDS=123456789
```

#### 3. Run the bot

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

* [ ] 🔁 Paginated project list with `< Prev | 1 | 2 | 3 | Next >` buttons
* [ ] 🧠 Cache project data to reduce API calls
* [ ] Add support for more endpoints like Deployments, Environments, Databases and more.

---

### 🙋‍♂️ Support

* Telegram Support: [@GuardxSupport](https://t.me/GuardxSupport)
* Updates Channel: [@FallenProjects](https://t.me/FallenProjects)

---

### 📜 License

MIT — do what you want, just give credit.
© 2025 AshokShau
