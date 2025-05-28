package db

import (
	"cloudcord/chat_api/models"
	"context"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatRepository struct {
	collection *mongo.Collection
}

// constructor
func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{
		collection: db.Collection("chats"),
	}
}

// add message to chat
func (r *ChatRepository) AddMessageToChat(ctx context.Context, users []string, message models.Message) error {
	filter := bson.M{"users": users}

	update := bson.M{
		"$push": bson.M{"messages": message},
	}

	result := r.collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() == mongo.ErrNoDocuments {
		chat := &models.Chat{
			Users:    users,
			Messages: []models.Message{message},
		}
		_, err := r.collection.InsertOne(ctx, chat)
		return err
	}
	return result.Err()
}

// Get the chat between two users
func (r *ChatRepository) GetChatByUsers(ctx context.Context, users []string) (*models.Chat, error) {
	sort.Strings(users)

	var chat models.Chat
	err := r.collection.FindOne(ctx, bson.M{"users": users}).Decode(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// Create new chat for existing users
func (r *ChatRepository) CreateChat(ctx context.Context, users []string) (*models.Chat, error) {
	sort.Strings(users)

	chat := &models.Chat{
		Users:    users,
		Messages: []models.Message{},
	}

	_, err := r.collection.InsertOne(ctx, chat)
	if err != nil {
		return nil, err
	}

	return chat, nil
}
