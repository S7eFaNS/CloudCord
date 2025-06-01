package main

import (
	"bytes"
	"cloudcord/chat_api/db"
	"cloudcord/chat_api/logic"
	"cloudcord/chat_api/models"
	"cloudcord/chat_api/mq"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	chatService *logic.ChatService
	mongoClient *mongo.Client
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASS")
	if user == "" || pass == "" {
		panic("Missing MongoDB credentials in env")
	}
	uri := fmt.Sprintf("mongodb+srv://%s:%s@messages.vbkzymr.mongodb.net/?retryWrites=true&w=majority&appName=Messages", user, pass)

	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mongoDB := mongoClient.Database("Messages")
	repo := db.NewChatRepository(mongoDB)
	chatService = logic.NewChatService(repo, &mq.NoopPublisher{})

	code := m.Run()

	_ = mongoClient.Disconnect(ctx)
	os.Exit(code)
}

func TestSendMessageHandler_Integration(t *testing.T) {
	payload := map[string]string{
		"sender":   "alice_test",
		"receiver": "bob_test",
		"content":  "Integration test message",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := sendMessageHandler(chatService)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := []string{"alice_test", "bob_test"}
	chat, err := chatService.GetChatByUsers(ctx, users[0], users[1])
	if err != nil {
		t.Fatalf("Failed to fetch chat after sending message: %v", err)
	}

	found := false
	for _, msg := range chat.Messages {
		if msg.Content == "Integration test message" && msg.SentByUser == "alice_test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Message not found in chat after sendMessageHandler")
	}
}

func TestGetChatHandler_Integration(t *testing.T) {
	user1 := "alice_test"
	user2 := "bob_test"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = chatService.CreateChat(ctx, user1, user2)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/get?user1=%s&user2=%s", user1, user2), nil)
	rr := httptest.NewRecorder()

	handler := getChatHandler(chatService)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var chat models.Chat
	if err := json.Unmarshal(rr.Body.Bytes(), &chat); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}
