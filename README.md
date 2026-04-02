# Fraud Platform

Asynchronous fraud detection backend built in Go.

This project simulates a production-like fraud processing pipeline:
- HTTP API to receive events
- asynchronous queue-based processing
- worker pool
- risk scoring
- decision engine (`APPROVE`, `REVIEW`, `BLOCK`)
- retry mechanism
- dead letter queue (DLQ)
- in-memory metrics and observability endpoints

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
go run ./cmd/api

Server listens on :8080.

API
POST /events
Body (all fields required):
{
  "user_id": "user-123",
  "event_type": "signup",
  "ip": "203.0.113.10",
  "country": "BR",
  "device": "iphone"
}

Example:
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u1","event_type":"login","ip":"1.1.1.1","country":"BR","device":"iphone"}'

Response:
202 Accepted
  
Note: country: "BR" with device: "unknown" fails business validation (example rule).

To exercise retries + DLQ, send "event_type": "fail_processing" (processing is simulated to fail every attempt).

GET /metrics
curl http://localhost:8080/metrics
Example response:
{
  "total_events_received": 5,
  "total_processed": 5,
  "total_failed": 1,
  "total_sent_to_dlq": 1,
  "total_approved": 2,
  "total_review": 2,
  "total_blocked": 1
}

GET /dlq
curl http://localhost:8080/dlq
Returns events that failed processing permanently.

## Project structure

fraud-platform/
  ├── cmd/
  │     └── api/
  │          └── main.go
  ├── internal/
  │     ├── decision/
  │     ├── domain/
  │     ├── handler/
  │     ├── logging/
  │     ├── metrics/
  │     ├── queue/
  │     ├── validators/
  │     ├── worker/
  ├── go.mod

## Features
- Receive fraud-related events via HTTP API
- Validate payload and business rules
- Queue events for asynchronous processing
- Process events concurrently using worker pool
- Calculate fraud risk score
- Decision engine:
  - APPROVE
  - REVIEW
  - BLOCK
- Retry mechanism for failed events
- Dead Letter Queue for permanently failed events
- Metrics endpoint
- DLQ inspection endpoint
- Structured logging

## Event Processing Flow
1. API receives event
2. Event is validated
3. Event is pushed to buffered queue
4. Worker consumes event
5. Risk score is calculated
6. Decision layer classifies event
7. Metrics updated
8. If processing fails repeatedly → DLQ

## Limitations (future ideas)
In-memory queue and DLQ only, no persistence or multi-instance safety.
No authentication on HTTP endpoints.
Risk rules are illustrative, not production fraud models.

## Possible Future Improvements
PostgreSQL persistence
Kafka or AWS SQS integration
Docker / Docker Compose
Environment-based configuration
Distributed tracing
Load testing
Authentication
Rate limiting
