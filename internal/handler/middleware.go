package handler

import (
	"context"
	"net/http"
	"strings"

	"hotel-test/internal/domain"
	"hotel-test/internal/service"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				WriteError(w, domain.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				WriteError(w, domain.ErrInvalidToken)
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				WriteError(w, domain.ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) (*service.TokenClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*service.TokenClaims)
	return claims, ok
}
