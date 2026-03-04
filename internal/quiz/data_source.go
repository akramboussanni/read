package quiz

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
)

// DataSource interface for loading and parsing quiz data
type DataSource interface {
	GetName() string
	GetType() string
	GetRandomAnswers(ctx context.Context, excludeQuestionID int64, count int, filters map[string]interface{}) ([]string, error)
	CanHandle(filename string, rawData []byte) bool
	Parse(rawData []byte) (*ParsedDeck, error)
}

// UniversalDataSource handles universal JSON format decks
type UniversalDataSource struct{}

// GetName returns the name of the data source
func (u *UniversalDataSource) GetName() string {
	return "Universal JSON"
}

// GetType returns the type of the data source
func (u *UniversalDataSource) GetType() string {
	return "universal"
}

// CanHandle checks if this data source can handle the given file
func (u *UniversalDataSource) CanHandle(filename string, rawData []byte) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".json" {
		return false
	}

	var deck UniversalDeck
	if err := json.Unmarshal(rawData, &deck); err != nil {
		return false
	}

	// Check if it has the required fields
	return deck.Metadata.Title != "" && len(deck.Entries) > 0
}

// Parse parses the raw data into a ParsedDeck
func (u *UniversalDataSource) Parse(rawData []byte) (*ParsedDeck, error) {
	var deck UniversalDeck
	if err := json.Unmarshal(rawData, &deck); err != nil {
		return nil, err
	}

	parsed := &ParsedDeck{
		DeckKey:    strings.ReplaceAll(strings.ToLower(deck.Metadata.Title), " ", "_"),
		Title:      deck.Metadata.Title,
		Version:    deck.Metadata.Version,
		Metadata:   deck.Metadata,
		Categories: make(map[string]string),
		Questions:  make([]ParsedQuestion, 0, len(deck.Entries)),
	}
	parsed.Metadata.Classes = deck.Classes

	// Populate categories
	for key, cat := range deck.Categories {
		parsed.Categories[key] = cat.Title
	}
	parsed.Metadata.CategoryTitles = parsed.Categories

	// Populate questions
	for _, entry := range deck.Entries {
		// For vocabulary, assume source_text and target_text
		sourceText, _ := entry.Data["source_text"].(string)
		targetText, _ := entry.Data["target_text"].(string)

		question := ParsedQuestion{
			ID:             entry.ID,
			CategoryKey:    entry.Category,
			QuestionText:   sourceText, // Default to source
			CorrectAnswer:  targetText,
			AdditionalData: entry.Data,
		}
		parsed.Questions = append(parsed.Questions, question)
	}

	return parsed, nil
}

// GetRandomAnswers gets random answers from the deck, filtered by categories
func (u *UniversalDataSource) GetRandomAnswers(ctx context.Context, excludeQuestionID int64, count int, filters map[string]interface{}) ([]string, error) {
	// This would need access to the parsed deck or database
	// For now, return empty - will be implemented when integrated with repo
	return []string{}, nil
}

// DataSourceRegistry manages available data sources
type DataSourceRegistry struct {
	sources []DataSource
}

// NewDataSourceRegistry creates a new registry
func NewDataSourceRegistry() *DataSourceRegistry {
	return &DataSourceRegistry{
		sources: []DataSource{
			&UniversalDataSource{},
		},
	}
}

// GetDataSourceForFile returns the appropriate data source for a file
func (r *DataSourceRegistry) GetDataSourceForFile(filename string, rawData []byte) DataSource {
	for _, source := range r.sources {
		if source.CanHandle(filename, rawData) {
			return source
		}
	}
	return nil
}

// GetAllSources returns all registered sources
func (r *DataSourceRegistry) GetAllSources() []DataSource {
	return r.sources
}
