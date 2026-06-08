package controller

import (
	"net/http"

	"art-backend/internal/response"
)

func Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, "server is running", map[string]string{
		"status": "ok",
	})
}
