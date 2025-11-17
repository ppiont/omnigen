package concurrency

import (
	"context"
	"sync"
)

// Semaphore provides a weighted semaphore for limiting concurrent goroutines
type Semaphore struct {
	weights chan struct{}
	mu      sync.Mutex
}

// NewSemaphore creates a new semaphore with the given capacity
func NewSemaphore(maxConcurrent int) *Semaphore {
	return &Semaphore{
		weights: make(chan struct{}, maxConcurrent),
	}
}

// Acquire acquires a semaphore slot, blocking until one is available or context is canceled
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.weights <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a semaphore slot
func (s *Semaphore) Release() {
	<-s.weights
}

// TryAcquire attempts to acquire a slot without blocking
// Returns true if successful, false if all slots are in use
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.weights <- struct{}{}:
		return true
	default:
		return false
	}
}

// Available returns the number of available slots
func (s *Semaphore) Available() int {
	return cap(s.weights) - len(s.weights)
}
