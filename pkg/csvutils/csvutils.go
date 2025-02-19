package csvutils

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
)

// ReadRows reads up to 'n' rows from the CSV reader.
func ReadRows(reader *csv.Reader, n int) ([][]string, []string, error) {
	records := make([][]string, 0, n)
	for i := 0; i < n; i++ {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return records, nil, err
		} else if errors.Is(err, csv.ErrFieldCount) {
			return records[:len(records)-1], records[len(records)-1], nil
		}
		records = append(records, record)
	}
	return records, nil, nil
}

// OpenFile opens a CSV file in read-only mode and initializes a CSV reader.
func OpenFile(filePath string) (*csv.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(file)
	// Determine number of fields per record by reading the first row.
	firstRow, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	csvReader.FieldsPerRecord = len(firstRow)
	csvReader.ReuseRecord = false
	csvReader.Comma = ','

	return csvReader, nil
}

// ReadAll reads all records from the CSV reader.
func ReadAll(reader *csv.Reader) ([][]string, error) {
	return reader.ReadAll()
}
