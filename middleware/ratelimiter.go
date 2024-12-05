package middleware

import (
	"go-expert-rater-limit/limiter"
	"net/http"
	"strings"
	"time"
)

type RateLimiterMiddleware struct {
	limiter        *limiter.RateLimiter
	ipLimit        int
	tokenLimit     int
	ipDuration     time.Duration
	ipBlockTime    time.Duration
	tokenBlockTime time.Duration
}

func NewRateLimiterMiddleware(
	limiter *limiter.RateLimiter,
	ipLimit, tokenLimit int,
	ipDuration, ipBlockTime, tokenBlockTime time.Duration,
) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter:        limiter,
		ipLimit:        ipLimit,
		tokenLimit:     tokenLimit,
		ipDuration:     ipDuration,
		ipBlockTime:    ipBlockTime,
		tokenBlockTime: tokenBlockTime,
	}
}

func (m *RateLimiterMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API_KEY")
		ip := realIP(r)

		if token != "" {
			if !m.limiter.IsAllowed("token:"+token, m.tokenLimit, m.ipDuration, m.tokenBlockTime) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, err := w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				if err != nil {
					return
				}
				return
			}
		} else {
			if !m.limiter.IsAllowed("ip:"+ip, m.ipLimit, m.ipDuration, m.ipBlockTime) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, err := w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				if err != nil {
					return
				}
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request) string {
	// X-Real-IP primeiro
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Tentar X-Forwarded-For
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// Pegar IP direto da requisição
	ip = r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		return ip[:i]
	}

	return ip
}
