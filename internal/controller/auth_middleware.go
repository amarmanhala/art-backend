package controller

import (
	"context"
	"net/http"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/response"
	"art-backend/internal/service"
)

type contextKey string

const (
	userIDContextKey contextKey = "user_id"
	userContextKey   contextKey = "user"
)

func RequireAuth(authService *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticateRequest(w, r, authService)
		if !ok {
			return
		}

		next(w, r.WithContext(contextWithUser(r.Context(), user)))
	}
}

func RequireAdmin(authService *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticateRequest(w, r, authService)
		if !ok {
			return
		}
		if user.Role != model.RoleAdmin {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "admin access required", "")
			return
		}

		next(w, r.WithContext(contextWithUser(r.Context(), user)))
	}
}

func authenticateRequest(w http.ResponseWriter, r *http.Request, authService *service.AuthService) (model.User, bool) {
	token := bearerToken(r)
	if token == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", "missing bearer token")
		return model.User{}, false
	}

	user, ok, err := authService.GetUserByToken(r.Context(), token)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not authenticate user", err.Error())
		return model.User{}, false
	}
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", "invalid bearer token")
		return model.User{}, false
	}

	return user, true
}

func contextWithUser(ctx context.Context, user model.User) context.Context {
	ctx = context.WithValue(ctx, userIDContextKey, user.ID)
	return context.WithValue(ctx, userContextKey, user)
}

func CurrentUserID(r *http.Request) int64 {
	userID, _ := r.Context().Value(userIDContextKey).(int64)
	return userID
}

func CurrentUser(r *http.Request) model.User {
	user, _ := r.Context().Value(userContextKey).(model.User)
	return user
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
