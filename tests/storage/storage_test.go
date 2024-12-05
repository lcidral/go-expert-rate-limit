package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go-expert-rater-limit/storage"
)

// Redis configuration for tests
func setupRedis(t *testing.T) (*redis.Client, func()) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	// Test connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("Error connecting to Redis: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		redisClient.FlushDB(context.Background())
		_ = redisClient.Close()
	}

	return redisClient, cleanup
}

func TestRedisStorage(t *testing.T) {
	redisClient, cleanup := setupRedis(t)
	defer cleanup()

	store := storage.NewRedisStorage(redisClient)

	t.Run("Get nonexistent key", func(t *testing.T) {
		val, err := store.Get("nonexistent")
		assert.NoError(t, err)
		assert.Equal(t, 0, val)
	})

	t.Run("Set and Get normal value", func(t *testing.T) {
		err := store.Set("test1", 42, time.Minute)
		assert.NoError(t, err)

		val, err := store.Get("test1")
		assert.NoError(t, err)
		assert.Equal(t, 42, val)
	})

	t.Run("Increment operations", func(t *testing.T) {
		key := "counter"

		// First increment
		err := store.Incr(key)
		assert.NoError(t, err)

		val, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 1, val)

		// Second increment
		err = store.Incr(key)
		assert.NoError(t, err)

		val, err = store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 2, val)
	})

	t.Run("Block and IsBlocked", func(t *testing.T) {
		key := "blocked_key"

		// Check not blocked initially
		assert.False(t, store.IsBlocked(key))

		// Block
		err := store.Block(key, time.Minute)
		assert.NoError(t, err)

		// Verify blocked
		assert.True(t, store.IsBlocked(key))

		// Wait for expiration (using shorter time for test)
		err = store.Block(key, time.Millisecond)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 2)

		// Verify no longer blocked
		assert.False(t, store.IsBlocked(key))
	})

	t.Run("expired key", func(t *testing.T) {
		key := "expire_test"

		err := store.Set(key, 1, time.Millisecond)
		assert.NoError(t, err)

		// Wait for expiration
		time.Sleep(time.Millisecond * 2)

		val, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 0, val)
	})

	t.Run("concurrent operations", func(t *testing.T) {
		key := "concurrent"
		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func() {
				err := store.Incr(key)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all goroutines to finish
		for i := 0; i < 10; i++ {
			<-done
		}

		val, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 10, val)
	})
}
