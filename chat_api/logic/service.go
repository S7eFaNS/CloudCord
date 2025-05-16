package logic

import (
	"cloudcord/chat/db"
	"cloudcord/chat/models"
	"context"
	"sort"
	"time"
)

type ChatService struct {
	repo *db.ChatRepository
}

func NewChatService(repo *db.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) CreateChatWithMessage(ctx context.Context, user1, user2 string, content string) error {
	users := []string{user1, user2}
	sort.Strings(users)

	message := models.Message{
		Content:    content,
		SentByUser: user1,
		Timestamp:  time.Now(),
	}

	chat := &models.Chat{
		Users:    users,
		Messages: []models.Message{message},
	}

	return s.repo.CreateChat(ctx, chat)
}
