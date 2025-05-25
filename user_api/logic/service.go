package logic

import (
	"cloudcord/user_api/models"
	"log"
)

type UserRepository interface {
	GetUserByAuth0ID(auth0ID string) (*models.User, error)
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetAllUsers() ([]models.User, error)
}

type UserLogic struct {
	repo UserRepository
}

func NewUserLogic(repo UserRepository) *UserLogic {
	return &UserLogic{repo: repo}
}

func (ul *UserLogic) CreateUserIfNotExists(auth0ID, username string) error {
	user, err := ul.repo.GetUserByAuth0ID(auth0ID)

	if err == nil && user != nil {
		log.Printf("User with Auth0ID %s already exists", auth0ID)
		return nil
	}

	user = &models.User{
		Auth0ID:  auth0ID,
		Username: username,
	}

	err = ul.repo.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}
	log.Printf("User created successfully: %v", user)

	return nil
}

func (ul *UserLogic) GetUserByIDHandler(id uint) (*models.User, error) {
	user, err := ul.repo.GetUserByID(id)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		return nil, err
	}
	return user, nil
}

func (ul *UserLogic) GetAllUsersHandler() ([]models.User, error) {
	users, err := ul.repo.GetAllUsers()
	if err != nil {
		log.Printf("Error retrieving all users: %v", err)
		return nil, err
	}
	log.Printf("Success retrieving all users: %v", users)
	return users, nil
}
