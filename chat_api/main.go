package main

import (
	"cloudcord/chat/db"
	"cloudcord/chat/logic"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type createChatRequest struct {
	User1   string `json:"user1"`
	User2   string `json:"user2"`
	Content string `json:"content"`
}

type sendMessageRequest struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
}

func sendMessageHandler(chatLogic *logic.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req sendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := chatLogic.SendMessageToUser(ctx, req.Sender, req.Receiver, req.Content)
		if err != nil {
			http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"message sent"}`))
	}
}

func getChatHandler(chatLogic *logic.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}

		user1 := r.URL.Query().Get("user1")
		user2 := r.URL.Query().Get("user2")
		log.Printf("📥 Received query: user1=%q, user2=%q", user1, user2)

		if user1 == "" || user2 == "" {
			http.Error(w, "Missing user1 or user2 query parameters", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		chat, err := chatLogic.GetChatByUsers(ctx, user1, user2)
		if err != nil {
			http.Error(w, "Chat not found: "+err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chat)
	}
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

	chatRepo := db.NewChatRepository(mongoDB)

	chatService := logic.NewChatService(chatRepo)

	http.HandleFunc("/send", sendMessageHandler(chatService))
	http.HandleFunc("/chat", getChatHandler(chatService))

	http.ListenAndServe(":8084", nil)
}
