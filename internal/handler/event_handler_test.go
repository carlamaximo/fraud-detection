package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"fraud-platform/internal/logging"
	"fraud-platform/internal/queue"
	"fraud-platform/internal/stats"
)

func TestMain(m *testing.M) {
	logging.Init()
	os.Exit(m.Run())
}

func resetIntegrationState(t *testing.T) {
	t.Helper()
	stats.Reset()
	queue.ResetDLQ()
	queue.DrainEventQueue()
	t.Cleanup(func() {
		queue.DrainEventQueue()
		stats.Reset()
		queue.ResetDLQ()
	})
}

func TestHandleEvent(t *testing.T) {
	t.Run("method_not_allowed", func(t *testing.T) {
		resetIntegrationState(t)
		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		rec := httptest.NewRecorder()
		
		HandleEvent(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d", rec.Code)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		resetIntegrationState(t)
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{`))
		rec := httptest.NewRecorder()

		HandleEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rec.Code)
		}
	})

	t.Run("validation_required_fields", func(t *testing.T) {
		resetIntegrationState(t)
		body := `{"user_id":"","event_type":"x","ip":"1.1.1.1","country":"BR","device":"iphone"}`
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		HandleEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rec.Code)
		}
	})

	t.Run("business_rule_rejects_br_unknown", func(t *testing.T) {
		resetIntegrationState(t)
		body := `{"user_id":"u1","event_type":"login","ip":"1.1.1.1","country":"BR","device":"unknown"}`
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		HandleEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rec.Code)
		}
	})

	t.Run("accepted", func(t *testing.T) {
		resetIntegrationState(t)
		body := `{"user_id":"u1","event_type":"login","ip":"1.1.1.1","country":"BR","device":"iphone"}`
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		HandleEvent(rec, req)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d", rec.Code)
		}

		if stats.GetSnapshot().TotalEventsReceived != 1 {
			t.Fatal("TotalEventsReceived not incremented")
		}

		select {
		case ev := <-queue.EventQueue:
			if ev.UserID != "u1" {
				t.Fatalf("queued user_id = %q", ev.UserID)
			}
			
		default:
			t.Fatal("expected event in queue")
		}
	})
}
