package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

//InMemory is default built-in in memory store.
type InMemory struct {
	*sync.Mutex
	cache  map[string]*item
	ttlSec uint64
}

//NewInMemoryStore is a factory to create InMemory instance. As a side effect it starts InMemory garbage collector.
func NewInMemoryStore(ctx context.Context, ttlSec uint64) *InMemory {
	store := &InMemory{
		Mutex:  &sync.Mutex{},
		cache:  make(map[string]*item),
		ttlSec: ttlSec,
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				store.gc()
			}
		}
	}()

	return store
}

func (s *InMemory) GetAndLock(key string) (window *Window, expireSec uint64, err error) {
	s.Mutex.Lock()
	existing, ok := s.cache[key]
	if !ok {
		existing = &item{
			Mutex:       &sync.Mutex{},
			validBefore: time.Now().Add(time.Duration(s.ttlSec) * time.Second),
			value: &Window{
				FullUntil: 0,
				Buckets:   make(map[uint64]uint64),
			},
		}
		s.cache[key] = existing
	}
	s.Mutex.Unlock()

	existing.Lock()
	if existing.expired() {
		existing.validBefore = time.Now().Add(time.Duration(s.ttlSec) * time.Second)
		existing.value = &Window{
			FullUntil: 0,
			Buckets:   make(map[uint64]uint64),
		}
	}

	window = existing.value
	expireSec = uint64(existing.validBefore.Sub(time.Now()).Seconds())
	return
}

func (s *InMemory) SetAndUnlock(key string, window *Window, expireSec uint64) error {
	it, ok := s.cache[key]
	if !ok {
		return errors.New("cannot update not existing item by passed key: " + key)
	}
	if window != nil {
		it.value = window
		it.validBefore = time.Now().Add(time.Duration(expireSec) * time.Second)
	}
	it.Unlock()
	return nil
}

func (s *InMemory) gc() {
	s.Lock()
	for key, it := range s.cache {
		it.Lock()
		if it.expired() {
			delete(s.cache, key)
		}
		it.Unlock()
	}
	s.Unlock()
}

type item struct {
	*sync.Mutex
	value       *Window
	validBefore time.Time
}

func (i *item) expired() bool {
	return i.validBefore.Before(time.Now())
}
