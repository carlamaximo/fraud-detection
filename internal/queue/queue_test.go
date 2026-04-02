package queue

import (
	"os"
	"testing"

	"fraud-platform/internal/domain"
)

func TestMain(m *testing.M) {
	DrainEventQueue()
	ResetDLQ()
	code := m.Run()
	DrainEventQueue()
	ResetDLQ()
	os.Exit(code)
}

func TestPushDeadLetter_and_Snapshot(t *testing.T) {
	ResetDLQ()
	t.Cleanup(ResetDLQ)

	ev := domain.Event{
		UserID: "dlq-user", EventType: "signup", IP: "10.0.0.1",
		Country: "BR", Device: "android",
	}
	PushDeadLetter(ev, "test reason")

	snap := DeadLettersSnapshot()
	if len(snap) != 1 {
		t.Fatalf("len(snap) = %d, want 1", len(snap))
	}

	if snap[0].Reason != "test reason" {
		t.Fatalf("Reason = %q", snap[0].Reason)
	}
	
	if snap[0].Event.UserID != ev.UserID {
		t.Fatalf("Event.UserID = %q", snap[0].Event.UserID)
	}

	if snap[0].At.IsZero() {
		t.Fatal("At should be set")
	}

	snap[0].Reason = "mutated"
	snap2 := DeadLettersSnapshot()
	if snap2[0].Reason != "test reason" {
		t.Fatal("internal DLQ was mutated via snapshot slice")
	}
}

func TestDrainEventQueue(t *testing.T) {
	DrainEventQueue()

	e1 := domain.Event{UserID: "a", EventType: "t", IP: "1.1.1.1", Country: "BR", Device: "x"}

	EventQueue <- e1
	DrainEventQueue()

	select {
	case EventQueue <- e1:
	default:
		t.Fatal("expected drained channel to accept send")
	}
	
	DrainEventQueue()
}

func TestResetDLQ(t *testing.T) {
	PushDeadLetter(domain.Event{UserID: "x", EventType: "t", IP: "1.1.1.1", Country: "BR", Device: "d"}, "r")
	ResetDLQ()
	if len(DeadLettersSnapshot()) != 0 {
		t.Fatal("DLQ not empty after ResetDLQ")
	}
}
