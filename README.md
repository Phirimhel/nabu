# Nabutilivanie backend

Compact Go backend for a Telegram Mini App. It separates **players** (identified by a verified Telegram ID in production) from anonymous leaderboard **targets**.

## Structure

```text
cmd/api                         application composition root
internal/domain                 business entities
internal/application            use cases and dependency ports
internal/adapters/postgres      PostgreSQL repository implementation
internal/adapters/redis         Redis cooldown adapter
internal/adapters/http/controllers
  *_handler.go                  one HTTP handler group per file
  web/                          embedded Mini App interface
migrations                      versioned PostgreSQL schema
```

## Start locally

1. Copy `.env.example` to `.env` and set a strong `WEBHOOK_SECRET`.
2. Start services: `docker compose up --build`. The one-shot `migrate` service records and applies the supplied migration before the API starts.
3. Open `http://localhost:8079` for the dark Telegram Mini App-style interface.

## API

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/v1/targets` | Add anonymous target: `name`, `photoUrl`, `description` |
| GET | `/api/v1/targets?limit=50` | Global leaderboard |
| GET | `/api/v1/targets/search?q=name` | Full-text target search |
| POST | `/api/v1/me/claim/active` | Claim one bottle; requires `X-Telegram-User-ID` |
| POST | `/api/v1/me/claim/passive` | Claim elapsed hourly bottles |
| POST | `/api/v1/targets/{id}/nabutilit` | Spend 1,000 bottles and submit immutable `comment` |
| POST | `/api/v1/payments/webhook` | HMAC-SHA256 payment webhook |

The `X-Telegram-User-ID` header is a development seam only. A production TMA must validate Telegram `initData` server-side, derive the user ID from it, and never trust a client-supplied identity header.

Webhook body example:

```json
{"provider":"cryptopay","externalId":"invoice-123","telegramId":123456,"bottles":1200}
```

Sign the exact raw JSON with `HMAC-SHA256(WEBHOOK_SECRET, body)` and send its lowercase hex digest in `X-Webhook-Signature`. `provider + externalId` is unique, making retries safe: only the first event credits the balance.

## Design notes

- PostgreSQL transactions make the 1,000-bottle spend, leaderboard update, and immutable comment atomic.
- Redis `SETNX` with a 40-second TTL enforces the active-claim cooldown.
- Passive rewards are calculated from the database timestamp, so they survive restarts.
- Payment providers should verify their own webhook signature before translating their event into this internal event format; keep a separate adapter per provider.
