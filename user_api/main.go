package main

import (
	"cloudcord/user/db"
	"cloudcord/user/logic"
	"cloudcord/user/middleware"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "200 Users! Current time is: %s", time.Now())

	log.Printf("Request received: Method: %s, Path: %s, Headers: %v\n", r.Method, r.URL.Path, r.Header)

	w.WriteHeader(http.StatusOK)
}

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

func handleGetUserByID(repo *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		userLogic := logic.NewUserLogic(repo)
		user, err := userLogic.GetUserByIDHandler(r.Context(), idStr)
		if err != nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleGetAllUsers(repo *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userLogic := logic.NewUserLogic(repo)
		users, err := userLogic.GetAllUsersHandler(r.Context())
		if err != nil {
			http.Error(w, "Could not retrieve users: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASS")

	if user == "" || pass == "" {
		log.Fatal("MongoDB credentials are not set in environment variables")
	}

	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@messages.vbkzymr.mongodb.net/?retryWrites=true&w=majority&appName=Messages",
		user,
		pass,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	log.Println("✅ Successfully connected to MongoDB Atlas")

	mongoDB := client.Database("Messages")

	repo := db.NewRepository(mongoDB)
	middleware.InitMiddleware(repo)

	http.Handle("/", middleware.ValidateJWT(http.HandlerFunc(handleOK)))
	http.Handle("/create", withCORS(middleware.ValidateJWT(http.HandlerFunc(handleCreateUser))))
	http.Handle("/user", middleware.ValidateJWT(handleGetUserByID(repo)))
	http.Handle("/users", middleware.ValidateJWT(handleGetAllUsers(repo)))

	fmt.Println("Starting server on :8081...")

	http.ListenAndServe(":8081", nil)
}
