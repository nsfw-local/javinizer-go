package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLimiter_ZeroDelay(t *testing.T) {
	limiter := NewLimiter(0)
	start := time.Now()
	err := limiter.Wait(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if elapsed > 10*time.Millisecond {
		t.Errorf("zero delay should return immediately, took %v", elapsed)
	}
}

func TestLimiter_DelayCausesWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}

	delay := 50 * time.Millisecond
	limiter := NewLimiter(delay)

	start := time.Now()
	_ = limiter.Wait(context.Background())
	firstElapsed := time.Since(start)

	start = time.Now()
	_ = limiter.Wait(context.Background())
	secondElapsed := time.Since(start)

	if firstElapsed > 10*time.Millisecond {
		t.Errorf("first call should return immediately, took %v", firstElapsed)
	}
	if secondElapsed < delay-10*time.Millisecond {
		t.Errorf("second call should wait at least %v, only waited %v", delay, secondElapsed)
	}
	if secondElapsed > delay+20*time.Millisecond {
		t.Errorf("second call should not wait much longer than %v, waited %v", delay, secondElapsed)
	}
}

func TestLimiter_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}

	delay := 100 * time.Millisecond
	limiter := NewLimiter(delay)
	_ = limiter.Wait(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("should return quickly on cancellation, took %v", elapsed)
	}
}

func TestLimiter_ConcurrentCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}

	delay := 20 * time.Millisecond
	limiter := NewLimiter(delay)

	var wg sync.WaitGroup
	const numGoroutines = 10
	times := make([]time.Time, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = limiter.Wait(context.Background())
			times[idx] = time.Now()
		}(i)
	}

	wg.Wait()

	for i := 1; i < numGoroutines; i++ {
		diff := times[i].Sub(times[i-1])
		if diff < 0 {
			diff = -diff
		}
		if diff < delay-5*time.Millisecond {
			t.Errorf("concurrent calls should be separated by at least %v, got %v between call %d and %d", delay, diff, i-1, i)
		}
	}
}

func TestLimiter_CancellationUnderContention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}

	delay := 100 * time.Millisecond
	limiter := NewLimiter(delay)

	ctx1 := context.Background()
	err1 := make(chan error, 1)
	go func() {
		err1 <- limiter.Wait(ctx1)
	}()

	time.Sleep(10 * time.Millisecond)

	ctx2, cancel2 := context.WithCancel(context.Background())
	err2 := make(chan error, 1)
	go func() {
		err2 <- limiter.Wait(ctx2)
	}()

	ctx3, cancel3 := context.WithCancel(context.Background())
	err3 := make(chan error, 1)
	go func() {
		err3 <- limiter.Wait(ctx3)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel2()

	select {
	case err := <-err2:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("cancellation should be observed quickly, not blocked by first waiter")
	}

	cancel3()
	select {
	case err := <-err3:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("third waiter cancellation should be observed quickly")
	}

	if err := <-err1; err != nil {
		t.Errorf("first waiter should complete successfully: %v", err)
	}
}
