package main

import (
	"cloudcord/user_api/db"
	"cloudcord/user_api/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	db.Connect()

	os.Exit(m.Run())
}

func TestGetAllUsersIntegration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handleGetAllUsers)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rr.Code)
	}

	var users []models.User
	if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(users) == 0 {
		t.Logf("No users returned — DB might be empty")
	} else {
		t.Logf("Returned %d users, first: %d", len(users), users[0].UserID)
	}
}
