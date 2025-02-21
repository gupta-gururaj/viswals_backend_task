package usecases

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/viswals_backend_task/pkg/encryptions"
	"github.com/viswals_backend_task/pkg/models"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 15 * time.Second
	batchSize      = 10
	// workerCount    = 3 // Reduce to avoid excessive goroutines
)

type Consumer struct {
	messageBroker MessageBroker
	cacheStore    CacheStore
	logger        *zap.Logger
	repo          UserRepository
	channel       <-chan amqp.Delivery
}

// NewConsumer initializes a new consumer instance	
func NewConsumer(messageBroker MessageBroker, userRepo UserRepository, cacheStore CacheStore, logger *zap.Logger) (*Consumer, error) {
	in, err := messageBroker.Subscribe(context.Background())
	if err != nil {
		return nil, err
	}

	return &Consumer{
		messageBroker: messageBroker,
		channel:       in,
		logger:        logger,
		repo:          userRepo,
		cacheStore:    cacheStore,
	}, nil
}

// Consume listens for incoming messages and processes them in batches
func (c *Consumer) Consume(wg *sync.WaitGroup, size int) {
	defer wg.Done()

	userDetailsChan := make(chan []*models.UserDetails, workerCount)
	errorChan := make(chan error, workerCount)

	var internalWg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		internalWg.Add(1)
		go c.processBatch(&internalWg, userDetailsChan, errorChan)
	}

	internalWg.Add(1)
	go c.logErrors(&internalWg, errorChan)

	var batch []*models.UserDetails
	timeout := time.NewTimer(1 * time.Second)


	for {
		select {
		case data, ok := <-c.channel:
			if !ok {
				c.logger.Warn("RabbitMQ channel closed, stopping consumer...")
				close(userDetailsChan)
				internalWg.Wait()
				close(errorChan)
				c.logger.Info("All data processed successfully.") // Final success message
				return
			}

			var user models.UserDetails
			if err := json.Unmarshal(data.Body, &user); err != nil {
				c.logger.Error("Error unmarshalling user details", zap.Error(err))
				continue
			}

			batch = append(batch, &user)

			if len(batch) >= batchSize {
				userDetailsChan <- batch
				batch = nil // Reset batch
			}

		case <-timeout.C:
			if len(batch) > 0 {
				userDetailsChan <- batch
				batch = nil
			}
			timeout.Reset(1 * time.Second)
		}
	}
}

// logErrors listens for errors and logs them
func (c *Consumer) logErrors(wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	for err := range errorChan {
		c.logger.Error("Consumer error", zap.Error(err))
	}
}

// processBatch processes a batch of user details, encrypting emails and storing them
func (c *Consumer) processBatch(wg *sync.WaitGroup, inputChan chan []*models.UserDetails, errorChan chan error) {
	defer wg.Done()

	for batch := range inputChan {
		// Encrypt email addresses
		for _, user := range batch {
			encEmail, err := encryptions.Encrypt(user.EmailAddress)
			if err != nil {
				errorChan <- err
				continue
			}
			user.EmailAddress = encEmail
		}

		// Store batch in the database with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		err := c.repo.CreateBulkUsers(ctx, batch)
		cancel()

		if err != nil {
			errorChan <- err
			continue
		}

		// Cache the processed batch
		if err := c.cacheStore.SetBulk(context.Background(), batch); err != nil {
			errorChan <- err
		}
	}
}

// Close shuts down the consumer gracefully
func (c *Consumer) Close() error {
	c.logger.Info("Closing consumer connection...")
	return c.messageBroker.Close()
}
