package models

import "gorm.io/gorm"

type User struct {
	UserID   uint   `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Auth0ID  string `gorm:"uniqueIndex;not null" json:"auth0_id"`
	Username string `gorm:"type:varchar(100);not null" json:"username"`
}

type Friendship struct {
	UserID   uint `gorm:"not null;index:idx_friendship,unique" json:"user_id"`
	FriendID uint `gorm:"not null;index:idx_friendship,unique" json:"friend_id"`
}

func (Friendship) TableName() string {
	return "friendships"
}

type UserDeletedMessage struct {
	Auth0ID string `json:"auth0_id"`
}

func MigrateAll(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Friendship{}); err != nil {
		return err
	}
	return nil
}
