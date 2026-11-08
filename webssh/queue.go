//go:build js && wasm

package main

import (
	"context"
	"sync"
)

// queue is an unbounded FIFO of keystroke chunks. Pushes come from JS event
// callbacks, which must never block, and keep their order.
type queue struct {
	mu     sync.Mutex
	items  [][]byte
	notify chan struct{}
}

func newQueue() *queue { return &queue{notify: make(chan struct{}, 1)} }

func (q *queue) push(b []byte) {
	q.mu.Lock()
	q.items = append(q.items, b)
	q.mu.Unlock()
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

// pop waits for the next chunk. It returns false once ctx is done.
func (q *queue) pop(ctx context.Context) ([]byte, bool) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			b := q.items[0]
			q.items = q.items[1:]
			q.mu.Unlock()
			return b, true
		}
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, false
		case <-q.notify:
		}
	}
}
