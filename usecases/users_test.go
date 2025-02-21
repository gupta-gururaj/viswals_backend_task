package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/viswals_backend_task/pkg/encryptions"
	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/redis/mockredis"
	"github.com/viswals_backend_task/repository/mockrepository"
	"go.uber.org/zap"
)

type MockEncryptionService struct{}

func (m *MockEncryptionService) Decrypt(s string) (string, error) {
	return "decrypted@example.com", nil
}

func TestUserService_GetUser(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "a8z9WmX2pQJ5YcQ6dT7m9LqFkX4r7BsY") // ✅ Ensure valid key length
	defer os.Unsetenv("ENCRYPTION_KEY")
	mockRepo := new(mockrepository.MockRepository)
	mockCache := new(mockredis.MockRedis)
	logger := zap.NewNop()
	// ✅ Ensure encryption key is initialized
	err := encryptions.InitEncryptionKey()
	require.NoError(t, err, "failed to initialize encryption key")

	service := NewUserService(mockRepo, mockCache, logger)

	user := &models.UserDetails{ID: 1, EmailAddress: "noA+gYljXJW+c6QW+eW2rOL6fiO9Ltz5D3sI2mg7B5otubqMWP2nKWk="}
	mockCache.On("Get", mock.Anything, "1").Return((*models.UserDetails)(nil), errors.New("cache miss"))

	// mockCache.On("Get", mock.Anything, "1").Return(nil, errors.New("cache miss"))
	mockRepo.On("GetUserByID", mock.Anything, "1").Return(user, nil)
	mockCache.On("Set", mock.Anything, "1", user).Return(nil)

	// encryptions.Decrypt = func(s string) (string, error) { return "decrypted@example.com", nil }

	result, err := service.GetUser(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, "LMurphy1964@earthlink.com", result.EmailAddress)
}

func TestUserService_GetAllUsers(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "a8z9WmX2pQJ5YcQ6dT7m9LqFkX4r7BsY") // ✅ Ensure valid key length
	defer os.Unsetenv("ENCRYPTION_KEY")
	
	mockRepo := new(mockrepository.MockRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, nil, logger)
	err := encryptions.InitEncryptionKey()
	require.NoError(t, err, "failed to initialize encryption key")

	users := []*models.UserDetails{{ID: 1,FirstName: "Liam",LastName: "Murphy",EmailAddress: "noA+gYljXJW+c6QW+eW2rOL6fiO9Ltz5D3sI2mg7B5otubqMWP2nKWk="}}
	mockRepo.On("GetAllUsers", mock.Anything).Return(users, nil)
	// Decrypt = func(s string) (string, error) { return "decrypted@example.com", nil }

	result, err := service.GetAllUsers(context.Background(),"Liam", "LMurphy1964@earthlink.com")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Liam", result[0].FirstName)
	require.Equal(t, "Murphy", result[0].LastName)
	require.Equal(t, "LMurphy1964@earthlink.com", result[0].EmailAddress)
}

func TestUserService_DeleteUser(t *testing.T) {
	mockRepo := new(mockrepository.MockRepository)
	mockCache := new(mockredis.MockRedis)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, mockCache, logger)

	mockRepo.On("DeleteUser", mock.Anything, "1").Return(nil)
	mockCache.On("Delete", mock.Anything, "1").Return(nil)

	err := service.DeleteUser(context.Background(), "1")
	require.NoError(t, err)
}

func TestUserService_CreateUser(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "a8z9WmX2pQJ5YcQ6dT7m9LqFkX4r7BsY") // ✅ Ensure valid key length
	defer os.Unsetenv("ENCRYPTION_KEY")
	mockRepo := new(mockrepository.MockRepository)
	mockCache := new(mockredis.MockRedis)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, mockCache, logger)
	err := encryptions.InitEncryptionKey()
	require.NoError(t, err, "failed to initialize encryption key")

	user := &models.UserDetails{ID: 1, EmailAddress: "noA+gYljXJW+c6QW+eW2rOL6fiO9Ltz5D3sI2mg7B5otubqMWP2nKWk="}
	// encryptions.Encrypt = func(s string) (string, error) { return "encrypted@example.com", nil }

	mockRepo.On("CreateUser", mock.Anything, user).Return(nil)
	mockCache.On("Set", mock.Anything, "1", user).Return(nil)

	err = service.CreateUser(context.Background(), user)
	require.NoError(t, err)
}

func TestUserService_GetAllUsersSSE(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "a8z9WmX2pQJ5YcQ6dT7m9LqFkX4r7BsY") 
	defer os.Unsetenv("ENCRYPTION_KEY")
	mockRepo := new(mockrepository.MockRepository)
	logger := zap.NewNop()
	// mockEncryptor := &MockEncryptionService{}
	err := encryptions.InitEncryptionKey()
	require.NoError(t, err, "failed to initialize encryption key")

	service := NewUserService(mockRepo, nil, logger)

	users := []*models.UserDetails{{ID: 1, EmailAddress: "noA+gYljXJW+c6QW+eW2rOL6fiO9Ltz5D3sI2mg7B5otubqMWP2nKWk="}}
	mockRepo.On("ListUsers", mock.Anything, int64(10), int64(0)).Return(users, nil)
	// encryptions.Decrypt = func(s string) (string, error) { return "decrypted@example.com", nil }

	data, err := service.GetAllUsersSSE(context.Background(), 10, 0)
	require.NoError(t, err)

	var result []*models.UserDetails
	json.Unmarshal(data, &result)
	require.Equal(t, "LMurphy1964@earthlink.com", result[0].EmailAddress)
}
