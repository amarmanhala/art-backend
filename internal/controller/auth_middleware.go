package controller

import (
	"context"
	"net/http"
	"strings"

	"art-backend/internal/response"
	"art-backend/internal/service"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func RequireAuth(authService *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", "missing bearer token")
			return
		}

		userID, ok := authService.GetUserIDByToken(token)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", "invalid bearer token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func CurrentUserID(r *http.Request) int64 {
	userID, _ := r.Context().Value(userIDContextKey).(int64)
	return userID
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}

	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return ""
	}

	return strings.TrimSpace(token)
}
