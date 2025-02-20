package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/viswals_backend_task/pkg/logger"
	"github.com/viswals_backend_task/pkg/postgres"
	"github.com/viswals_backend_task/pkg/rabbitmq"
	"github.com/viswals_backend_task/pkg/redis"
	"github.com/viswals_backend_task/repository"
	"go.uber.org/zap"
)
var (
	DevelopmentMode = "development"
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


	repo:=repository.NewRepository(pg)

	log.Debug("repository layer initialized")

	ttl, err := time.ParseDuration(os.Getenv("REDIS_TTL"))
	if err != nil {
		log.Error("error fetching redis TTL throws error", zap.Error(err), zap.String("ttl", ttlstr))
		return
	}

	cacheStore,err:=redis.New(os.Getenv("REDIS_CONNECTION_STRING"),ttl)
	if err != nil {
		log.Error("error initializing redis throws error", zap.Error(err))
		return
	}

	messageBroker,err:=rabbitmq.New(os.Getenv("REDIS_CONNECTION_STRING"),os.Getenv("RABBITMQ_QUEUE_NAME"))
	if err != nil {
		log.Error("error initializing rabbitmq throws error", zap.Error(err))
		return
	}
	log.Debug("message broker initialized")

}
