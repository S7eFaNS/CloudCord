package db

import (
	"cloudcord/chat/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{
		collection: db.Collection("chats"),
	}
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *models.Chat) error {
	_, err := r.collection.InsertOne(ctx, chat)
	return err
}

func (r *ChatRepository) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	var chat models.Chat
	err := r.collection.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"chat_id": chatID})
	return err
}
