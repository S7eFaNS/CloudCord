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
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type createChatRequest struct {
	User1   string `json:"user1"`
	User2   string `json:"user2"`
	Content string `json:"content"`
}

func chatHandler(chatLogic *logic.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req createChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		users := []string{req.User1, req.User2}
		sort.Strings(users)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := chatLogic.CreateChatWithMessage(ctx, users[0], users[1], req.Content)
		if err != nil {
			http.Error(w, "Failed to create chat: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"chat created"}`))
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

	http.HandleFunc("/chat", chatHandler(chatService))

	http.ListenAndServe(":8084", nil)
}
