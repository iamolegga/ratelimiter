package ratelimiter_test

import (
	"context"
	"github.com/iamolegga/ratelimiter"
	"testing"
	"time"
)

func TestRateLimiter_IncrementSimple(t *testing.T) {
	store := ratelimiter.NewInMemoryStore(context.Background(), 3)
	rl := ratelimiter.New(3, 3, 10, store)

	for i := 0; i < 10; i++ {
		ok, err := rl.Increment("test")
		if !ok {
			t.Errorf("should be ok on iteration %d", i)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	ok, err := rl.Increment("test")
	if ok {
		t.Error("should not be ok when limit is reached")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
func TestRateLimiter_IncrementWithTime(t *testing.T) {
	store := ratelimiter.NewInMemoryStore(context.Background(), 3)
	rl := ratelimiter.New(3, 3, 3, store)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		ok, err := rl.Increment("test")
		if !ok {
			t.Errorf("should be ok on iteration %d", i)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	ok, err := rl.Increment("test")
	if ok {
		t.Error("should not be ok when limit is reached")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
