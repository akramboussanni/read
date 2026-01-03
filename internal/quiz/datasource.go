package quiz

import (
	"context"
)

// DataSource provides questions and answers for quiz generation
// Implementations can handle different data formats (Quran words, general knowledge, etc.)
type DataSource interface {
	// GetName returns the unique identifier for this data source
	GetName() string

	// GetType returns the category type (e.g., "vocabulary", "general_knowledge")
	GetType() string

	// GetRandomAnswers retrieves random wrong answers for generating MCQ options
	// Filters can include:
	//   - "deck_id": int64 - Limit to specific deck
	//   - "category_id": int64 - Limit to specific category
	//   - "category_ids": []int64 - Limit to multiple categories
	//   - "exclude_category": int64 - Exclude specific category
	//   - "difficulty": string - Match difficulty level
	//   - "language": string - For language-specific answers
	GetRandomAnswers(ctx context.Context, excludeQuestionID int64, count int, filters map[string]interface{}) ([]string, error)

	// CanHandle determines if this data source can parse the given file
	CanHandle(filename string, rawData []byte) bool

	// Parse converts raw data into the normalized ParsedDeck structure
	Parse(rawData []byte) (*ParsedDeck, error)
}

// DataSourceRegistry manages all available data sources
type DataSourceRegistry struct {
	sources map[string]DataSource
}

// NewDataSourceRegistry creates a new registry
func NewDataSourceRegistry() *DataSourceRegistry {
	return &DataSourceRegistry{
		sources: make(map[string]DataSource),
	}
}

// Register adds a data source to the registry
func (r *DataSourceRegistry) Register(ds DataSource) {
	r.sources[ds.GetName()] = ds
}

// Get retrieves a data source by name
func (r *DataSourceRegistry) Get(name string) (DataSource, bool) {
	ds, ok := r.sources[name]
	return ds, ok
}

// GetAll returns all registered data sources
func (r *DataSourceRegistry) GetAll() map[string]DataSource {
	return r.sources
}

// FindImporter finds the appropriate data source for a file
func (r *DataSourceRegistry) FindImporter(filename string, rawData []byte) DataSource {
	for _, ds := range r.sources {
		if ds.CanHandle(filename, rawData) {
			return ds
		}
	}
	return nil
}
