package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"wago-backend/internal/config"
	"wago-backend/internal/database"
	"wago-backend/internal/handler"
	"wago-backend/internal/middleware"
	"wago-backend/internal/repository"
	"wago-backend/internal/service"
	"wago-backend/internal/webhook"
	"wago-backend/internal/websocket"
	"wago-backend/internal/whatsapp"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to Database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run Migrations
	// Get current file path to locate migrations folder
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../..")
	migrationsDir := filepath.Join(root, "migrations")

	if err := database.RunMigrations(migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize Repositories
	userRepo := repository.NewUserRepository(database.DB)
	sessionRepo := repository.NewSessionRepository(database.DB)
	analyticsRepo := repository.NewAnalyticsRepository(database.DB)

	// Initialize Services
	authService := service.NewAuthService(userRepo, cfg)
	webhookService := webhook.NewWebhookService() // Changed this line

	// Initialize WhatsApp Client Manager
	clientMgr := whatsapp.NewClientManager(cfg, sessionRepo, analyticsRepo, wsHub, webhookService)

	// Reconnect existing sessions
	clientMgr.ReconnectAllSessions()

	sessionService := service.NewSessionService(sessionRepo, clientMgr)

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	sessionHandler := handler.NewSessionHandler(sessionService, wsHub, cfg)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsRepo)

	// Initialize Middleware
	mw := middleware.NewMiddleware(cfg, userRepo)

	r := mux.NewRouter()
	r.Use(mw.CORS)
	r.Use(mw.RateLimitMiddleware)

	// API Routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Auth Routes
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/generate-pin", authHandler.GeneratePIN).Methods("POST")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	auth.HandleFunc("/logout", authHandler.Logout).Methods("POST")

	// NOTE: These routes MUST be registered BEFORE the sessions subrouter
	// so gorilla/mux tries them first before the PathPrefix("/sessions") subrouter intercepts.
	// Send message/image can be authorized with Bearer token or PIN (for simpler integrations like n8n)
	api.Handle("/sessions/{id}/send-message", mw.TokenOrPINMiddleware(http.HandlerFunc(sessionHandler.SendMessage))).Methods("POST")
	api.Handle("/sessions/{id}/send-image", mw.TokenOrPINMiddleware(http.HandlerFunc(sessionHandler.SendImage))).Methods("POST")

	// Session Routes (registered AFTER specific routes, to avoid prefix conflicts)
	sessions := api.PathPrefix("/sessions").Subrouter()
	sessions.Use(mw.AuthMiddleware)
	sessions.HandleFunc("", sessionHandler.CreateSession).Methods("POST")
	sessions.HandleFunc("", sessionHandler.GetSessions).Methods("GET")
	sessions.HandleFunc("/{id}/start", sessionHandler.StartSession).Methods("POST")
	sessions.HandleFunc("/{id}/stop", sessionHandler.StopSession).Methods("POST")
	sessions.HandleFunc("/{id}", sessionHandler.UpdateSession).Methods("PUT")
	sessions.HandleFunc("/{id}", sessionHandler.DeleteSession).Methods("DELETE")
	sessions.HandleFunc("/{id}/analytics", analyticsHandler.GetSessionAnalytics).Methods("GET")
	sessions.HandleFunc("/{id}/contacts", analyticsHandler.GetSessionContacts).Methods("GET")

	// WebSocket Route
	r.HandleFunc("/ws/sessions/{id}", sessionHandler.WebSocketHandler)

	// Protected Routes Example (to be used later)
	// protected := api.PathPrefix("/").Subrouter()
	// protected.Use(mw.AuthMiddleware)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port %s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientMgr.Shutdown()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	database.Close()
	log.Println("Server stopped gracefully")
}
