package usecases

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageBroker interface {
	Publish(ctx context.Context, message []byte) error
	Subscribe(ctx context.Context) (<-chan amqp.Delivery, error)
	Close() error
}
