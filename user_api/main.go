package main

import (
	"cloudcord/user_api/db"
	"cloudcord/user_api/logic"
	"cloudcord/user_api/middleware"
	"cloudcord/user_api/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "200 Users! Current time is: %s", time.Now())

	log.Printf("Request received: Method: %s, Path: %s, Headers: %v\n", r.Method, r.URL.Path, r.Header)

	w.WriteHeader(http.StatusOK)
}

// user create api
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized: no valid token claims found", http.StatusUnauthorized)
		return
	}

	auth0ID, ok := claims["sub"].(string)
	if !ok || auth0ID == "" {
		http.Error(w, "Invalid token: sub claim missing", http.StatusUnauthorized)
		return
	}

	nickname, _ := claims["nickname"].(string)
	if nickname == "" {
		nickname = "unknown"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "User processed successfully",
		"auth0ID":  auth0ID,
		"username": nickname,
	})
}

// user get by userid
func handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	userLogic := logic.NewUserLogic(db.NewRepository(db.DB))

	user, err := userLogic.GetUserByIDHandler(uint(id))

	if err != nil || user == nil {
		http.Error(w, "User ID not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":  "User retrieved successfully",
		"userID":   user.UserID,
		"username": user.Username,
	}
	json.NewEncoder(w).Encode(response)
}

func handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userLogic := logic.NewUserLogic(db.NewRepository(db.DB))

	users, err := userLogic.GetAllUsersHandler()
	if err != nil {
		http.Error(w, "Could not retrieve users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:3000" || origin == "https://cloudcord.com" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db.Connect()

	repo := db.NewRepository(db.DB)

	middleware.InitMiddleware(repo)

	err := models.MigrateUsers(db.DB)
	if err != nil {
		log.Fatal("could not migrate db")
	}
	log.Println("Database migrated successfully")

	http.Handle("/", middleware.ValidateJWT(http.HandlerFunc(handleOK)))
	http.Handle("/create", withCORS(middleware.ValidateJWT(http.HandlerFunc(handleCreateUser))))
	http.Handle("/user", middleware.ValidateJWT(http.HandlerFunc(handleGetUserByID)))
	http.Handle("/users", middleware.ValidateJWT(http.HandlerFunc(handleGetAllUsers)))

	go func() {
		fmt.Println("Starting metrics server on :2112...")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	fmt.Println("Starting server on :8081...")

	http.ListenAndServe(":8081", nil)
}
