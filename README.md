# Webhook Inbox

[![Go CI](https://github.com/YOUR_GH_USERNAME/webhook-inbox/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/YOUR_GH_USERNAME/webhook-inbox/actions)
[![Docker Publish](https://github.com/YOUR_GH_USERNAME/webhook-inbox/actions/workflows/backend-docker.yml/badge.svg)](https://github.com/YOUR_GH_USERNAME/webhook-inbox/actions)

A lightweight Go + Postgres service to **capture, inspect, and replay webhooks**.  
Perfect for debugging integrations with **GitHub, Stripe, Paystack, Flutterwave**, and more.

---

## Features
- Endpoint provisioning – generate unique webhook URLs with secrets  
- Message storage – save headers + body in PostgreSQL  
- Inspect payloads – see exactly what your service received  
- Replay – resend stored messages to any target URL  
- Provider verification – GitHub, Stripe, Paystack, Flutterwave  
- Dockerized – one-command local setup  

---

## Tech Stack
- **Go** (chi router, net/http)  
- **PostgreSQL**  
- **Docker & Docker Compose**  
- **GitHub Actions** (CI and Docker publishing)  

---

## Quickstart

```bash
# clone
git clone https://github.com/YOUR_GH_USERNAME/webhook-inbox.git
cd webhook-inbox

# (optional) copy example env file
cp backend/.env.example backend/.env

# run stack
docker compose up --build -d

# verify
curl http://localhost:8080/health
# -> ok
````

---

## Usage

### 1. Create an endpoint (GitHub example)

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

---

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

---

### 3. List messages

```bash
curl -s "http://localhost:8080/api/messages?limit=5"
```

---

### 4. Replay a message

Replay ID `1` to a test server:

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

Make sure the **secret** in GitHub matches the one you set in `/api/endpoints`.

---

## API Reference

* `POST /api/endpoints` – create new endpoint
* `POST /hooks/{token}` – receive webhook (signature verified)
* `GET /api/messages` – list stored messages
* `GET /api/messages/{id}` – fetch one message
* `POST /api/messages/{id}/replay` – resend message to target URL

---

## Development

```bash
# run stack
docker compose up --build

# tail logs
docker compose logs -f backend
```

---

## Security Notes

* Keep `.env` out of version control
* Use HTTPS when exposing publicly
* Rotate secrets regularly

---

## License

MIT — free to use, modify, and distribute.

---

## Contributing

Pull requests welcome. For major changes, please open an issue first.

