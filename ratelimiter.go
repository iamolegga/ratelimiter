//Package ratelimiter provides rate limiter that works with sliding window algorithm.
package ratelimiter

//RateLimiter is rate limiter that works with sliding window algorithm.
type RateLimiter struct {
	windowDurationSec uint64
	bucketDuration    uint64
	limit             uint64
	store             Store
}

//New is a factory to create RateLimiter instance. windowDurationSec is the duration of sliding window.  bucketsCount is
//a count of buckets in one window. Bucket is a chunk made for optimization. The more buckets - the less it's size and
//the slower rate limiter summarize counter for window, but it's more accurate. limit is the limit for single window.
//store is the storage of windows
func New(
	windowDurationSec uint64,
	bucketsCount uint64,
	limit uint64,
	store Store,
) *RateLimiter {
	bucketDuration := windowDurationSec / bucketsCount
	if windowDurationSec%bucketsCount != 0 {
		bucketDuration++
	}

	return &RateLimiter{
		windowDurationSec: windowDurationSec,
		bucketDuration:    bucketDuration,
		limit:             limit,
		store:             store,
	}
}

//Increment increments counter for key. If the limit for the key is reached ok will be false, otherwise true. Also it's
//pass underlying store error if got one.
func (rl *RateLimiter) Increment(key string) (ok bool, err error) {
	window, expire, err := rl.store.GetAndLock(key)
	if err != nil {
		return
	}

	ok, changed := window.Increment(rl.windowDurationSec, rl.bucketDuration, rl.limit)
	if changed {
		if ok {
			// if successfully incremented then should be saved
			// with full expiration duration
			err = rl.store.SetAndUnlock(key, window, rl.windowDurationSec)
		} else {
			// if not incremented then should be saved
			// with existing expiration duration
			err = rl.store.SetAndUnlock(key, window, expire)
		}
	} else {
		err = rl.store.SetAndUnlock(key, nil, 0)
	}
	return
}
