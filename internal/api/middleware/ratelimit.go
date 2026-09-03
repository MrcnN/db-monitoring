package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/api/response"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	b        int
}

func NewRateLimiter(r float64, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		r:        rate.Limit(r),
		b:        b,
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			for ip, lim := range rl.limiters {
				if time.Since(lim.lastSeen) > 3*time.Minute {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v, exists := rl.limiters[ip]; exists {
		v.lastSeen = time.Now()
		return v.limiter
	}

	lim := rate.NewLimiter(rl.r, rl.b)
	rl.limiters[ip] = &ipLimiter{limiter: lim, lastSeen: time.Now()}
	return lim
}

func (rl *RateLimiter) Limit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}

			if !rl.getLimiter(ip).Allow() {
				w.Header().Set("Retry-After", "1")
				response.Error(w, r, &apperrors.AppError{
					Code:       apperrors.CodeRateLimit,
					Message:    "Too many requests. Please try again later.",
					HTTPStatus: http.StatusTooManyRequests,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
