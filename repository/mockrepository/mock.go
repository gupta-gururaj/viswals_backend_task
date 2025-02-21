package mockrepository

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/viswals_backend_task/pkg/models"
)

type MockRepository struct {
	mock.Mock
}

func (db *MockRepository) GetUserByID(ctx context.Context, id string) (*models.UserDetails, error) {
	args := db.Called(ctx, id)
	return args.Get(0).(*models.UserDetails), args.Error(1)
}

func (db *MockRepository) CreateUser(ctx context.Context, user *models.UserDetails) error {
	args := db.Called(ctx, user)
	return args.Error(0)
}

func (db *MockRepository) CreateBulkUsers(ctx context.Context, user []*models.UserDetails) error {
	args := db.Called(ctx, user)
	return args.Error(0)
}

func (db *MockRepository) GetAllUsers(ctx context.Context) ([]*models.UserDetails, error) {
	args := db.Called(ctx)
	return args.Get(0).([]*models.UserDetails), args.Error(1)
}

func (db *MockRepository) DeleteUser(ctx context.Context, id string) error {
	args := db.Called(ctx, id)
	return args.Error(0)
}

func (db *MockRepository) ListUsers(ctx context.Context, limit, offset int64) ([]*models.UserDetails, error) {
	args := db.Called(ctx, limit, offset)
	return args.Get(0).([]*models.UserDetails), args.Error(1)
}
