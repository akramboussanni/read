package quiz

import (
	"fmt"
	"io"
	"os"
)

// UniversalImporter imports decks from files using appropriate data sources
type UniversalImporter struct {
	registry *DataSourceRegistry
}

// NewUniversalImporter creates a new universal importer
func NewUniversalImporter() *UniversalImporter {
	return &UniversalImporter{
		registry: NewDataSourceRegistry(),
	}
}

// ImportFromFile imports a deck from a file
func (u *UniversalImporter) ImportFromFile(filePath string) (*ParsedDeck, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	rawData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	dataSource := u.registry.GetDataSourceForFile(filePath, rawData)
	if dataSource == nil {
		return nil, fmt.Errorf("no suitable data source found for file: %s", filePath)
	}

	parsedDeck, err := dataSource.Parse(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deck: %w", err)
	}

	return parsedDeck, nil
}

// ImportFromBytes imports a deck from raw bytes
func (u *UniversalImporter) ImportFromBytes(filename string, rawData []byte) (*ParsedDeck, error) {
	dataSource := u.registry.GetDataSourceForFile(filename, rawData)
	if dataSource == nil {
		return nil, fmt.Errorf("no suitable data source found for data")
	}

	parsedDeck, err := dataSource.Parse(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deck: %w", err)
	}

	return parsedDeck, nil
}
