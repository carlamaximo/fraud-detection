package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"fraud-platform/internal/domain"
	"fraud-platform/internal/logging"
	"fraud-platform/internal/metrics"
	"fraud-platform/internal/queue"
	"fraud-platform/internal/validators"

	playground "github.com/go-playground/validator/v10"
)

var validate = playground.New()

func HandleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event domain.Event

	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logging.L.Info("event_received",
		slog.String("component", "api"),
		slog.String("user_id", event.UserID),
		slog.String("event_type", event.EventType),
		slog.String("risk_score", "pending"),
		slog.String("decision", "pending"),
		slog.Int("retry_count", 0),
		slog.Bool("dlq", false),
	)

	err = validate.Struct(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validators.ValidateBusinessRules(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queue.EventQueue <- event
	metrics.IncEventsReceived()

	w.WriteHeader(http.StatusAccepted)
}
