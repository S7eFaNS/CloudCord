package logic

import (
	"cloudcord/user/db"
	"cloudcord/user/models"
	"log"
)

type UserLogic struct {
	repo *db.Repository
}

func NewUserLogic(repo *db.Repository) *UserLogic {
	return &UserLogic{repo: repo}
}

func (ul *UserLogic) CreateUserHandler(username string) {
	user := &models.User{Username: username}
	err := ul.repo.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return
	}
	log.Println("User created successfully")
}

func (ul *UserLogic) GetUserByIDHandler(id uint) (*models.User, error) {
	user, err := ul.repo.GetUserByID(id)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		return nil, err
	}
	log.Printf("Retrieved user: %v", user)
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
