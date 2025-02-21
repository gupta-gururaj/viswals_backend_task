package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/viswals_backend_task/pkg/csvutils"
	"github.com/viswals_backend_task/pkg/logger"
	"github.com/viswals_backend_task/pkg/rabbitmq"
	"github.com/viswals_backend_task/usecases"
	"go.uber.org/zap"
)
var (
	DevelopmentMode = "development"
)

func main() {
	// Initialize the logger
	log, err := logger.Init(os.Stdout, strings.ToLower(os.Getenv("ENVIRONMENT")) == DevelopmentMode)
	if err != nil {
		fmt.Println("Logger initialization failed:", err)
		return
	}

	// Open the CSV file as a reader
	csvReader, err := csvutils.OpenFile(os.Getenv("CSV_FILE_PATH"))
	if err != nil {
		log.Error("Unable to open the CSV file", zap.Error(err))
		return
	}

	// Establish a connection with the message broker
	connStr := os.Getenv("RABBITMQ_CONNECTION_STRING")
	if connStr == "" {
		log.Error("Missing RabbitMQ connection string. Please set the RABBITMQ_CONNECTION_STRING environment variable.")
		return
	}

	queue := os.Getenv("RABBITMQ_QUEUE_NAME")
	if queue == "" {
		log.Error("Missing RabbitMQ queue name. Please set the RABBITMQ_QUEUE_NAME environment variable.")
		return
	}

	// Initialize the message broker
	messageBroker, err := rabbitmq.New(connStr, queue)
	if err != nil {
		log.Error("Failed to initialize RabbitMQ queue", zap.Error(err), zap.String("queueName", queue))
		return
	}
	defer messageBroker.Close()

	// Initialize the producer service
	producer := usecases.NewProducer(csvReader, messageBroker, log)

	log.Info("Initializing the producer service")

	// Start the producer service		
	err=producer.Start()
	if err!=nil{
		log.Error("Producer service failed to start", zap.Error(err), zap.String("queueName", queue))
		return
	}

	log.Info("Producer operation completed successfully.")



}
