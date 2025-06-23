package handlers

import (
	"cloudcord/user_api/db"
	"cloudcord/user_api/logic"
	"cloudcord/user_api/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

// user create api
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
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
func HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
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

func HandleGetUserByAuth0ID(w http.ResponseWriter, r *http.Request) {
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

func HandleGetAllUsers(w http.ResponseWriter, r *http.Request) {
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

func HandleDeleteUser(userLogic *logic.UserLogic) http.HandlerFunc {
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

func HandleAddFriend(userLogic *logic.UserLogic) http.HandlerFunc {
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

func HandleAreFriends(userLogic *logic.UserLogic) http.HandlerFunc {
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

func HandleFriendRecommendations(userLogic *logic.UserLogic) http.HandlerFunc {
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
