package handlers

import (
	"cloudcord/chat_api/logic"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type sendMessageRequest struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
}

func SendMessageHandler(chatLogic *logic.ChatService) http.HandlerFunc {
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

// get chat by two users
func GetChatHandler(chatLogic *logic.ChatService) http.HandlerFunc {
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
			if err == mongo.ErrNoDocuments {
				chat, err = chatLogic.CreateChat(ctx, user1, user2)
				if err != nil {
					http.Error(w, "Failed to create chat: "+err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "Error retrieving chat: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chat)
	}
}
