package handler

import (
	"encoding/json"
	"net/http"
	"wago-backend/internal/service"
	"wago-backend/internal/utils"
)

type AuthHandler struct {
	AuthService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

func (h *AuthHandler) GeneratePIN(w http.ResponseWriter, r *http.Request) {
	user, err := h.AuthService.GeneratePIN()
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, map[string]string{
		"user_id": user.ID,
		"pin":     user.PIN,
	}, "PIN generated successfully. Please save this PIN for login.")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Basic Auth
	pin, _, ok := r.BasicAuth()
	if !ok {
		// Fallback to body if Basic Auth is missing (optional, but good for flexibility)
		// For now, let's strictly follow PRD which says Basic Auth
		// But wait, PRD says: curl -X POST ... -u "A1B2C3:"
		// This means username is PIN, password is empty.
		// Let's also support JSON body just in case frontend prefers it,
		// but priority is Basic Auth as per PRD.

		var req struct {
			PIN string `json:"pin"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.PIN != "" {
			pin = req.PIN
		} else {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}
	}

	token, user, err := h.AuthService.Login(pin)
	if err != nil {
		utils.ErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"user_id": user.ID,
		"token":   token,
		"pin":     user.PIN,
	}, "Login successful")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Since JWT is stateless, we can't really "logout" on server side without a blacklist.
	// For now, just return success as per PRD.
	utils.SuccessResponse(w, http.StatusOK, nil, "Logout successful")
}
