package g0database

import (
	"encoding/csv"
	"os"
)

type DataSourceType string

const (
	DataSourceTypeCsv DataSourceType = "csv"
)

type DataSource interface {
	LoadIntoTable(table *Table, filePath string) error
}

func NewDataSource(sourceType DataSourceType) DataSource {
	switch sourceType {
	case DataSourceTypeCsv:
		return &csvDataSource{}
	default:
		panic("Unsupported data source type")
	}
}

type csvDataSource struct{}

// LoadIntoTable loads CSV data into an existing table
// First row is treated as header and skipped
func (d *csvDataSource) LoadIntoTable(table *Table, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Skip header row, insert data rows
	for i := 1; i < len(records); i++ {
		values := make([]interface{}, len(records[i]))
		for j, v := range records[i] {
			values[j] = v
		}
		if err := table.Insert(values); err != nil {
			return err
		}
	}

	return nil
}
