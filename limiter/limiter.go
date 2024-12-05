package limiter

import (
	"go-expert-rater-limit/storage"
	"time"
)

type RateLimiter struct {
	storage storage.Storage
}

func NewRateLimiter(storage storage.Storage) *RateLimiter {
	return &RateLimiter{storage: storage}
}

func (r *RateLimiter) IsAllowed(key string, limit int, duration time.Duration, blockTime time.Duration) bool {
	if r.storage.IsBlocked(key) {
		return false
	}

	current, err := r.storage.Get(key)
	if err != nil {
		return false
	}

	if current >= limit {
		err := r.storage.Block(key, blockTime)
		if err != nil {
			return false
		}
		return false
	}

	if current == 0 {
		return r.storage.Set(key, 1, duration) == nil
	}

	return r.storage.Incr(key) == nil
}

func (r *RateLimiter) Block(key string, duration time.Duration) error {
	return r.storage.Block(key, duration)
}
