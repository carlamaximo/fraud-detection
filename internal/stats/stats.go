package stats

import "sync"

// Snapshot is JSON for GET /metrics (field names are stable for anyone scraping this).
type Snapshot struct {
	TotalEventsReceived uint64 `json:"total_events_received"`
	TotalProcessed      uint64 `json:"total_processed"`
	TotalFailed         uint64 `json:"total_failed"`
	TotalSentToDLQ      uint64 `json:"total_sent_to_dlq"`
	TotalApproved       uint64 `json:"total_approved"`
	TotalReview         uint64 `json:"total_review"`
	TotalBlocked        uint64 `json:"total_blocked"`
}

var (
	mu sync.Mutex
	s  Snapshot
)

func IncEventsReceived() {
	mu.Lock()
	s.TotalEventsReceived++
	mu.Unlock()
}

func IncFailed() {
	mu.Lock()
	s.TotalFailed++
	mu.Unlock()
}

func IncSentToDLQ() {
	mu.Lock()
	s.TotalSentToDLQ++
	mu.Unlock()
}

func RecordSuccessfulProcess(outcome string) {
	mu.Lock()
	s.TotalProcessed++
	switch outcome {
	case "APPROVE":
		s.TotalApproved++
	case "REVIEW":
		s.TotalReview++
	case "BLOCK":
		s.TotalBlocked++
	}
	mu.Unlock()
}

func GetSnapshot() Snapshot {
	mu.Lock()
	defer mu.Unlock()
	return s
}

func Reset() {
	mu.Lock()
	s = Snapshot{}
	mu.Unlock()
}
