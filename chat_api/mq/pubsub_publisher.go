package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type PubSubPublisher struct {
	topic *pubsub.Topic
}

func NewPubSubPublisher(ctx context.Context, projectID, topicName string) (*PubSubPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	topic := client.Topic(topicName)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check topic existence: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("pubsub topic %q does not exist", topicName)
	}

	return &PubSubPublisher{
		topic: topic,
	}, nil
}

func (p *PubSubPublisher) Publish(ctx context.Context, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	_, err = result.Get(ctx)
	return err
}
