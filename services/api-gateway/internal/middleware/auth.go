package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a typed key for context values to prevent collisions.
type contextKey string

// UserIDKey is the context key for the authenticated user ID.
const UserIDKey contextKey = "user_id"

var publicPaths = []string{
	"/api/v1/users/anonymous",
	"/api/v1/users/refresh",
	"/metrics",
}

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// ALWAYS remove X-User-ID from incoming request to prevent spoofing
			r.Header.Del("X-User-ID")

			if r.Method == http.MethodOptions || isPublicPath(path) {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := ""
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
			if tokenString == "" {
				if cookie, err := r.Cookie("access_token"); err == nil {
					tokenString = cookie.Value
				}
			}
			if tokenString == "" {
				tokenString = r.URL.Query().Get("token")
			}

			if tokenString == "" || tokenString == "undefined" || tokenString == "null" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := validateToken(tokenString, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-ID", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicPath(path string) bool {
	return strings.HasSuffix(path, "/health") ||
		strings.Contains(path, "/docs") ||
		slices.Contains(publicPaths, path)
}

func validateToken(tokenString, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", jwt.ErrTokenInvalidClaims
	}

	if typ, _ := claims["typ"].(string); typ != "" && typ != "access" {
		return "", jwt.ErrTokenInvalidClaims
	}

	return sub, nil
}
