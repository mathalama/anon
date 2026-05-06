package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// ALWAYS remove X-User-ID from incoming request to prevent spoofing
			r.Header.Del("X-User-ID")

			// Skip for health checks and docs
			if strings.HasSuffix(path, "/health") || strings.Contains(path, "/docs") {
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
				tokenString = r.URL.Query().Get("token")
			}

			userID := ""

			if tokenString != "" && tokenString != "undefined" && tokenString != "null" {
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					// Enforce expected algorithm to avoid "alg" confusion attacks.
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %T", token.Method)
					}
					return []byte(secret), nil
				})

				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if sub, ok := claims["sub"].(string); ok {
							userID = sub
						}
					}
				}
			}

			// If no valid userID from token, generate a UUID
			if userID == "" {
				// Generate a valid RFC4122 v4 UUID using crypto/rand
				b := make([]byte, 16)
				if _, err := rand.Read(b); err != nil {
					http.Error(w, "failed to generate id", http.StatusInternalServerError)
					return
				}
				b[6] = (b[6] & 0x0f) | 0x40 // Version 4
				b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
				userID = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
			}

			r.Header.Set("X-User-ID", userID)
			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
