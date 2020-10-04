package ratelimiter

type Store interface {
	//GetAndLock locks window by key to prevent concurrent access and returns it with left expire duration in seconds
	GetAndLock(key string) (window *Window, expireSec uint64, err error)
	//SetAndUnlock save window value in store and unlocks it. Passing nil as window argument just unlocks key without
	// making any changes (expireSec argument is not used in that case also)
	SetAndUnlock(key string, window *Window, expireSec uint64) error
}
