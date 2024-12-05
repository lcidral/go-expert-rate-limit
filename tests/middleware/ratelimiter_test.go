package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-expert-rater-limit/limiter"
	"go-expert-rater-limit/middleware"
)

type MockStorage struct {
	requests    map[string]int
	blocked     map[string]bool
	blockedTime map[string]time.Time
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		requests:    make(map[string]int),
		blocked:     make(map[string]bool),
		blockedTime: make(map[string]time.Time),
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

func (m *MockStorage) Block(key string, duration time.Duration) error {
	m.blocked[key] = true
	m.blockedTime[key] = time.Now().Add(duration)
	return nil
}

func TestRateLimiterMiddleware(t *testing.T) {
	storage := NewMockStorage()
	rateLimiter := limiter.NewRateLimiter(storage)

	middleware := middleware.NewRateLimiterMiddleware(
		rateLimiter,
		5,  // IP limit
		10, // Token limit
		time.Second,
		5*time.Minute, // IP block time
		6*time.Minute, // Token block time
	)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		executeCount   int
		expectedStatus int
	}{
		{
			name: "IP within limit",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				return req
			},
			executeCount:   3,
			expectedStatus: http.StatusOK,
		},
		{
			name: "IP exceeds limit",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.2:12345"
				return req
			},
			executeCount:   6,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "Token within limit",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("API_KEY", "test-token")
				return req
			},
			executeCount:   8,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Token exceeds limit",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("API_KEY", "test-token-2")
				return req
			},
			executeCount:   11,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "Token takes precedence over IP",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.3:12345"
				req.Header.Set("API_KEY", "test-token-3")
				return req
			},
			executeCount:   8, // Dentro do limite de token (10) mas acima do IP (5)
			expectedStatus: http.StatusOK,
		},
		{
			name: "Different IPs should have separate limits",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.4:12345"
				return req
			},
			executeCount:   3,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Different tokens should have separate limits",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("API_KEY", "test-token-4")
				return req
			},
			executeCount:   8,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastStatus int

			for i := 0; i < tt.executeCount; i++ {
				req := tt.setupRequest()
				rr := httptest.NewRecorder()

				handler := middleware.Handle(nextHandler)
				handler.ServeHTTP(rr, req)

				lastStatus = rr.Code
			}

			if lastStatus != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					lastStatus, tt.expectedStatus)
			}
		})
	}
}

func TestRateLimiterMiddlewareHeaders(t *testing.T) {
	storage := NewMockStorage()
	rateLimiter := limiter.NewRateLimiter(storage)
	middleware := middleware.NewRateLimiterMiddleware(
		rateLimiter,
		5,
		10,
		time.Second,
		5*time.Minute,
		6*time.Minute,
	)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedKey  string
		executeCount int
	}{
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedKey:  "ip:10.0.0.1",
			executeCount: 6,
		},
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.2, 10.0.0.3",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedKey:  "ip:10.0.0.2",
			executeCount: 6,
		},
		{
			name:         "Remote Addr only",
			headers:      map[string]string{},
			remoteAddr:   "192.168.1.1:12345",
			expectedKey:  "ip:192.168.1.1",
			executeCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			var lastStatus int
			for i := 0; i < tt.executeCount; i++ {
				rr := httptest.NewRecorder()
				handler := middleware.Handle(nextHandler)
				handler.ServeHTTP(rr, req)
				lastStatus = rr.Code
			}

			if tt.executeCount > 5 && lastStatus != http.StatusTooManyRequests {
				t.Errorf("Expected status 429 after %d requests, got %d", tt.executeCount, lastStatus)
			}
		})
	}
}
