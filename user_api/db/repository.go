package db

import (
	"cloudcord/user_api/models"
	"log"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateUser(user *models.User) error {
	result := r.DB.Create(user)
	return result.Error
}

func (r *Repository) GetUserByID(id uint) (*models.User, error) {
	var user models.User

	result := r.DB.First(&user, id)

	if result.Error != nil {
		log.Printf("Error getting user by id: %v", result.Error)
		return nil, result.Error
	}
	log.Printf("Success getting user by id")

	return &user, nil
}

func (r *Repository) GetUserByAuth0ID(auth0ID string) (*models.User, error) {
	var user models.User

	result := r.DB.Where("auth0_id = ?", auth0ID).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *Repository) GetAllUsers() ([]models.User, error) {
	var users []models.User

	result := r.DB.Find(&users)

	if result.Error != nil {
		log.Printf("Error retrieving users: %v", result.Error)
		return nil, result.Error
	}
	log.Printf("Successfully retrieved all users")

	return users, nil
}

func (r *Repository) DeleteUserByAuth0ID(auth0ID string) error {
	result := r.DB.Where("auth0_id = ?", auth0ID).Delete(&models.User{})
	return result.Error
}

func (r *Repository) AddFriend(userID, friendID uint) error {
	if userID == friendID {
		return nil
	}

	var count int64
	r.DB.Model(&models.Friendship{}).Where("user_id = ? AND friend_id = ?", userID, friendID).Count(&count)
	if count > 0 {
		return nil
	}

	friendships := []models.Friendship{
		{UserID: userID, FriendID: friendID},
		{UserID: friendID, FriendID: userID},
	}

	return r.DB.Create(&friendships).Error
}

func (r *Repository) AreFriends(userID, otherUserID uint) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, otherUserID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
