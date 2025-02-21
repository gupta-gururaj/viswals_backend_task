package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/viswals_backend_task/pkg/encryptions"
	"github.com/viswals_backend_task/pkg/models"
	database "github.com/viswals_backend_task/pkg/postgres"
	"go.uber.org/zap"
)

type UserService struct {
	dataStore UserRepository
	memStore  CacheStore
	logger    *zap.Logger
}

// NewUserService initializes and returns a new UserService instance
func NewUserService(dataStore UserRepository, memStore CacheStore, logger *zap.Logger) *UserService {
	return &UserService{
		dataStore: dataStore,
		memStore:  memStore,
		logger:    logger,
	}
}

// GetUser retrieves user details by first checking the cache and then the database if needed
func (us *UserService) GetUser(ctx context.Context, userID string) (*models.UserDetails, error) {
	// first try to fetch data from cache.
	var user *models.UserDetails
	var err error

	user, err = us.memStore.Get(ctx, userID)
	if err != nil {
		us.logger.Warn("UserService: error getting user from cache", zap.String("user_id", userID), zap.Error(err))
		// if error fetch data from database
		user, err = us.dataStore.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		us.logger.Debug("UserService: user fetched from database", zap.String("user_id", userID))
		// if data is successfully fetched, update the cache.
		err = us.memStore.Set(ctx, userID, user)
		if err != nil {
			// log the error and we can safely ignore this error.
			us.logger.Warn("UserService: error setting user in cache", zap.String("user_id", userID), zap.Error(err))
		}
	}
	decryptedEmail, err := encryptions.Decrypt(user.EmailAddress)
	if err != nil {
		return nil, err
	}

	user.EmailAddress = decryptedEmail

	return user, nil
}

// GetAllUsers retrieves all users from the database and decrypts their emails
func (us *UserService) GetAllUsers(ctx context.Context, name, email string) ([]*models.UserDetails, error) {
	// fetch and return data from db for now.
	users, err := us.dataStore.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	// Apply filtering in-memory
	var filteredUsers []*models.UserDetails
	for _, user := range users {
		decryptedEmail, err := encryptions.Decrypt(user.EmailAddress)
		if err != nil {
			us.logger.Error("error decrypting email", zap.String("email", user.EmailAddress), zap.Error(err))
			return nil, err
		}

		user.EmailAddress = decryptedEmail

		// Convert fields to lowercase for case-insensitive search
		lowerFirstName := strings.ToLower(user.FirstName)
		lowerLastName := strings.ToLower(user.LastName)
		lowerEmail := strings.ToLower(decryptedEmail)
		lowerNameFilter := strings.ToLower(name)
		lowerEmailFilter := strings.ToLower(email)

		// Apply filtering
		if (name == "" || strings.Contains(lowerFirstName, lowerNameFilter) || strings.Contains(lowerLastName, lowerNameFilter)) &&
			(email == "" || strings.Contains(lowerEmail, lowerEmailFilter)) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers, nil
}

// DeleteUser removes a user from both the database and cache
func (us *UserService) DeleteUser(ctx context.Context, userID string) error {
	// delete user from db first
	err := us.dataStore.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	// delete user from memory.
	err = us.memStore.Delete(ctx, userID)
	if err != nil {
		us.logger.Warn("UserService: error deleting user from cache", zap.String("user_id", userID))
		// the data will be automatically expired with TTL.
	}

	return nil
}

// CreateUser encrypts the email and stores the user in both database and cache
func (us *UserService) CreateUser(ctx context.Context, user *models.UserDetails) error {
	// encrypt users email id
	newEmail, err := encryptions.Encrypt(user.EmailAddress)
	if err != nil {
		return err
	}

	user.EmailAddress = newEmail

	// first, insert the data in a database.
	err = us.dataStore.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	// upon successful insertion update the cache
	err = us.memStore.Set(ctx, fmt.Sprint(user.ID), user)
	if err != nil {
		us.logger.Warn("UserService: error setting user in cache", zap.Error(err), zap.Any("user", user))
	}

	return nil
}

// GetAllUsersSSE retrieves paginated user data, decrypts emails, and returns JSON
func (us *UserService) GetAllUsersSSE(ctx context.Context, limit, offset int64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var isLastData bool

	// Fetch paginated users from database
	users, err := us.dataStore.ListUsers(ctx, limit, offset)
	if err != nil {
		if errors.Is(err, database.ErrNoData) {
			isLastData = true
		}
		return nil, err
	}

	// Decrypt emails for each user
	for _, user := range users {
		decryptedEmail, err := encryptions.Decrypt(user.EmailAddress)
		if err != nil {
			us.logger.Error("error decrypting email", zap.String("email", user.EmailAddress), zap.Error(err))
			return nil, err
		}

		user.EmailAddress = decryptedEmail
	}

	// Convert user list to JSON format
	data, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}

	// If no more data, return special error to indicate end of pagination
	if isLastData {
		return data, database.ErrNoData
	}

	return data, nil
}
