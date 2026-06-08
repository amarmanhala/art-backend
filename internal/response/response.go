package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Error   ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, code string, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Message: message,
		Error: ErrorBody{
			Code:    code,
			Details: details,
		},
	})
}
