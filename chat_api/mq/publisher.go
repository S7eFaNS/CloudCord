package mq

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Publisher struct {
	channel *amqp.Channel
	queue   amqp.Queue
}

type NoopPublisher struct{}

func (n *NoopPublisher) Publish(msg interface{}) error {
	return nil
}

func NewPublisher(amqpURL, queueName string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		channel: ch,
		queue:   q,
	}, nil
}

func (p *Publisher) Publish(notification interface{}) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		"",
		p.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
