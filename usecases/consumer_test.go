package usecases

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/rabbitmq/mockrabbitmq"
	"github.com/viswals_backend_task/pkg/redis/mockredis"
	"github.com/viswals_backend_task/repository/mockrepository"
	"go.uber.org/zap"
)

// TestConsumer validates the consume workflow with different scenarios.
func TestConsumer(t *testing.T) {
	mockUserRepo := new(mockrepository.MockRepository)
	mockCacheStore := new(mockredis.MockRedis)
	mockQueueStore := new(mockrabbitmq.MockRabbitMQ)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	// Mock RabbitMQ subscription
	deliveryChannel := make(chan amqp.Delivery, 10)
	mockQueueStore.On("Subscribe", mock.Anything).Return((<-chan amqp.Delivery)(deliveryChannel), nil)

	consumer, err := NewConsumer(mockQueueStore, mockUserRepo, mockCacheStore, logger)
	assert.NoError(t, err)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Start consumer
	go consumer.Consume(wg, 1)

	// Send valid user data
	testBody := []byte(`{
		"id": 1,
		"first_name": "John",
		"last_name": "Doe",
		"email_address": "john@doe.com",
		"created_at": null,
		"deleted_at": null,
		"merged_at": null,
		"parent_user_id": 1
	}`)

	deliveryChannel <- amqp.Delivery{Body: testBody}

	time.Sleep(2 * time.Second) // Allow processing time

	close(deliveryChannel)
	wg.Wait()

	// Verify mock expectations
	mockUserRepo.AssertExpectations(t)
	mockCacheStore.AssertExpectations(t)
	mockQueueStore.AssertExpectations(t)
}

// TestConvertToUserDetails validates JSON parsing.
func TestConvertToUserDetails(t *testing.T) {
	// logger, err := zap.NewDevelopment()
	// assert.NoError(t, err)

	// encryption := encryptions.InitEncryptionKey()
	// assert.NoError(t, err)

	testCases := []struct {
		testName  string
		testInput []byte
		expectErr bool
	}{
		{
			testName: "Valid input",
			testInput: []byte(`[
				{
					"id":1,
					"first_name":"John",
					"last_name":"Doe",
					"email_address":"john@doe.com",
					"created_at":null,
					"deleted_at":null,
					"merged_at":null,
					"parent_user_id":1
				}
			]`),
			expectErr: false,
		},
		{
			testName:  "Nil input",
			testInput: nil,
			expectErr: true,
		},
		{
			testName: "Invalid JSON",
			testInput: []byte(`[
				{
					"id":1,
					"first_name": 546,
					"last_name":"Doe",
					"email_address":"john@doe.com",
					"created_at":null,
					"deleted_at":null,
					"merged_at":null,
					"parent_user_id":1
				}
			]`),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var userDetails []*models.UserDetails
			err := json.Unmarshal(tc.testInput, &userDetails)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, userDetails)
			}
		})
	}
}
