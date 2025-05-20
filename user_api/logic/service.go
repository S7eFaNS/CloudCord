package logic

import (
	"cloudcord/user/db"
	"cloudcord/user/models"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserLogic struct {
	repo *db.Repository
}

func NewUserLogic(repo *db.Repository) *UserLogic {
	return &UserLogic{repo: repo}
}

func (ul *UserLogic) CreateUserIfNotExists(ctx context.Context, auth0ID, username string) error {
	user, err := ul.repo.GetUserByAuth0ID(ctx, auth0ID)
	if err == nil && user != nil {
		log.Printf("User with Auth0ID %s already exists", auth0ID)
		return nil
	}

	user = &models.User{
		Auth0ID:  auth0ID,
		Username: username,
	}
	err = ul.repo.CreateUser(ctx, user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}
	log.Printf("User created successfully: %v", user)
	return nil
}

func (ul *UserLogic) GetUserByIDHandler(ctx context.Context, id string) (*models.User, error) {
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Invalid ID format: %v", err)
		return nil, err
	}

	user, err := ul.repo.GetUserByID(ctx, id)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		return nil, err
	}
	log.Printf("Retrieved user: %v", user)
	return user, nil
}

func (ul *UserLogic) GetAllUsersHandler(ctx context.Context) ([]models.User, error) {
	users, err := ul.repo.GetAllUsers(ctx)
	if err != nil {
		log.Printf("Error retrieving all users: %v", err)
		return nil, err
	}
	log.Printf("Success retrieving all users: %v", users)
	return users, nil
}
