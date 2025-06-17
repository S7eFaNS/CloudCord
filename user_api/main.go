package main

import (
	"cloudcord/user_api/db"
	"cloudcord/user_api/graphdb"
	"cloudcord/user_api/logic"
	"cloudcord/user_api/middleware"
	"cloudcord/user_api/models"
	"cloudcord/user_api/mq"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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

func handleGetUserByAuth0ID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	auth0ID := r.URL.Query().Get("auth0_id")
	if auth0ID == "" {
		http.Error(w, "auth0_id is required", http.StatusBadRequest)
		return
	}

	userLogic := logic.NewUserLogic(db.NewRepository(db.DB))

	user, err := userLogic.GetUserByAuth0ID(auth0ID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":  "User retrieved successfully",
		"userID":   user.UserID,
		"username": user.Username,
		"auth0_id": user.Auth0ID,
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

func handleDeleteUser(userLogic *logic.UserLogic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		auth0ID := r.URL.Query().Get("auth0_id")
		if auth0ID == "" {
			http.Error(w, "auth0_id is required", http.StatusBadRequest)
			return
		}

		err := userLogic.DeleteUserByAuth0ID(auth0ID)
		if err != nil {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		err = middleware.DeleteUserFromAuth0(auth0ID)
		if err != nil {
			fmt.Printf("Failed to delete user from Auth0: %v\n", err)
			http.Error(w, "Failed to delete user from Auth0", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User deleted successfully",
		})
	}
}

func handleAddFriend(userLogic *logic.UserLogic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		type AddFriendRequest struct {
			UserID   uint `json:"user_id"`
			FriendID uint `json:"friend_id"`
		}

		var req AddFriendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.UserID == 0 || req.FriendID == 0 {
			http.Error(w, "Missing user IDs", http.StatusBadRequest)
			return
		}

		err := userLogic.AddFriend(req.UserID, req.FriendID)
		if err != nil {
			http.Error(w, "Failed to add friend", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Friend added successfully"})
	}
}

func handleAreFriends(userLogic *logic.UserLogic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		otherIDStr := r.URL.Query().Get("other_id")

		userID, err1 := strconv.ParseUint(userIDStr, 10, 32)
		otherID, err2 := strconv.ParseUint(otherIDStr, 10, 32)

		if err1 != nil || err2 != nil || userID == 0 || otherID == 0 {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		areFriends, err := userLogic.AreFriends(uint(userID), uint(otherID))
		if err != nil {
			http.Error(w, "Error checking friendship", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]bool{"are_friends": areFriends})
	}
}

func handleFriendRecommendations(userLogic *logic.UserLogic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.URL.Query().Get("user_id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil || userID == 0 {
			http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
			return
		}

		recommendations, err := userLogic.GetFriendRecommendations(uint(userID))
		if err != nil {
			http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recommendations)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:3000" || origin == "https://cloudcord.com" || origin == "https://cloudcord.info" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE,OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db.Connect()

	err := graphdb.Connect()
	if err != nil {
		log.Fatalf("failed to connect to Neo4j: %v", err)
	}
	defer graphdb.Close()

	repo := db.NewRepository(db.DB)

	middleware.InitMiddleware(repo)

	err = models.MigrateAll(db.DB)
	if err != nil {
		log.Fatal("could not migrate db")
	}
	log.Println("Database migrated successfully")

	var publisher *mq.Publisher
	var err2 error
	rabbitURI := os.Getenv("RABBITMQ_URI")
	if rabbitURI == "" {
		log.Fatal("RabbitMQ path not set in environment")
	}

	for i := 0; i < 10; i++ {
		publisher, err2 = mq.NewPublisher(rabbitURI, "user_deletion")
		if err2 == nil {
			log.Println("✅ RabbitMQ publisher set up successfully")
			break
		}
		log.Printf("Attempt %d: Failed to set up RabbitMQ publisher: %v", i+1, err2)
		time.Sleep(3 * time.Second)
	}

	if err2 != nil {
		log.Fatalf("Failed to set up RabbitMQ publisher after retries: %v", err2)
	}

	userLogic := logic.NewUserLogicRabbitMQ(repo, publisher)

	http.HandleFunc("/", handleOK)

	http.Handle("/user/create", withCORS(middleware.ValidateJWT(http.HandlerFunc(handleCreateUser))))
	http.Handle("/user/user", middleware.ValidateJWT(http.HandlerFunc(handleGetUserByID)))
	http.Handle("/user/auth-user", middleware.ValidateJWT(http.HandlerFunc(handleGetUserByAuth0ID)))
	http.Handle("/user/users", withCORS(middleware.ValidateJWT(http.HandlerFunc(handleGetAllUsers))))
	http.Handle("/user/delete", withCORS(middleware.ValidateJWT(handleDeleteUser(userLogic))))
	http.Handle("/user/add-friend", withCORS(middleware.ValidateJWT(handleAddFriend(userLogic))))
	http.Handle("/user/is-friend", withCORS(middleware.ValidateJWT(handleAreFriends(userLogic))))
	http.Handle("/user/recommendations", withCORS(middleware.ValidateJWT(handleFriendRecommendations(userLogic))))

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
