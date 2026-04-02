package worker

import (
	"errors"
	"os"
	"testing"

	"fraud-platform/internal/decision"
	"fraud-platform/internal/domain"
	"fraud-platform/internal/logging"
	"fraud-platform/internal/metrics"
	"fraud-platform/internal/queue"
)

func TestMain(m *testing.M) {
	logging.Init()

	origProc := processingDelay
	origBackoff := retryBackoffDelay
	processingDelay = 0
	retryBackoffDelay = 0

	code := m.Run()

	processingDelay = origProc
	retryBackoffDelay = origBackoff
	os.Exit(code)
}

func TestCalculateRisk(t *testing.T) {
	tests := []struct {
		name  string
		event domain.Event
		want  float64
	}{
		{
			name: "baseline_br",
			event: domain.Event{
				UserID: "1", EventType: "signup", IP: "8.8.8.8",
				Country: "BR", Device: "iphone",
			},
			want: 0,
		},
		{
			name: "non_br",
			event: domain.Event{
				UserID: "1", EventType: "signup", IP: "8.8.8.8",
				Country: "US", Device: "iphone",
			},
			want: 0.3,
		},
		{
			name: "unknown_device",
			event: domain.Event{
				UserID: "1", EventType: "signup", IP: "8.8.8.8",
				Country: "BR", Device: "unknown",
			},
			want: 0.2,
		},
		{
			name: "flagged_ip",
			event: domain.Event{
				UserID: "1", EventType: "signup", IP: "1.1.1.1",
				Country: "BR", Device: "iphone",
			},
			want: 0.2,
		},
		{
			name: "combined_us_unknown_ip",
			event: domain.Event{
				UserID: "1", EventType: "signup", IP: "1.1.1.1",
				Country: "US", Device: "unknown",
			},
			want: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRisk(tt.event)
			if got != tt.want {
				t.Fatalf("calculateRisk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunProcessing(t *testing.T) {
	t.Run("fail_processing", func(t *testing.T) {
		_, _, err := runProcessing(domain.Event{
			UserID: "u", EventType: "fail_processing", IP: "1.1.1.1",
			Country: "BR", Device: "iphone",
		})
		
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		score, outcome, err := runProcessing(domain.Event{
			UserID: "u", EventType: "signup", IP: "8.8.8.8",
			Country: "BR", Device: "iphone",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if want := decision.FromScore(score); outcome != want {
			t.Fatalf("outcome %q vs FromScore %q", outcome, want)
		}

		if score != 0 {
			t.Fatalf("score = %v, want 0", score)
		}
	})
}

func TestProcessWithRetries_success(t *testing.T) {
	metrics.Reset()
	t.Cleanup(metrics.Reset)

	ev := domain.Event{
		UserID: "u", EventType: "signup", IP: "8.8.8.8",
		Country: "BR", Device: "iphone",
	}

	if err := processWithRetries(1, ev); err != nil {
		t.Fatalf("processWithRetries: %v", err)
	}

	s := metrics.GetSnapshot()
	if s.TotalProcessed != 1 || s.TotalApproved != 1 {
		t.Fatalf("metrics: %+v", s)
	}
}

func TestProcessWithRetries_exhaustsRetries(t *testing.T) {
	metrics.Reset()
	queue.ResetDLQ()
	t.Cleanup(func() {
		metrics.Reset()
		queue.ResetDLQ()
	})

	ev := domain.Event{
		UserID: "u", EventType: "fail_processing", IP: "8.8.8.8",
		Country: "BR", Device: "iphone",
	}

	err := processWithRetries(7, ev)
	if err == nil {
		t.Fatal("expected error after retries")
	}

	if got := metrics.GetSnapshot().TotalFailed; got != 3 {
		t.Fatalf("TotalFailed = %d, want 3", got)
	}
}

func TestSendToDLQ(t *testing.T) {
	metrics.Reset()
	queue.ResetDLQ()
	t.Cleanup(func() {
		metrics.Reset()
		queue.ResetDLQ()
	})

	ev := domain.Event{
		UserID: "dlq", EventType: "signup", IP: "1.1.1.1",
		Country: "BR", Device: "iphone",
	}
	sendToDLQ(2, ev, errors.New("boom"))

	if metrics.GetSnapshot().TotalSentToDLQ != 1 {
		t.Fatal("expected DLQ metric increment")
	}

	snap := queue.DeadLettersSnapshot()
	if len(snap) != 1 || snap[0].Reason != "boom" {
		t.Fatalf("unexpected DLQ snapshot: %+v", snap)
	}
}
