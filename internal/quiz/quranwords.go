package quiz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akramboussanni/gocode/internal/repo"
)

// QuranWordsDataSource handles the Quranic vocabulary format
type QuranWordsDataSource struct {
	repos *repo.Repos
}

// NewQuranWordsDataSource creates a new Quran words data source
func NewQuranWordsDataSource(repos *repo.Repos) *QuranWordsDataSource {
	return &QuranWordsDataSource{repos: repos}
}

// GetName returns the identifier for this data source
func (ds *QuranWordsDataSource) GetName() string {
	return "quran_words"
}

// GetType returns the data source type
func (ds *QuranWordsDataSource) GetType() string {
	return "vocabulary"
}

// GetRandomAnswers retrieves random answers from the Quran words dataset
func (ds *QuranWordsDataSource) GetRandomAnswers(ctx context.Context, excludeQuestionID int64, count int, filters map[string]interface{}) ([]string, error) {
	if ds.repos == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	// Extract filters
	deckID, hasDeck := filters["deck_id"].(int64)
	categoryID, hasCategory := filters["category_id"].(int64)
	categoryIDs, hasCategories := filters["category_ids"].([]int64)
	excludeCategoryID, hasExcludeCategory := filters["exclude_category"].(int64)

	// Build query based on filters
	if hasCategories && len(categoryIDs) > 0 {
		return ds.repos.Question.GetRandomAnswersFromCategories(ctx, categoryIDs, excludeQuestionID, count)
	} else if hasCategory {
		return ds.repos.Question.GetRandomAnswersFromCategory(ctx, categoryID, excludeQuestionID, count)
	} else if hasExcludeCategory && hasDeck {
		return ds.repos.Question.GetRandomAnswersFromDeckExcludingCategory(ctx, deckID, excludeCategoryID, excludeQuestionID, count)
	} else if hasDeck {
		return ds.repos.Question.GetRandomAnswersFromDeck(ctx, deckID, excludeQuestionID, count)
	}

	// Fallback to any random answers
	return ds.repos.Question.GetRandomAnswers(ctx, excludeQuestionID, count)
}

// CanHandle determines if this data source can parse the given file
func (ds *QuranWordsDataSource) CanHandle(filename string, rawData []byte) bool {
	// Check if it has the expected structure
	var data struct {
		Deck struct {
			ID string `json:"id"`
		} `json:"deck"`
		Categories map[string]string `json:"cats"`
	}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return false
	}
	return data.Deck.ID != "" && data.Categories != nil
}

// Parse converts raw data into the normalized ParsedDeck structure
func (ds *QuranWordsDataSource) Parse(rawData []byte) (*ParsedDeck, error) {
	var data struct {
		Deck struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Version int    `json:"version"`
		} `json:"deck"`
		Categories map[string]string `json:"cats"`
		Questions  []struct {
			ID       string `json:"id"`
			Category string `json:"cat"`
			French   string `json:"fr"`
			Arabic   string `json:"ar"`
		} `json:"items"`
	}

	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	parsed := &ParsedDeck{
		DeckKey:    data.Deck.ID,
		Title:      data.Deck.Title,
		Version:    data.Deck.Version,
		Categories: data.Categories,
		Questions:  make([]ParsedQuestion, 0, len(data.Questions)),
	}

	for _, q := range data.Questions {
		parsed.Questions = append(parsed.Questions, ParsedQuestion{
			ID:            q.ID,
			CategoryKey:   q.Category,
			QuestionText:  q.Arabic,
			CorrectAnswer: q.French,
			AdditionalData: map[string]string{
				"arabic": q.Arabic,
				"french": q.French,
			},
		})
	}

	return parsed, nil
}

// QuranWordsImporter handles the Quranic vocabulary format (for parsing only, no database access)
type QuranWordsImporter struct{}

// GetName returns the identifier for this data source
func (i *QuranWordsImporter) GetName() string {
	return "quran_words"
}

// GetType returns the data source type
func (i *QuranWordsImporter) GetType() string {
	return "vocabulary"
}

// GetRandomAnswers is not supported for the importer (use QuranWordsDataSource instead)
func (i *QuranWordsImporter) GetRandomAnswers(ctx context.Context, excludeQuestionID int64, count int, filters map[string]interface{}) ([]string, error) {
	return nil, fmt.Errorf("GetRandomAnswers not supported for QuranWordsImporter, use QuranWordsDataSource with repos")
}

func (i *QuranWordsImporter) CanHandle(filename string, rawData []byte) bool {
	// Check if it has the expected structure
	var data struct {
		Deck struct {
			ID string `json:"id"`
		} `json:"deck"`
		Categories map[string]string `json:"cats"`
	}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return false
	}
	return data.Deck.ID != "" && data.Categories != nil
}

func (i *QuranWordsImporter) Parse(rawData []byte) (*ParsedDeck, error) {
	var data struct {
		Deck struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Version int    `json:"version"`
		} `json:"deck"`
		Categories map[string]string `json:"cats"`
		Questions  []struct {
			ID       string `json:"id"`
			Category string `json:"cat"`
			French   string `json:"fr"`
			Arabic   string `json:"ar"`
		} `json:"items"`
	}

	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	parsed := &ParsedDeck{
		DeckKey:    data.Deck.ID,
		Title:      data.Deck.Title,
		Version:    data.Deck.Version,
		Categories: data.Categories,
		Questions:  make([]ParsedQuestion, 0, len(data.Questions)),
	}

	for _, q := range data.Questions {
		parsed.Questions = append(parsed.Questions, ParsedQuestion{
			ID:            q.ID,
			CategoryKey:   q.Category,
			QuestionText:  q.Arabic,
			CorrectAnswer: q.French,
			AdditionalData: map[string]string{
				"arabic": q.Arabic,
				"french": q.French,
			},
		})
	}

	return parsed, nil
}
