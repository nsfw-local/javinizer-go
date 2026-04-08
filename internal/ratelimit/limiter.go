package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Limiter struct {
	mu              sync.Mutex
	nextAllowedTime time.Time
	delay           time.Duration
}

func NewLimiter(delay time.Duration) *Limiter {
	return &Limiter{delay: delay}
}

func (l *Limiter) Wait(ctx context.Context) error {
	if l.delay <= 0 {
		return nil
	}

	l.mu.Lock()

	now := time.Now()
	if l.nextAllowedTime.IsZero() || now.After(l.nextAllowedTime) {
		l.nextAllowedTime = now.Add(l.delay)
		l.mu.Unlock()
		return nil
	}

	waitDuration := l.nextAllowedTime.Sub(now)
	l.nextAllowedTime = l.nextAllowedTime.Add(l.delay)
	l.mu.Unlock()

	select {
	case <-time.After(waitDuration):
		return nil
	case <-ctx.Done():
		l.mu.Lock()
		l.nextAllowedTime = l.nextAllowedTime.Add(-l.delay)
		l.mu.Unlock()
		return ctx.Err()
	}
}
