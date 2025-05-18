package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mmd-moradi/goup/internal/auth"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/mmd-moradi/goup/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Authenticate(tokenSVC *auth.TokenService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				response.Error(w, apperrors.New(apperrors.Unauthorized, "authorization token is required"))
				return
			}

			parts := strings.SplitN(token, " ", 2)

			if !(len(parts) == 2 && parts[0] == "Bearer") {
				response.Error(w, apperrors.New(apperrors.Unauthorized, "authorization token is malformed, expected format: Bearer <token>"))
				return
			}

			userID, err := tokenSVC.ValidateToken(parts[1])
			if err != nil {
				response.Error(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", apperrors.New(apperrors.Unauthorized, "user ID not found in context")
	}
	return userID, nil
}
