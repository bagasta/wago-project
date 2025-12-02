package handler

import (
	"encoding/json"
	"net/http"
	"wago-backend/internal/repository"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	Repo *repository.AnalyticsRepository
}

func NewAnalyticsHandler(repo *repository.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{Repo: repo}
}

func (h *AnalyticsHandler) GetSessionAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	stats, err := h.Repo.GetSessionAnalytics(sessionID)
	if err != nil {
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AnalyticsHandler) GetSessionContacts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	contacts, err := h.Repo.GetUniqueContacts(sessionID)
	if err != nil {
		http.Error(w, "Failed to fetch contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}
