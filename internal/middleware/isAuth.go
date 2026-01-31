package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

type Claims struct {
	UserID int    `json:"id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const ClaimsKey contextKey = "claims"

var secretKey = []byte("your-256-bit-secret")

func verifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	return claims, ok
}

func IsAuth(jwtSecret string) func(http.Handler) http.Handler {
	secretKey = []byte(jwtSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: Missing or invalid token format", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := verifyToken(tokenString)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
