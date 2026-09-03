package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dbplatform/api/internal/auth"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/api/response"
	"github.com/dbplatform/api/internal/user"
)

type contextKey string

const claimsKey contextKey = "claims"

func Authenticate(jwtSvc *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, r, apperrors.NewUnauthorized())
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, r, apperrors.NewUnauthorized())
				return
			}

			claims, err := jwtSvc.ValidateAccessToken(parts[1])
			if err != nil {
				response.Error(w, r, apperrors.NewUnauthorized())
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.Claims)
	return claims, ok
}

func RequireRole(roles ...user.Role) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[string(r)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok {
				response.Error(w, r, apperrors.NewUnauthorized())
				return
			}

			if !allowed[claims.Role] {
				response.Error(w, r, apperrors.NewForbidden())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
