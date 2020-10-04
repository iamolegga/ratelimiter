package ratelimiter

import (
	"math"
	"time"
)

type Window struct {
	FullUntil uint64
	Buckets   map[uint64]uint64
}

func (w *Window) Increment(
	windowDuration, bucketDuration, limit uint64,
) (
	ok, changed bool,
) {
	now := uint64(time.Now().Unix())
	if now < w.FullUntil {
		return
	}

	boundary := now - windowDuration
	currentSum := uint64(0)
	oldestBucketTS := uint64(math.MaxUint64)

	for bucketTS, bucketSum := range w.Buckets {
		if bucketTS <= boundary {
			delete(w.Buckets, bucketTS)
		} else {
			if bucketTS < oldestBucketTS {
				oldestBucketTS = bucketTS
			}
			currentSum += bucketSum
		}
	}

	if currentSum < limit {
		w.FullUntil = 0
		bucketTS := now - (now % bucketDuration)
		w.Buckets[bucketTS]++
		ok = true
		changed = true
		return
	}

	w.FullUntil = oldestBucketTS + windowDuration
	changed = true
	return
}
