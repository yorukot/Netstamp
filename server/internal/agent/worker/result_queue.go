package worker

import (
	"log/slog"
)

type ResultQueue struct {
	ch       chan ResultEnvelope
	log      *slog.Logger
}

func NewResultQueue(capacity int, log *slog.Logger) *ResultQueue {
	return &ResultQueue{
		ch:       make(chan ResultEnvelope, capacity),
		log:      log,
	}
}

func (q *ResultQueue) Enqueue(result ResultEnvelope) {
	select {
	case q.ch <- result:
		return
	default:
	}

	select {
	case <-q.ch:
		q.log.Warn("dropped oldest result because result queue is full", "check_id", result.CheckID, "check_type", result.Type)
	default:
	}

	select {
	case q.ch <- result:
	default:
		q.log.Warn("dropped result because result queue remained full", "check_id", result.CheckID, "check_type", result.Type)
	}
}

func (q *ResultQueue) Channel() <-chan ResultEnvelope {
	return q.ch
}

func (q *ResultQueue) Drain(maxItems int) []ResultEnvelope {
	results := make([]ResultEnvelope, 0, maxItems)
	for len(results) < maxItems {
		select {
		case result := <-q.ch:
			results = append(results, result)
		default:
			return results
		}
	}

	return results
}
