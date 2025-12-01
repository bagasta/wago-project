package middleware

import (
	"context"
	"net/http"
	"strings"
	"wago-backend/internal/config"
	"wago-backend/internal/utils"

	"sync"
	"time"
)

type Middleware struct {
	Config       *config.Config
	rateLimiters sync.Map
}

func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{Config: cfg}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		userID, err := utils.ParseUserIDFromToken(parts[1], m.Config.JWTSecret)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) CORS(next http.Handler) http.Handler {
	allowed := m.Config.AllowedOrigins
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if originAllowed(origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(allowed) == 1 && allowed[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func originAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return true // non-browser clients
	}
	for _, o := range allowed {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}

// simple token bucket per IP
type limiter struct {
	tokens     int
	lastRefill time.Time
}

func (m *Middleware) RateLimitMiddleware(next http.Handler) http.Handler {
	const (
		maxTokens    = 60
		refillPeriod = time.Minute
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]

		val, _ := m.rateLimiters.LoadOrStore(ip, &limiter{tokens: maxTokens, lastRefill: time.Now()})
		lim := val.(*limiter)

		now := time.Now()
		if since := now.Sub(lim.lastRefill); since > refillPeriod {
			lim.tokens = maxTokens
			lim.lastRefill = now
		}

		if lim.tokens <= 0 {
			utils.ErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}
		lim.tokens--

		next.ServeHTTP(w, r)
	})
}
