package models

import "time"

type Message struct {
	Content    string    `bson:"content" json:"content"`
	SentByUser string    `bson:"sent_by_user" json:"sent_by_user"`
	Timestamp  time.Time `bson:"timestamp" json:"timestamp"`
}

type Chat struct {
	Users    []string  `bson:"users" json:"users"`
	Messages []Message `bson:"messages" json:"messages"`
}

type MessageNotification struct {
	ReceiverID string `json:"receiver_id"`
	Message    string `json:"message"`
}

type UserDeletedMessage struct {
	Auth0ID string `json:"auth0_id"`
}
