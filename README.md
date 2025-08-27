# Webhook Inbox

[![Go CI](https://github.com/sudo-init-do/webhook-inbox/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/sudo-init-do/webhook-inbox/actions)
[![Docker Publish](https://github.com/sudo-init-do/webhook-inbox/actions/workflows/backend-docker.yml/badge.svg)](https://github.com/sudo-init-do/webhook-inbox/actions)

A modern Go + Postgres service to **capture, inspect, and replay webhooks**. Fast, secure, and developer-friendly. Perfect for debugging integrations with **GitHub, Stripe, Paystack, Flutterwave**, and more.

---

## Table of Contents
- [Webhook Inbox](#webhook-inbox)
  - [Table of Contents](#table-of-contents)
  - [🚀 Features](#-features)
  - [🛠️ Tech Stack](#️-tech-stack)
  - [⚡ Quickstart](#-quickstart)
  - [📖 Usage Examples](#-usage-examples)
    - [1. Create an endpoint (GitHub)](#1-create-an-endpoint-github)
    - [2. Send a signed GitHub webhook](#2-send-a-signed-github-webhook)
    - [3. List messages](#3-list-messages)
    - [4. Replay a message to a test server](#4-replay-a-message-to-a-test-server)
  - [🌐 Expose to GitHub with ngrok](#-expose-to-github-with-ngrok)
  - [📚 API Reference](#-api-reference)
  - [🧑‍💻 Development](#-development)
  - [🔒 Security Notes](#-security-notes)
  - [📝 License](#-license)
  - [🤝 Contributing](#-contributing)

---

## 🚀 Features
- **Endpoint provisioning** – generate unique webhook URLs with secrets
- **Message storage** – save headers + body in PostgreSQL
- **Payload inspection** – view exactly what your service received
- **Replay** – resend stored messages to any target URL
- **Provider verification** – built-in checks for GitHub, Stripe, Paystack, Flutterwave
- **Dockerized** – one-command local setup

---

## 🛠️ Tech Stack
- **Go** (chi router, net/http)
- **PostgreSQL**
- **Docker & Docker Compose**
- **GitHub Actions** (CI and Docker publishing)

---

## ⚡ Quickstart
```bash
# Clone the repo
git clone https://github.com/sudo-init-do/webhook-inbox.git
cd webhook-inbox

# (optional) copy example env file
cp backend/.env.example backend/.env

# Start the stack
docker compose up --build -d

# Verify health
curl http://localhost:8080/health
# -> ok
```

---

## 📖 Usage Examples

### 1. Create an endpoint (GitHub)
```bash
SECRET=supersecret
curl -s -X POST http://localhost:8080/api/endpoints \
  -H "Content-Type: application/json" \
  -d '{"provider":"github","secret":"'"$SECRET"'"}'
```
Response:
```json
{
  "id": 1,
  "token": "e3b0c442-98fc-4624-b87e-b567f77934f2",
  "url": "http://localhost:8080/hooks/e3b0c442-98fc-4624-b87e-b567f77934f2"
}
```

### 2. Send a signed GitHub webhook
```bash
HOOK_URL="http://localhost:8080/hooks/<token-from-above>"
PAYLOAD='{"hello":"world"}'
SIG=$(printf '%s' "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.*= //')

curl -i -X POST "$HOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: ping" \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  -d "$PAYLOAD"
```

### 3. List messages
```bash
curl -s "http://localhost:8080/api/messages?limit=5"
```

### 4. Replay a message to a test server
```bash
curl -s -X POST "http://localhost:8080/api/messages/1/replay" \
  -H "Content-Type: application/json" \
  -d '{"target_url":"https://httpbin.org/post"}'
```

---

## 🌐 Expose to GitHub with ngrok
```bash
ngrok http 8080
```
Use the public URL as your webhook target:
`https://<ngrok-id>.ngrok-free.app/hooks/<token>`
**Note:** The secret in your GitHub webhook settings must match the one you set when creating the endpoint.

---

## 📚 API Reference
| Method | Path | Description |
|--------|------|-------------|
| POST   | `/api/endpoints`         | Create new endpoint |
| POST   | `/hooks/{token}`         | Receive webhook (signature verified) |
| GET    | `/api/messages`          | List stored messages |
| GET    | `/api/messages/{id}`     | Fetch one message |
| POST   | `/api/messages/{id}/replay` | Resend message to target URL |

---

## 🧑‍💻 Development
```bash
# Start stack
docker compose up --build

# Tail logs
docker compose logs -f backend

# Run tests (inside backend container)
docker compose exec backend go test ./... -v
```

---

## 🔒 Security Notes
- **Never commit `.env` or secrets to version control**
- **Use HTTPS** when exposing publicly
- **Rotate secrets** regularly

---

## 📝 License
MIT — free to use, modify, and distribute.

---

## 🤝 Contributing
Pull requests are welcome! For major changes, please open an issue first to discuss what you’d like to change.

