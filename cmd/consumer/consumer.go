package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/viswals_backend_task/pkg/encryptions"
	"github.com/viswals_backend_task/pkg/logger"
	"github.com/viswals_backend_task/pkg/postgres"
	"github.com/viswals_backend_task/pkg/rabbitmq"
	"github.com/viswals_backend_task/pkg/redis"
	"github.com/viswals_backend_task/repository"
	"github.com/viswals_backend_task/usecases"
	"go.uber.org/zap"
)

var (
	DevelopmentMode = "development"
	defaultBufferSize="50"
)

func main() {
	// initializing logger
	log, err := logger.Init(os.Stdout, strings.ToLower(os.Getenv("ENVIRONMENT")) == DevelopmentMode)
	if err != nil {
		fmt.Printf("can't initialise logger throws error : %v", err)
		return
	}

	pg, err := postgres.New(os.Getenv("POSTGRES_CONNECTION_STRING"))
	if err != nil {
		log.Fatal("failed to initialize postgres", zap.Error(err))
	}

	repo := repository.NewRepository(pg)

	log.Debug("repository layer initialized")

	ttl, err := time.ParseDuration(os.Getenv("REDIS_TTL"))
	if err != nil {
		log.Error("error fetching redis TTL throws error", zap.Error(err))
		return
	}

	cacheStore, err := redis.New(os.Getenv("REDIS_CONNECTION_STRING"), ttl)
	if err != nil {
		log.Error("error initializing redis throws error", zap.Error(err))
		return
	}
	log.Debug("cache store initialized")

	messageBroker, err := rabbitmq.New(os.Getenv("RABBITMQ_CONNECTION_STRING"), os.Getenv("RABBITMQ_QUEUE_NAME"))
	if err != nil {
		log.Error("error initializing rabbitmq throws error", zap.Error(err))
		return
	}
	log.Debug("message broker initialized")

	// Initialize the encryption key
	if err := encryptions.InitEncryptionKey(); err != nil {
		log.Error("error in initializing encryptions", zap.Error(err))
		return
	}

	uc, err := usecases.NewConsumer(messageBroker, repo, cacheStore, log)
	if err != nil {
		log.Error("error initializing consumer service throws error", zap.Error(err))
		return
	}

	log.Debug("service layer initialized")

	// Fetch and validate buffer size for channel
	channelCapacity := os.Getenv("CHANNEL_SIZE")
	if channelCapacity == "" {
		log.Warn("Buffer size is not set using environment variable 'CHANNEL_SIZE', using default buffer size", zap.Any("buffer_size", defaultBufferSize))

		channelCapacity = defaultBufferSize
	}

	bufferSize, err := strconv.Atoi(channelCapacity)
	if err != nil {
		log.Error("error parsing buffer size throws error", zap.Error(err), zap.Int("buffer_size", bufferSize))
		return
	}

	// create a separate go routine to handle upcoming data.
	wg := &sync.WaitGroup{}
	log.Info("starting consumer")

	wg.Add(1)
	go uc.Consume(wg, bufferSize)

	
	wg.Wait()

	


}
