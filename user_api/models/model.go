package models

import "gorm.io/gorm"

type User struct {
	UserID   uint   `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Auth0ID  string `gorm:"uniqueIndex;not null" json:"auth0_id"`
	Username string `gorm:"type:varchar(100);not null" json:"username"`
}

type UserDeletedMessage struct {
	Auth0ID string `json:"auth0_id"`
}

func MigrateUsers(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})
	return err
}
