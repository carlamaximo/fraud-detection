package metrics

import (
	"sync"
	"testing"
)

func TestReset(t *testing.T) {
	IncEventsReceived()
	IncFailed()
	IncSentToDLQ()
	RecordSuccessfulProcess("APPROVE")

	Reset()

	s := GetSnapshot()
	if s.TotalEventsReceived != 0 || s.TotalProcessed != 0 || s.TotalFailed != 0 ||
		s.TotalSentToDLQ != 0 || s.TotalApproved != 0 {
		t.Fatalf("Reset did not zero snapshot: %+v", s)
	}
}

func TestRecordSuccessfulProcess_Outcomes(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	RecordSuccessfulProcess("APPROVE")
	RecordSuccessfulProcess("REVIEW")
	RecordSuccessfulProcess("BLOCK")
	RecordSuccessfulProcess("UNKNOWN")

	s := GetSnapshot()
	if s.TotalProcessed != 4 {
		t.Fatalf("TotalProcessed = %d, want 4", s.TotalProcessed)
	}
	
	if s.TotalApproved != 1 || s.TotalReview != 1 || s.TotalBlocked != 1 {
		t.Fatalf("unexpected outcome counts: %+v", s)
	}
}

func TestConcurrentIncrements(t *testing.T) {
	const goroutines = 64
	const ops = 100

	Reset()
	t.Cleanup(Reset)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				IncEventsReceived()
				IncFailed()
			}
		}()
	}
	wg.Wait()

	s := GetSnapshot()
	want := uint64(goroutines * ops)
	if s.TotalEventsReceived != want || s.TotalFailed != want {
		t.Fatalf("got %+v, want received=%d failed=%d", s, want, want)
	}
}
