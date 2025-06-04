package mq

import (
	"cloudcord/chat_api/db"
	"cloudcord/chat_api/models"
	"context"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func StartUserDeletionConsumer(amqpURL string, queueName string, repo *db.ChatRepository) error {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		true,  // auto-ack
		false, // exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var msg models.UserDeletedMessage
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Failed to parse user deletion message: %v", err)
				continue
			}

			log.Printf("Received user deletion for Auth0ID: %s", msg.Auth0ID)
			if err := repo.DeleteChatsByAuth0ID(context.Background(), msg.Auth0ID); err != nil {
				log.Printf("❌ Failed to delete chats for user %s: %v", msg.Auth0ID, err)
			} else {
				log.Printf("✅ Deleted chats for user %s", msg.Auth0ID)
			}
		}
	}()

	return nil
}
