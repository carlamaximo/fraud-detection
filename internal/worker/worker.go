package worker

import (
	"errors"
	"log/slog"
	"time"

	"fraud-platform/internal/decision"
	"fraud-platform/internal/domain"
	"fraud-platform/internal/logging"
	"fraud-platform/internal/metrics"
	"fraud-platform/internal/queue"
)

const (
	maxRetries      = 3
	componentWorker = "worker"
	componentDLQ    = "dlq"
)

// Valores ajustáveis em testes (TestMain) para evitar sleeps longos.
var (
	processingDelay   = 2 * time.Second
	retryBackoffDelay = 200 * time.Millisecond
)

func StartWorkers(n int) {
	for i := 1; i <= n; i++ {
		go workerLoop(i)
	}
}

func workerLoop(id int) {
	for event := range queue.EventQueue {
		logging.L.Info("worker_event_picked_up",
			slog.String("component", componentWorker),
			slog.Int("worker_id", id),
			slog.String("user_id", event.UserID),
			slog.String("event_type", event.EventType),
			slog.String("risk_score", "pending"),
			slog.String("decision", "pending"),
			slog.Int("retry_count", 0),
			slog.Bool("dlq", false),
		)

		err := processWithRetries(id, event)
		if err != nil {
			sendToDLQ(id, event, err)
			continue
		}

		logging.L.Info("worker_event_finished",
			slog.String("component", componentWorker),
			slog.Int("worker_id", id),
			slog.String("user_id", event.UserID),
			slog.Bool("dlq", false),
		)
	}
}

func processWithRetries(workerID int, event domain.Event) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		score, outcome, err := runProcessing(event)
		if err != nil {
			lastErr = err
			metrics.IncFailed()
			logging.L.Info("event_process_attempt_failed",
				slog.String("component", componentWorker),
				slog.Int("worker_id", workerID),
				slog.String("user_id", event.UserID),
				slog.String("risk_score", "n/a"),
				slog.String("decision", "n/a"),
				slog.Int("retry_count", attempt),
				slog.Bool("dlq", false),
				slog.String("error", err.Error()),
			)
		
			if attempt < maxRetries {
				time.Sleep(retryBackoffDelay * time.Duration(attempt))
			}

			continue
		}

		logging.L.Info("event_processed",
			slog.String("component", componentWorker),
			slog.Int("worker_id", workerID),
			slog.String("user_id", event.UserID),
			slog.Float64("risk_score", score),
			slog.String("decision", string(outcome)),
			slog.Int("retry_count", attempt),
			slog.Bool("dlq", false),
		)

		metrics.RecordSuccessfulProcess(string(outcome))

		return nil
	}

	return lastErr
}

func runProcessing(event domain.Event) (score float64, outcome decision.Outcome, err error) {
	time.Sleep(processingDelay)

	if event.EventType == "fail_processing" {
		return 0, "", errors.New("simulated processing failure")
	}

	score = calculateRisk(event)
	outcome = decision.FromScore(score)

	return score, outcome, nil
}

func sendToDLQ(workerID int, event domain.Event, err error) {
	queue.PushDeadLetter(event, err.Error())
	metrics.IncSentToDLQ()

	logging.L.Info("event_sent_to_dlq",
		slog.String("component", componentDLQ),
		slog.Int("worker_id", workerID),
		slog.String("user_id", event.UserID),
		slog.String("risk_score", "n/a"),
		slog.String("decision", "n/a"),
		slog.Int("retry_count", maxRetries),
		slog.Bool("dlq", true),
		slog.String("reason", err.Error()),
	)
}

func calculateRisk(event domain.Event) float64 {
	score := 0.0

	if event.Country != "BR" {
		score += 0.3
	}

	if event.Device == "unknown" {
		score += 0.2
	}

	if event.IP == "1.1.1.1" {
		score += 0.2
	}

	return score
}
