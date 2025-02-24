package usecases

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/viswals_backend_task/pkg/models"
)

type MessageBroker interface {
	Publish(ctx context.Context, message []byte) error
	Subscribe(ctx context.Context) (<-chan amqp.Delivery, error)
	Close() error
}

type UserRepository interface {
	CreateBulkUsers(ctx context.Context, users []*models.UserDetails) error
	CreateUser(ctx context.Context, user *models.UserDetails) error
	GetUserByID(ctx context.Context, id string) (*models.UserDetails, error)
	GetAllUsers(ctx context.Context) ([]*models.UserDetails, error)
	ListUsers(ctx context.Context, limit, offset int64) ([]*models.UserDetails, error)
	DeleteUser(ctx context.Context, id string) error
}

type CacheStore interface {
	Get(ctx context.Context, key string) (*models.UserDetails, error)
	Set(ctx context.Context, key string, userDetails *models.UserDetails) error
	SetBulk(ctx context.Context, userDetails []*models.UserDetails) error 
	Delete(ctx context.Context, key string) error
}
