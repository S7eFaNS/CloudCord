package logic_test

import (
	"cloudcord/user_api/logic"
	"cloudcord/user_api/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetUserByAuth0ID(auth0ID string) (*models.User, error) {
	args := m.Called(auth0ID)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
}

func (m *MockUserRepo) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) GetUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetAllUsers() ([]models.User, error) {
	args := m.Called()
	users := args.Get(0)
	if users == nil {
		return nil, args.Error(1)
	}
	return users.([]models.User), args.Error(1)
}

func TestCreateUserIfNotExists_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	existingUser := &models.User{Auth0ID: "auth0|123", Username: "alice"}

	mockRepo.On("GetUserByAuth0ID", "auth0|123").Return(existingUser, nil)

	err := userLogic.CreateUserIfNotExists("auth0|123", "alice")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateUserIfNotExists_UserDoesNotExist_CreateSuccess(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	mockRepo.On("GetUserByAuth0ID", "auth0|123").Return(nil, assert.AnError) // simulate not found
	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := userLogic.CreateUserIfNotExists("auth0|123", "alice")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateUserIfNotExists_CreateFails(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	mockRepo.On("GetUserByAuth0ID", "auth0|123").Return(nil, assert.AnError)
	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(assert.AnError)

	err := userLogic.CreateUserIfNotExists("auth0|123", "alice")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByIDHandler_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	expectedUser := &models.User{Auth0ID: "auth0|123", Username: "alice"}

	mockRepo.On("GetUserByID", uint(1)).Return(expectedUser, nil)

	user, err := userLogic.GetUserByIDHandler(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByIDHandler_Failure(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	mockRepo.On("GetUserByID", uint(1)).Return(nil, assert.AnError)

	user, err := userLogic.GetUserByIDHandler(1)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestGetAllUsersHandler_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	users := []models.User{
		{Auth0ID: "auth0|123", Username: "alice"},
		{Auth0ID: "auth0|456", Username: "bob"},
	}

	mockRepo.On("GetAllUsers").Return(users, nil)

	returnedUsers, err := userLogic.GetAllUsersHandler()

	assert.NoError(t, err)
	assert.Equal(t, users, returnedUsers)
	mockRepo.AssertExpectations(t)
}

func TestGetAllUsersHandler_Failure(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userLogic := logic.NewUserLogic(mockRepo)

	mockRepo.On("GetAllUsers").Return(nil, assert.AnError)

	users, err := userLogic.GetAllUsersHandler()

	assert.Error(t, err)
	assert.Nil(t, users)
	mockRepo.AssertExpectations(t)
}
