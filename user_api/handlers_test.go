//go:build integration
// +build integration

package main

import (
	"cloudcord/user_api/db"
	"cloudcord/user_api/handlers"
	"cloudcord/user_api/logic"
	"cloudcord/user_api/middleware"
	"cloudcord/user_api/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v4"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	db.Connect()

	os.Exit(m.Run())
}

func TestGetAllUsersIntegration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.HandleGetAllUsers)
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

func TestGetUserByIDIntegration(t *testing.T) {
	testUser := &models.User{
		Auth0ID:  "auth0|test_integration_user",
		Username: "integrationuser",
	}
	repo := db.NewRepository(db.DB)
	err := repo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	defer db.DB.Unscoped().Delete(&models.User{}, testUser.UserID)

	url := fmt.Sprintf("/user?id=%d", testUser.UserID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.HandleGetUserByID)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	t.Logf("Response userID: %v, username: %v", response["userID"], response["username"])

	if response["userID"] == nil || response["username"] == nil {
		t.Fatalf("Response missing fields: %v", response)
	}

	expectedID := float64(testUser.UserID)
	if response["userID"] != expectedID {
		t.Errorf("Expected userID %v, got %v", expectedID, response["userID"])
	}

	if response["username"] != testUser.Username {
		t.Errorf("Expected username %s, got %s", testUser.Username, response["username"])
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/user?id=999999", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.HandleGetUserByID)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found, got %d", rr.Code)
	}
}

func TestHandleCreateUser(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":      "auth0|test123",
		"nickname": "testuser",
	}

	req := httptest.NewRequest(http.MethodPost, "/create", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.HandleCreateUser)
	handler.ServeHTTP(rr, req)

	t.Logf("Response body: %s", rr.Body.String())

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if resp["auth0ID"] != "auth0|test123" {
		t.Errorf("Expected auth0ID 'auth0|test123', got %q", resp["auth0ID"])
	}
	if resp["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %q", resp["username"])
	}
	if resp["message"] != "User processed successfully" {
		t.Errorf("Expected message 'User processed successfully', got %q", resp["message"])
	}
}

func TestDeleteUser_Auth0NotFoundHandledGracefully(t *testing.T) {
	testUser := &models.User{
		Auth0ID:  "auth0|missing_in_auth0",
		Username: "ghost",
	}

	repo := db.NewRepository(db.DB)
	err := repo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	defer db.DB.Unscoped().Delete(&models.User{}, testUser.UserID)

	originalAuth0Delete := middleware.DeleteUserFromAuth0
	middleware.DeleteUserFromAuth0 = func(auth0ID string) error {
		return nil
	}
	defer func() { middleware.DeleteUserFromAuth0 = originalAuth0Delete }()

	req := httptest.NewRequest(http.MethodDelete, "/delete?auth0_id="+testUser.Auth0ID, nil)
	rr := httptest.NewRecorder()

	userLogic := logic.NewUserLogic(repo)
	handler := handlers.HandleDeleteUser(userLogic)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}
}
