package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func JSONResponse(w http.ResponseWriter, statusCode int, success bool, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: success,
		Data:    data,
		Message: message,
	})
}

func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, false, nil, message)
}

func SuccessResponse(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	JSONResponse(w, statusCode, true, data, message)
}
