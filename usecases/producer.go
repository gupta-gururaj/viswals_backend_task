package usecases

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/viswals_backend_task/pkg/csvutils"
	"github.com/viswals_backend_task/pkg/models"
	"go.uber.org/zap"
)

const (
	publishTimeout = 15 * time.Second
	workerCount    = 5 // Number of concurrent workers
)

type Producer struct {
	csvReader *csv.Reader
	broker    MessageBroker
	logger    *zap.Logger
}

func NewProducer(csvReader *csv.Reader, broker MessageBroker, logger *zap.Logger) *Producer {
	return &Producer{
		csvReader: csvReader,
		broker:    broker,
		logger:    logger,
	}
}

func (p *Producer) Start() error {
	p.logger.Info("Starting producer")

	jobs := make(chan *models.UserDetails, workerCount*2)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go p.worker(jobs, &wg)
	}

	// Read CSV and send jobs to workers
	for {
		rows, invalidRows, err := p.readCSV()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			p.logger.Error("Error reading CSV", zap.Error(err))
			continue
		}

		if len(invalidRows) > 0 {
			p.logger.Warn("Invalid rows encountered", zap.Any("data", invalidRows))
		}

		users := p.transformData(rows)
		for _, user := range users {
			jobs <- user
		}
	}

	close(jobs) // Close the channel after sending all jobs
	wg.Wait()   // Wait for workers to complete

	return nil
}

func (p *Producer) readCSV() ([][]string, []string, error) {
	return csvutils.ReadRows(p.csvReader, 1) // Read one row at a time
}

// Worker function to process messages concurrently
func (p *Producer) worker(jobs <-chan *models.UserDetails, wg *sync.WaitGroup) {
	defer wg.Done()
	for user := range jobs {
		ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
		err := p.publishMessage(ctx, user)
		cancel()

		if err != nil {
			p.logger.Error("Failed to publish message", zap.Error(err))
		}
	}
}

func (p *Producer) publishMessage(ctx context.Context, user *models.UserDetails) error {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, jsonData)
}

func (p *Producer) transformData(data [][]string) []*models.UserDetails {
	var result []*models.UserDetails
	for _, row := range data {
		if len(row) != 8 {
			p.logger.Warn("Skipping incomplete row", zap.Any("row", row))
			continue
		}

		result = append(result, &models.UserDetails{
			ID:           parseInt64(row[0]),
			FirstName:    row[1],
			LastName:     row[2],
			EmailAddress: row[3],
			CreatedAt:    parseNullTime(row[4]),
			DeletedAt:    parseNullTime(row[5]),
			MergedAt:     parseNullTime(row[6]),
			ParentUserId: parseInt64(row[7]),
		})
	}
	return result
}

func parseInt64(value string) int64 {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func parseNullTime(value string) sql.NullTime {
	if value == "-1" {
		return sql.NullTime{Valid: false}
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: time.UnixMilli(n), Valid: true}
}

func (p *Producer) Close() error {
	return p.broker.Close()
}
