package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"wago-backend/internal/config"
	"wago-backend/internal/service"
	"wago-backend/internal/utils"
	"wago-backend/internal/websocket"

	"github.com/gorilla/mux"
)

type SessionHandler struct {
	SessionService *service.SessionService
	WSHub          *websocket.Hub
	Config         *config.Config
}

func NewSessionHandler(sessionService *service.SessionService, wsHub *websocket.Hub, cfg *config.Config) *SessionHandler {
	return &SessionHandler{
		SessionService: sessionService,
		WSHub:          wsHub,
		Config:         cfg,
	}
}

func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		SessionName string `json:"session_name"`
		WebhookURL  string `json:"webhook_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.SessionName) == "" || len(req.SessionName) > 100 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session name")
		return
	}

	if _, err := url.ParseRequestURI(req.WebhookURL); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid webhook URL")
		return
	}

	session, err := h.SessionService.CreateSession(userID, req.SessionName, req.WebhookURL)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, session, "Session created successfully")
}

func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	sessions, err := h.SessionService.GetSessions(userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, sessions, "Sessions retrieved successfully")
}

func (h *SessionHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if strings.TrimSpace(id) == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid session id")
		return
	}

	err := h.SessionService.StartSession(id)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]string{
		"session_id": id,
		"status":     "qr", // Assuming it goes to QR or connected
	}, "Session started")
}

func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.SessionService.DeleteSession(id, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, nil, "Session deleted successfully")
}

func (h *SessionHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Validate token from query param since WS doesn't support headers easily in browser JS
	token := r.URL.Query().Get("token")
	if token == "" {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Missing token")
		return
	}

	userID, err := utils.ParseUserIDFromToken(token, h.Config.JWTSecret)
	if err != nil {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Ensure session belongs to user
	session, err := h.SessionService.GetSession(id)
	if err != nil || session == nil || session.UserID != userID {
		utils.ErrorResponse(w, http.StatusForbidden, "Session not accessible")
		return
	}

	websocket.ServeWs(h.WSHub, w, r, id, h.Config.AllowedOrigins)
}
