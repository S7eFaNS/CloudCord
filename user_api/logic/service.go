package logic

import (
	"cloudcord/user_api/graphdb"
	"cloudcord/user_api/models"
	"cloudcord/user_api/mq"
	"log"
)

type UserRepository interface {
	GetUserByAuth0ID(auth0ID string) (*models.User, error)
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	DeleteUserByAuth0ID(auth0ID string) error
	AddFriend(userID, friendID uint) error
	AreFriends(userID, otherUserID uint) (bool, error)
}

type UserLogic struct {
	repo      UserRepository
	publisher *mq.Publisher
}

type Publisher interface {
	Publish(deleteUsr interface{}) error
}

func NewUserLogic(repo UserRepository) *UserLogic {
	return &UserLogic{repo: repo}
}

func NewUserLogicRabbitMQ(repo UserRepository, publisher *mq.Publisher) *UserLogic {
	return &UserLogic{repo: repo, publisher: publisher}
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

func (ul *UserLogic) GetUserByAuth0ID(auth0ID string) (*models.User, error) {
	user, err := ul.repo.GetUserByAuth0ID(auth0ID)
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

func (ul *UserLogic) DeleteUserByAuth0ID(auth0ID string) error {
	err := ul.repo.DeleteUserByAuth0ID(auth0ID)
	if err != nil {
		log.Printf("Failed to delete user from DB: %v", err)
		return err
	}

	if ul.publisher != nil {
		msg := models.UserDeletedMessage{
			Auth0ID: auth0ID,
		}
		if err := ul.publisher.Publish(msg); err != nil {
			log.Printf("Failed to publish user deletion message: %v", err)
			return err
		} else {
			log.Printf("Auth0 ID: %s", auth0ID)
		}
	}

	log.Printf("Successfully deleted user with Auth0 ID: %s", auth0ID)
	return nil
}

func (ul *UserLogic) AddFriend(userID, friendID uint) error {
	if userID == friendID {
		log.Printf("User %d cannot befriend themselves", userID)
		return nil
	}

	err := ul.repo.AddFriend(userID, friendID)
	if err != nil {
		log.Printf("Failed to add friend: %v", err)
		return err
	}

	err = graphdb.CreateFriendship(userID, friendID)
	if err != nil {
		log.Printf("Failed to sync friendship to Neo4j: %v", err)
		return err
	}

	log.Printf("User %d and User %d are now friends", userID, friendID)
	return nil
}

func (ul *UserLogic) AreFriends(userID, otherUserID uint) (bool, error) {
	areFriends, err := ul.repo.AreFriends(userID, otherUserID)
	if err != nil {
		log.Printf("Error checking friendship status: %v", err)
		return false, err
	}
	return areFriends, nil
}

func (ul *UserLogic) GetFriendRecommendations(userID uint) ([]models.UserRecommendation, error) {
	_, err := ul.repo.GetUserByID(userID)
	if err != nil {
		log.Printf("User %d does not exist or could not be retrieved: %v", userID, err)
		return nil, err
	}

	recs, err := graphdb.GetFriendRecommendations(userID)
	if err != nil {
		log.Printf("Failed to get friend recommendations from graphdb: %v", err)
		return nil, err
	}

	var recommendations []models.UserRecommendation

	for _, rec := range recs {
		user, err := ul.repo.GetUserByID(rec.UserID)
		if err != nil {
			log.Printf("Failed to fetch user %d from PostgreSQL: %v", rec.UserID, err)
			return nil, err
		}

		recommendations = append(recommendations, models.UserRecommendation{
			ID:                user.UserID,
			Username:          user.Username,
			MutualFriendCount: rec.MutualFriendCount,
		})
	}

	return recommendations, nil
}
