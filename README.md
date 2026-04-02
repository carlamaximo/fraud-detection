# Fraud Platform

Asynchronous fraud-style event processing in Go: HTTP ingest, buffered queue, worker pool, simple risk score → **APPROVE / REVIEW / BLOCK**, retries, in-memory DLQ, and JSON endpoints for counters and DLQ inspection.

## What it does

- HTTP API to receive events  
- Buffered channel as a queue  
- Worker pool for concurrent processing  
- Heuristic risk score and classification (`internal/risk`)  
- Retries with backoff, then dead-letter queue  
- In-memory stats (`GET /metrics`) and DLQ listing (`GET /dlq`)  
- Structured logs (`log/slog`)

## Architecture

```text
Client
  ↓
POST /events
  ↓
Validation
  ↓
Buffered Event Queue
  ↓
Worker Pool
  ↓
Risk Scoring
  ↓
Decision Layer
  ↓
Success metrics / Retry / DLQ
```

## Requirements
- [Go](https://go.dev/dl/) **1.22+** (see `go.mod`).

## Run

```bash
go run ./cmd/api
```

Server listens on :8080.

## API

### `POST /events`

Accepts JSON. All fields are required.

```json
{
  "user_id": "user-123",
  "event_type": "signup",
  "ip": "203.0.113.10",
  "country": "BR",
  "device": "iphone"
}
```

**Example**

```bash
curl -s -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u1","event_type":"login","ip":"1.1.1.1","country":"BR","device":"iphone"}'
```

**Response:** `202 Accepted` — event is queued; processing finishes asynchronously.

**Notes**

- `country: "BR"` with `device: "unknown"` fails business validation (intentional demo rule).
- Use `"event_type": "fail_processing"` to force worker failure and exercise retries + DLQ.

### `GET /metrics`

```bash
curl -s http://localhost:8080/metrics
```

**Example response**

```json
{
  "total_events_received": 5,
  "total_processed": 5,
  "total_failed": 1,
  "total_sent_to_dlq": 1,
  "total_approved": 2,
  "total_review": 2,
  "total_blocked": 1
}
```

### `GET /dlq`

```bash
curl -s http://localhost:8080/dlq
```

Returns JSON for events that exhausted retries (in-memory only).

## Project layout

```text
fraud-platform/
├── cmd/api/main.go
├── internal/
│   ├── domain/       # Event model
│   ├── handler/      # HTTP (events + admin: metrics, dlq)
│   ├── logging/      # slog setup
│   ├── queue/        # channel + DLQ slice
│   ├── risk/         # score → APPROVE / REVIEW / BLOCK
│   ├── stats/        # counters for /metrics
│   ├── validators/   # business rules
│   └── worker/       # consumers, retry, processing
├── go.mod
└── go.sum
```

## Event flow

1. API receives and decodes JSON  
2. Structural + business validation  
3. Event pushed onto buffered channel  
4. Worker picks up event  
5. Risk score computed, label assigned  
6. Stats updated on success  
7. On repeated failure → DLQ + stat increment  

## Limitations

- Queue, DLQ, and stats live in process memory — no persistence or multi-instance safety.  
- No authentication on HTTP endpoints.  
- Risk rules are illustrative, not a production fraud model.

## Possible next steps

- Persistence (e.g. PostgreSQL)  
- Real broker (Kafka, SQS, Redis streams)  
- Docker / Compose, config via env  
- Tracing, load tests, auth, rate limits  

## Tests

```bash
go test ./...
```
