package main

import (
	"cloudcord/user/db"
	"cloudcord/user/logic"
	"cloudcord/user/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "200 Users! Current time is: %s", time.Now())

	log.Printf("Request received: Method: %s, Path: %s, Headers: %v\n", r.Method, r.URL.Path, r.Header)

	w.WriteHeader(http.StatusOK)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var userRequest struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userLogic := logic.NewUserLogic(db.NewRepository(db.DB)) // Using the repository injected into logic
	userLogic.CreateUserHandler(userRequest.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"message": "User created successfully", "username": userRequest.Username}
	json.NewEncoder(w).Encode(response)
}

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

func main() {
	db.Connect()

	err := models.MigrateUsers(db.DB)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	log.Println("Database migrated successfully")

	http.HandleFunc("/", handleOK)               // 200 Ok
	http.HandleFunc("/create", handleCreateUser) // Endpoint to create a user
	http.HandleFunc("/user", handleGetUserByID)  // GetUserByUserId endpoint
	http.HandleFunc("/users", handleGetAllUsers)

	fmt.Println("Starting server on :8081...")

	http.ListenAndServe(":8081", nil)
}
