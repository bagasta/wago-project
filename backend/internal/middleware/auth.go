package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"wago-backend/internal/config"
	"wago-backend/internal/repository"
	"wago-backend/internal/utils"

	"sync"
	"time"
)

type Middleware struct {
	Config       *config.Config
	UserRepo     *repository.UserRepository
	rateLimiters sync.Map
}

func NewMiddleware(cfg *config.Config, userRepo *repository.UserRepository) *Middleware {
	return &Middleware{
		Config:   cfg,
		UserRepo: userRepo,
	}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.parseToken(r.Header.Get("Authorization"))
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TokenOrPINMiddleware allows Authorization via JWT Bearer token or PIN (Authorization: Pin <pin> or X-Pin header).
func (m *Middleware) TokenOrPINMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try Bearer token first to stay backward compatible
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			userID, err := m.parseTokenOrPin(authHeader)
			if err == nil {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Fallback: X-Pin header
		if pin := strings.TrimSpace(r.Header.Get("X-Pin")); pin != "" {
			userID, err := m.userIDFromPIN(pin)
			if err == nil {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		utils.ErrorResponse(w, http.StatusUnauthorized, "Missing or invalid credentials")
	})
}

func (m *Middleware) parseToken(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization format")
	}
	return utils.ParseUserIDFromToken(parts[1], m.Config.JWTSecret)
}

func (m *Middleware) parseTokenOrPin(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", errors.New("invalid authorization format")
	}

	switch parts[0] {
	case "Bearer":
		return utils.ParseUserIDFromToken(parts[1], m.Config.JWTSecret)
	case "Pin", "PIN", "pin":
		return m.userIDFromPIN(parts[1])
	default:
		return "", errors.New("invalid authorization format")
	}
}

func (m *Middleware) userIDFromPIN(pin string) (string, error) {
	if m.UserRepo == nil {
		return "", errors.New("user repository not configured")
	}

	user, err := m.UserRepo.GetUserByPIN(strings.TrimSpace(pin))
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}
	return user.ID, nil
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
