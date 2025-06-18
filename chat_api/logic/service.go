package logic

import (
	"cloudcord/chat_api/models"
	"cloudcord/chat_api/mq"
	"context"
	"log"
	"sort"
	"time"
)

// Define interfaces for dependency inversion
type ChatRepository interface {
	AddMessageToChat(ctx context.Context, users []string, message models.Message) error
	GetChatByUsers(ctx context.Context, users []string) (*models.Chat, error)
	CreateChat(ctx context.Context, users []string) (*models.Chat, error)
	DeleteChatsByAuth0ID(ctx context.Context, auth0ID string) error
}

type Publisher interface {
	Publish(msg interface{}) error
}

// ChatService depends on interfaces, not concrete types
type ChatService struct {
	repo            ChatRepository
	publisher       Publisher
	pubsubPublisher *mq.PubSubPublisher
}

// Constructor takes interfaces now
func NewChatService(repo ChatRepository, publisher Publisher, pubsub *mq.PubSubPublisher) *ChatService {
	return &ChatService{
		repo:            repo,
		publisher:       publisher,
		pubsubPublisher: pubsub,
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

	if s.pubsubPublisher != nil {
		if err := s.pubsubPublisher.Publish(ctx, notification); err != nil {
			log.Printf("❌ Failed to publish to Pub/Sub: %v", err)
		} else {
			log.Println("✅ Notification published to Pub/Sub")
		}
	}

	return nil
}

// get chat by two users
func (s *ChatService) GetChatByUsers(ctx context.Context, user1, user2 string) (*models.Chat, error) {
	users := []string{user1, user2}
	sort.Strings(users)

	return s.repo.GetChatByUsers(ctx, users)
}

func (s *ChatService) CreateChat(ctx context.Context, user1, user2 string) (*models.Chat, error) {
	users := []string{user1, user2}
	sort.Strings(users)

	return s.repo.CreateChat(ctx, users)
}

func (s *ChatService) DeleteChatsByAuth0ID(ctx context.Context, auth0ID string) error {
	err := s.repo.DeleteChatsByAuth0ID(ctx, auth0ID)
	if err != nil {
		log.Printf("Failed to delete chats for user %s: %v", auth0ID, err)
		return err
	}
	log.Printf("Successfully deleted chats for user %s", auth0ID)
	return nil
}
