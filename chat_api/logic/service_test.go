package logic_test

import (
	"cloudcord/chat_api/logic"
	"cloudcord/chat_api/models"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) AddMessageToChat(ctx context.Context, users []string, message models.Message) error {
	args := m.Called(ctx, users, message)
	return args.Error(0)
}

func (m *MockRepo) GetChatByUsers(ctx context.Context, users []string) (*models.Chat, error) {
	args := m.Called(ctx, users)
	chat, _ := args.Get(0).(*models.Chat)
	return chat, args.Error(1)
}

func (m *MockRepo) CreateChat(ctx context.Context, users []string) (*models.Chat, error) {
	args := m.Called(ctx, users)
	chat, _ := args.Get(0).(*models.Chat)
	return chat, args.Error(1)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestSendMessageToUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	sender := "alice"
	receiver := "bob"
	content := "Hey Bob!"

	users := []string{sender, receiver}
	sort.Strings(users)

	msgMatcher := mock.MatchedBy(func(m models.Message) bool {
		return m.Content == content && m.SentByUser == sender
	})

	notification := models.MessageNotification{
		ReceiverID: receiver,
		Message:    "You have a new message by " + sender,
	}

	mockRepo.On("AddMessageToChat", ctx, users, msgMatcher).Return(nil)
	mockPub.On("Publish", notification).Return(nil)

	err := service.SendMessageToUser(ctx, sender, receiver, content)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

// Test SendMessageToUser when AddMessageToChat fails
func TestSendMessageToUser_AddMessageFails(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	sender := "alice"
	receiver := "bob"
	content := "Hey Bob!"

	users := []string{sender, receiver}
	sort.Strings(users)

	msgMatcher := mock.MatchedBy(func(m models.Message) bool {
		return m.Content == content && m.SentByUser == sender
	})

	mockRepo.On("AddMessageToChat", ctx, users, msgMatcher).Return(assert.AnError)

	err := service.SendMessageToUser(ctx, sender, receiver, content)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockPub.AssertNotCalled(t, "Publish", mock.Anything)
}

// Test SendMessageToUser when Publish fails (should not return error)
func TestSendMessageToUser_PublishFails(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	sender := "alice"
	receiver := "bob"
	content := "Hey Bob!"

	users := []string{sender, receiver}
	sort.Strings(users)

	msgMatcher := mock.MatchedBy(func(m models.Message) bool {
		return m.Content == content && m.SentByUser == sender
	})

	notification := models.MessageNotification{
		ReceiverID: receiver,
		Message:    "You have a new message by " + sender,
	}

	mockRepo.On("AddMessageToChat", ctx, users, msgMatcher).Return(nil)
	mockPub.On("Publish", notification).Return(assert.AnError)

	err := service.SendMessageToUser(ctx, sender, receiver, content)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestGetChatByUsers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	user1 := "alice"
	user2 := "bob"
	users := []string{user1, user2}
	sort.Strings(users)

	expectedChat := &models.Chat{
		Users: users,
		Messages: []models.Message{
			{Content: "Hi", SentByUser: "alice", Timestamp: time.Now()},
		},
	}

	mockRepo.On("GetChatByUsers", ctx, users).Return(expectedChat, nil)

	chat, err := service.GetChatByUsers(ctx, user1, user2)

	assert.NoError(t, err)
	assert.Equal(t, expectedChat, chat)
	mockRepo.AssertExpectations(t)
}

// Test GetChatByUsers when GetChatByUsers repo method fails
func TestGetChatByUsers_RepoFails(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	user1 := "alice"
	user2 := "bob"
	users := []string{user1, user2}
	sort.Strings(users)

	mockRepo.On("GetChatByUsers", ctx, users).Return(nil, assert.AnError)

	chat, err := service.GetChatByUsers(ctx, user1, user2)

	assert.Error(t, err)
	assert.Nil(t, chat)
	mockRepo.AssertExpectations(t)
}

// Test successful chat creation
func TestCreateChat(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	user1 := "alice"
	user2 := "bob"
	users := []string{user1, user2}
	sort.Strings(users)

	expectedChat := &models.Chat{
		Users:    users,
		Messages: []models.Message{},
	}

	mockRepo.On("CreateChat", ctx, users).Return(expectedChat, nil)

	chat, err := service.CreateChat(ctx, user1, user2)

	assert.NoError(t, err)
	assert.Equal(t, expectedChat, chat)
	mockRepo.AssertExpectations(t)
}

// Test CreateChat failure from repository
func TestCreateChat_RepoFails(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockPub := new(MockPublisher)

	service := logic.NewChatService(mockRepo, mockPub)

	user1 := "alice"
	user2 := "bob"
	users := []string{user1, user2}
	sort.Strings(users)

	mockRepo.On("CreateChat", ctx, users).Return(nil, assert.AnError)

	chat, err := service.CreateChat(ctx, user1, user2)

	assert.Error(t, err)
	assert.Nil(t, chat)
	mockRepo.AssertExpectations(t)
}
