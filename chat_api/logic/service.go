package logic

import (
	"cloudcord/chat/db"
	"cloudcord/chat/models"
	"cloudcord/chat/mq"
	"context"
	"log"
	"sort"
	"time"
)

type ChatService struct {
	repo      *db.ChatRepository
	publisher *mq.Publisher
}

// constructor
func NewChatService(repo *db.ChatRepository, publisher *mq.Publisher) *ChatService {
	return &ChatService{
		repo:      repo,
		publisher: publisher,
	}
}

// send message to user and publish notification to rabbitmq queue
func (s *ChatService) SendMessageToUser(ctx context.Context, sender, receiver, content string) error {
	users := []string{sender, receiver}
	sort.Strings(users)

	message := models.Message{
		Content:    content,
		SentByUser: sender,
		Timestamp:  time.Now(),
	}

	err := s.repo.AddMessageToChat(ctx, users, message)
	if err != nil {
		return err
	}

	notification := models.MessageNotification{
		ReceiverID: receiver,
		Message:    "You have a new message by " + sender,
	}

	if err := s.publisher.Publish(notification); err != nil {
		log.Printf("Failed to publish notification: %v", err)
	}

	return nil
}

// get chat by two users
func (s *ChatService) GetChatByUsers(ctx context.Context, user1, user2 string) (*models.Chat, error) {
	users := []string{user1, user2}
	sort.Strings(users)

	return s.repo.GetChatByUsers(ctx, users)
}
