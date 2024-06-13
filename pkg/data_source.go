package pkg

import (
	"encoding/csv"
	"fmt"
	"os"
)

type DataSourceType string

const (
	DataSourceTypeCsv DataSourceType = "csv"
)

type DataSource interface {
	Schema() (*Schema, error)
	FromFile(filePath string) error
}

func NewDataSource(sourceType DataSourceType) DataSource {
	switch sourceType {
	case DataSourceTypeCsv:
		return &csvDataSource{}
	default:
		panic("Unsupported data source type")
	}
}

type csvDataSource struct {
	schema *Schema
}

func (d *csvDataSource) Schema() (*Schema, error) {
	return d.schema, nil
}

func (d *csvDataSource) FromFile(filePath string) error {
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
	for _, eachRecord := range records {
		fmt.Println(eachRecord)
	}
	return nil
}
