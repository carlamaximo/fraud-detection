package queue

import (
	"sync"
	"time"

	"fraud-platform/internal/domain"
)

var EventQueue = make(chan domain.Event, 100)

type DeadLetter struct {
	Event  domain.Event `json:"event"`
	Reason string       `json:"reason"`
	At     time.Time    `json:"at"`
}

var (
	dlqMu   sync.Mutex
	dlq     []DeadLetter
	maxDLQ  = 1000 
)

func PushDeadLetter(event domain.Event, reason string) {
	dlqMu.Lock()
	defer dlqMu.Unlock()
	if len(dlq) >= maxDLQ {
		dlq = dlq[1:]
	}

	dlq = append(dlq, DeadLetter{Event: event, Reason: reason, At: time.Now()})
}

func DeadLettersSnapshot() []DeadLetter {
	dlqMu.Lock()
	defer dlqMu.Unlock()
	out := make([]DeadLetter, len(dlq))
	copy(out, dlq)
	return out
}

func ResetDLQ() {
	dlqMu.Lock()
	dlq = nil
	dlqMu.Unlock()
}

func DrainEventQueue() {
	for {
		select {
		case <-EventQueue:
		default:
			return
		}
	}
}
