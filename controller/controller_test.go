package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/postgres"
)

// Mock UserService

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetAllUsers(ctx context.Context) ([]*models.UserDetails, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.UserDetails), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id string) (*models.UserDetails, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.UserDetails), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.UserDetails) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsersSSE(ctx context.Context, limit, offset int64) ([]byte, error) {
    args := m.Called(ctx, limit, offset)
    return args.Get(0).([]byte), args.Error(1)
}

func setupTestController() (*Controller, *MockUserService, *fiber.App) {
	mockService := new(MockUserService)
	ctrl := &Controller{UserService: mockService}
	app := fiber.New()
	return ctrl, mockService, app
}

func TestGetAllUsers(t *testing.T) {
	ctrl, mockService, app := setupTestController()
	app.Get("/users", ctrl.GetAllUsers)

	users := []*models.UserDetails{{ID: 1, EmailAddress: "test@example.com"}}
	mockService.On("GetAllUsers", mock.Anything).Return(users, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, _ := app.Test(req, -1)

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetUser_NotFound(t *testing.T) {
	ctrl, mockService, app := setupTestController()
	app.Get("/users/:id", ctrl.GetUser)

	mockService.On("GetUser", mock.Anything, "1").Return((*models.UserDetails)(nil), postgres.ErrNoData)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	resp, _ := app.Test(req, -1)

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestCreateUser(t *testing.T) {
	ctrl, mockService, app := setupTestController()
	app.Post("/users", ctrl.CreateUser)

	user := models.UserDetails{ID: 1, EmailAddress: "test@example.com"}
	mockService.On("CreateUser", mock.Anything, &user).Return(nil)

	body, _ := json.Marshal(user)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestDeleteUser(t *testing.T) {
	ctrl, mockService, app := setupTestController()
	app.Delete("/users/:id", ctrl.DeleteUser)

	mockService.On("DeleteUser", mock.Anything, "1").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	resp, _ := app.Test(req, -1)

	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}
