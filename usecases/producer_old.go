package usecases

// import (
// 	"context"
// 	"database/sql"
// 	"encoding/csv"
// 	"encoding/json"
// 	"errors"
// 	"io"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"github.com/viswals_backend_task/pkg/csvutils"
// 	"github.com/viswals_backend_task/pkg/models"
// 	"go.uber.org/zap"
// )

// var (
// 	publishTimeout = 15 * time.Second
// )

// type Producer struct {
// 	csvReader *csv.Reader
// 	broker    MessageBroker
// 	logger    *zap.Logger
// }

// func NewProducer(csvReader *csv.Reader, broker MessageBroker, logger *zap.Logger) *Producer {
// 	return &Producer{
// 		csvReader: csvReader,
// 		broker:    broker,
// 		logger:    logger,
// 	}
// }

// func (p *Producer) Start(batchSize int) error {
// 	p.logger.Info("Starting producer", zap.Int("batchSize", batchSize))

// 	for {
// 		rows, invalidRows, err := p.readCSV(batchSize)
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				break
// 			}
// 			p.logger.Error("Error reading CSV", zap.Error(err))
// 			continue
// 		}

// 		if len(invalidRows) > 0 {
// 			p.logger.Warn("Invalid rows encountered", zap.Any("data", invalidRows))
// 		}

// 		users := p.transformData(rows)
// 		if len(users) == 0 {
// 			continue
// 		}

// 		p.publishConcurrently(users)
// 	}

// 	return nil
// }
// func (p *Producer) readCSV(batchSize int) ([][]string, []string, error) {
// 	return csvutils.ReadRows(p.csvReader, batchSize)
// }

// func (p *Producer) publishConcurrently(data []*models.UserDetails) {
// 	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
// 	defer cancel()

// 	var wg sync.WaitGroup
// 	wg.Add(1)

// 	go func() {
// 		defer wg.Done()
// 		if err := p.publishMessages(ctx, data); err != nil {
// 			p.logger.Error("Failed to publish messages", zap.Error(err))
// 		}
// 	}()

// 	wg.Wait()
// }

// func (p *Producer) publishMessages(ctx context.Context, data []*models.UserDetails) error {
// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}
// 	return p.broker.Publish(ctx, jsonData)
// }

// func (p *Producer) transformData(data [][]string) []*models.UserDetails {
// 	var result []*models.UserDetails

// 	for _, row := range data {
// 		if len(row) != 8 {
// 			p.logger.Warn("Skipping incomplete row", zap.Any("row", row))
// 			continue
// 		}

// 		user := &models.UserDetails{
// 			ID:           parseInt64(row[0]),
// 			FirstName:    row[1],
// 			LastName:     row[2],
// 			EmailAddress: row[3],
// 			CreatedAt:    parseNullTime(row[4]),
// 			DeletedAt:    parseNullTime(row[5]),
// 			MergedAt:     parseNullTime(row[6]),
// 			ParentUserId: parseInt64(row[7]),
// 		}

// 		result = append(result, user)
// 	}

// 	return result
// }

// func parseInt64(value string) int64 {
// 	n, err := strconv.ParseInt(value, 10, 64)
// 	if err != nil {
// 		return 0
// 	}
// 	return n
// }

// func parseNullTime(value string) sql.NullTime {
// 	if value == "-1" {
// 		return sql.NullTime{Valid: false}
// 	}
// 	n, err := strconv.ParseInt(value, 10, 64)
// 	if err != nil {
// 		return sql.NullTime{Valid: false}
// 	}
// 	return sql.NullTime{Time: time.UnixMilli(n), Valid: true}
// }

// func (p *Producer) Close() error {
// 	return p.broker.Close()
// }
