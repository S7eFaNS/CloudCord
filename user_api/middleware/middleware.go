package middleware

import (
	"cloudcord/user/db"
	"cloudcord/user/models"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

type ContextKey string

const UserContextKey = ContextKey("user")

var (
	auth0Domain = "https://dev-p3oldabcwb4l1kia.us.auth0.com/"
	audience    = "https://cloudcord/api"
	jwksURL     = auth0Domain + ".well-known/jwks.json"
	jwks        *keyfunc.JWKS
	repo        *db.Repository
)

func InitMiddleware(r *db.Repository) {
	repo = r

	var err error
	jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			fmt.Printf("Error refreshing JWKS: %v\n", err)
		},
		RefreshUnknownKID: true,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWKS from URL: %v", err))
	}
}

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		if claims["iss"] != auth0Domain {
			http.Error(w, "Invalid token issuer", http.StatusUnauthorized)
			return
		}

		audClaim := claims["aud"]
		validAud := false
		switch aud := audClaim.(type) {
		case string:
			if aud == audience {
				validAud = true
			}
		case []interface{}:
			for _, a := range aud {
				if s, ok := a.(string); ok && s == audience {
					validAud = true
					break
				}
			}
		}
		if !validAud {
			http.Error(w, "Invalid token audience", http.StatusUnauthorized)
			return
		}

		auth0ID, ok := claims["sub"].(string)
		if !ok || auth0ID == "" {
			http.Error(w, "Invalid token: missing sub claim", http.StatusUnauthorized)
			return
		}

		nickname, _ := claims["nickname"].(string)

		user, err := repo.GetUserByAuth0ID(auth0ID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			newUser := &models.User{
				Auth0ID:  auth0ID,
				Username: nickname,
			}
			err = repo.CreateUser(newUser)
			if err != nil {
				http.Error(w, "Error creating user", http.StatusInternalServerError)
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
