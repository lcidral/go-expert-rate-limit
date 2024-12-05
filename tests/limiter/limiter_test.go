package limiter

import (
	limiter2 "go-expert-rater-limit/limiter"
	"testing"
	"time"
)

type MockStorage struct {
	requests map[string]int
	blocked  map[string]bool
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		requests: make(map[string]int),
		blocked:  make(map[string]bool),
	}
}

func (m *MockStorage) Get(key string) (int, error) {
	return m.requests[key], nil
}

func (m *MockStorage) Set(key string, value int, _ time.Duration) error {
	m.requests[key] = value
	return nil
}

func (m *MockStorage) Incr(key string) error {
	m.requests[key]++
	return nil
}

func (m *MockStorage) IsBlocked(key string) bool {
	return m.blocked[key]
}

func (m *MockStorage) Block(key string, _ time.Duration) error {
	m.blocked[key] = true
	return nil
}

func TestRateLimiter(t *testing.T) {
	mockStorage := NewMockStorage()
	limiter := limiter2.NewRateLimiter(mockStorage)

	tests := []struct {
		name      string
		key       string
		limit     int
		duration  time.Duration
		blockTime time.Duration
		times     int
		want      bool
	}{
		{
			name:      "Dentro do limite",
			key:       "test-1",
			limit:     5,
			duration:  time.Second,
			blockTime: time.Minute,
			times:     3,
			want:      true,
		},
		{
			name:      "No limite",
			key:       "test-2",
			limit:     5,
			duration:  time.Second,
			blockTime: time.Minute,
			times:     5,
			want:      true,
		},
		{
			name:      "Limite excedido",
			key:       "test-3",
			limit:     5,
			duration:  time.Second,
			blockTime: time.Minute,
			times:     6,
			want:      false,
		},
		{
			name:      "Bloqueado",
			key:       "test-4",
			limit:     5,
			duration:  time.Second,
			blockTime: time.Minute,
			times:     7,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var allowed bool
			for i := 0; i < tt.times; i++ {
				allowed = limiter.IsAllowed(tt.key, tt.limit, tt.duration, tt.blockTime)
			}
			if allowed != tt.want {
				t.Errorf("IsAllowed() = %v, want %v", allowed, tt.want)
			}
		})
	}
}
