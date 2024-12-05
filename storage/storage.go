package storage

import "time"

type Storage interface {
	Get(key string) (int, error)
	Set(key string, value int, expiration time.Duration) error
	Incr(key string) error
	IsBlocked(key string) bool
	Block(key string, duration time.Duration) error
}
