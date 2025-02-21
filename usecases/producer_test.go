package usecases

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/rabbitmq/mockrabbitmq"
	"go.uber.org/zap"
)

// TestProducer_Start tests the producer's Start method.
func TestProducer_Start(t *testing.T) {
	// Mock CSV input
	csvData := `1,John,Doe,john@example.com,1622548800000,-1,-1,0
2,Jane,Doe,jane@example.com,1622548800000,-1,-1,1`

	reader := csv.NewReader(bytes.NewReader([]byte(csvData)))

	// Initialize mock broker
	mockBroker := new(mockrabbitmq.MockRabbitMQ)
	logger := zap.NewNop() // No-op logger for tests

	// Create producer instance
	producer := NewProducer(reader, mockBroker, logger)

	// Expected published messages
	expectedUsers := []*models.UserDetails{
		{
			ID:           1,
			FirstName:    "John",
			LastName:     "Doe",
			EmailAddress: "john@example.com",
			CreatedAt:    parseNullTime("1622548800000"),
			DeletedAt:    parseNullTime("-1"),
			MergedAt:     parseNullTime("-1"),
			ParentUserId: 0,
		},
		{
			ID:           2,
			FirstName:    "Jane",
			LastName:     "Doe",
			EmailAddress: "jane@example.com",
			CreatedAt:    parseNullTime("1622548800000"),
			DeletedAt:    parseNullTime("-1"),
			MergedAt:     parseNullTime("-1"),
			ParentUserId: 1,
		},
	}

	// Set up expected calls to Publish for each user
	for _, user := range expectedUsers {
		jsonData, _ := json.Marshal(user)
		mockBroker.On("Publish", mock.Anything, jsonData).Return(nil)
	}

	// Run the producer
	err := producer.Start()
	require.NoError(t, err)

	// Verify all expectations were met
	mockBroker.AssertExpectations(t)
}

// TestProducer_Start_Error tests the producer's handling of publish errors.
func TestProducer_Start_Error(t *testing.T) {
	csvData := `1,John,Doe,john@example.com,1622548800000,-1,-1,0`
	reader := csv.NewReader(bytes.NewReader([]byte(csvData)))

	mockBroker := new(mockrabbitmq.MockRabbitMQ)
	logger := zap.NewNop()

	producer := NewProducer(reader, mockBroker, logger)

	// Simulate publish failure
	mockBroker.On("Publish", mock.Anything, mock.Anything).Return(errors.New("publish error"))

	err := producer.Start()
	require.NoError(t, err) // Producer should handle errors internally

	// Verify Publish was called at least once
	mockBroker.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

// TestProducer_Close tests the Close method.
func TestProducer_Close(t *testing.T) {
	mockBroker := new(mockrabbitmq.MockRabbitMQ)
	mockBroker.On("Close").Return(nil)

	producer := NewProducer(nil, mockBroker, zap.NewNop())

	err := producer.Close()
	require.NoError(t, err)

	mockBroker.AssertExpectations(t)
}
