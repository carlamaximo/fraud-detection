package worker

import (
	"errors"
	"log/slog"
	"time"

	"fraud-platform/internal/domain"
	"fraud-platform/internal/logging"
	"fraud-platform/internal/queue"
	"fraud-platform/internal/risk"
	"fraud-platform/internal/stats"
)

const maxRetries = 3

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
		logging.L.Info("dequeued",
			slog.Int("worker", id),
			slog.String("user", event.UserID),
			slog.String("type", event.EventType),
		)

		err := processWithRetries(id, event)
		if err != nil {
			sendToDLQ(id, event, err)
			continue
		}

		logging.L.Info("finished", slog.Int("worker", id), slog.String("user", event.UserID))
	}
}

func processWithRetries(workerID int, event domain.Event) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		score, label, err := runProcessing(event)
		if err != nil {
			lastErr = err
			stats.IncFailed()
			logging.L.Warn("process failed",
				slog.Int("worker", workerID),
				slog.String("user", event.UserID),
				slog.Int("attempt", attempt),
				slog.String("err", err.Error()),
			)
			
			if attempt < maxRetries {
				time.Sleep(retryBackoffDelay * time.Duration(attempt))
			}

			continue
		}

		logging.L.Info("processed",
			slog.Int("worker", workerID),
			slog.String("user", event.UserID),
			slog.Float64("score", score),
			slog.String("label", string(label)),
			slog.Int("attempts", attempt),
		)

		stats.RecordSuccessfulProcess(string(label))
		return nil
	}

	return lastErr
}

func runProcessing(event domain.Event) (score float64, label risk.Label, err error) {
	time.Sleep(processingDelay)

	if event.EventType == "fail_processing" {
		return 0, "", errors.New("processing rejected")
	}

	score = calculateRisk(event)
	label = risk.Classify(score)
	return score, label, nil
}

func sendToDLQ(workerID int, event domain.Event, err error) {
	queue.PushDeadLetter(event, err.Error())
	stats.IncSentToDLQ()

	logging.L.Error("dlq",
		slog.Int("worker", workerID),
		slog.String("user", event.UserID),
		slog.String("err", err.Error()),
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
