package quiz

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/jmoiron/sqlx"
)

//go:embed data/*.json
var embeddedData embed.FS

// DeckImporter is an interface for importing different deck formats
type DeckImporter interface {
	CanHandle(filename string, rawData []byte) bool
	Parse(rawData []byte) (*ParsedDeck, error)
}

// ParsedDeck is a normalized deck structure that all importers produce
type ParsedDeck struct {
	DeckKey    string
	Title      string
	Version    int
	Categories map[string]string // key -> title
	Questions  []ParsedQuestion
}

type ParsedQuestion struct {
	ID             string
	CategoryKey    string
	QuestionText   string
	CorrectAnswer  string
	AdditionalData map[string]string // For language-specific or custom fields
}

// Registry of all available importers
var importers = []DeckImporter{
	&QuranWordsImporter{},
	// Add more importers here for different formats
	// &SimpleQAImporter{},
	// &MultipleChoiceImporter{},
}

// LoadEmbeddedDeck loads and parses a deck using the appropriate importer
func LoadEmbeddedDeck(filename string) (*ParsedDeck, error) {
	path := "data/" + filename
	rawData, err := embeddedData.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded file %s: %w", filename, err)
	}

	// Create a temporary registry to find the right importer
	registry := NewDataSourceRegistry()
	registry.Register(&QuranWordsImporter{})

	dataSource := registry.FindImporter(filename, rawData)
	if dataSource == nil {
		return nil, fmt.Errorf("no importer found for file %s", filename)
	}

	return dataSource.Parse(rawData)
}

// SeedAllDecks imports all embedded JSON files into the database
func SeedAllDecks(ctx context.Context, db *sqlx.DB, repos *repo.Repos) error {
	// Check if already seeded
	var count int
	err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM quiz_decks")
	if err != nil {
		return fmt.Errorf("failed to check seed status: %w", err)
	}

	if count > 0 {
		log.Println("Quiz data already seeded, skipping")
		return nil
	}

	log.Println("Seeding quiz data from embedded files...")

	// Auto-discover all JSON files in data directory
	entries, err := embeddedData.ReadDir("data")
	if err != nil {
		return fmt.Errorf("failed to read embedded data directory: %w", err)
	}

	imported := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Skip non-JSON files
		if !strings.HasSuffix(filename, ".json") {
			continue
		}

		// Skip README or other metadata files
		if strings.HasPrefix(filename, ".") || strings.HasPrefix(filename, "README") {
			continue
		}

		if err := ImportDeck(ctx, db, repos, filename); err != nil {
			log.Printf("Warning: Failed to import %s: %v", filename, err)
			continue
		}

		imported++
	}

	log.Printf("✓ Quiz data seeding completed (%d decks imported)", imported)
	return nil
}

// ImportDeck imports a single deck into the database
func ImportDeck(ctx context.Context, db *sqlx.DB, repos *repo.Repos, filename string) error {
	parsedDeck, err := LoadEmbeddedDeck(filename)
	if err != nil {
		return err
	}

	// Check if this specific deck already exists
	existingDeck, _ := repos.Deck.GetByKey(ctx, parsedDeck.DeckKey)
	if existingDeck != nil {
		log.Printf("Deck '%s' already exists, skipping", parsedDeck.Title)
		return nil
	}

	log.Printf("Importing deck '%s' from %s...", parsedDeck.Title, filename)

	// Create deck
	deckID := utils.GenerateSnowflakeID()
	deck := &model.Deck{
		ID:         deckID,
		DeckKey:    parsedDeck.DeckKey,
		Title:      parsedDeck.Title,
		Version:    parsedDeck.Version,
		SourceFile: filename,
		IsSystem:   true,
		CreatedAt:  time.Now().Unix(),
	}

	if err := repos.Deck.Create(ctx, deck); err != nil {
		return fmt.Errorf("failed to create deck: %w", err)
	}

	// Create categories
	categoryMap := make(map[string]int64)
	order := 0
	for categoryKey, categoryTitle := range parsedDeck.Categories {
		categoryID := utils.GenerateSnowflakeID()
		category := &model.Category{
			ID:           categoryID,
			DeckID:       deckID,
			CategoryKey:  categoryKey,
			Title:        categoryTitle,
			DisplayOrder: order,
			CreatedAt:    time.Now().Unix(),
		}

		if err := repos.Category.Create(ctx, category); err != nil {
			return fmt.Errorf("failed to create category %s: %w", categoryTitle, err)
		}

		categoryMap[categoryKey] = categoryID
		order++
	}

	// Bulk insert questions using transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	log.Printf("  Importing %d questions...", len(parsedDeck.Questions))

	for _, question := range parsedDeck.Questions {
		categoryID, ok := categoryMap[question.CategoryKey]
		if !ok {
			log.Printf("  Warning: unknown category '%s' for item '%s'", question.CategoryKey, question.ID)
			continue
		}

		questionID := utils.GenerateSnowflakeID()

		// Build question model
		q := &model.Question{
			ID:            questionID,
			DeckID:        deckID,
			CategoryID:    categoryID,
			QuestionKey:   question.ID,
			QuestionText:  question.QuestionText,
			CorrectAnswer: question.CorrectAnswer,
			QuestionType:  "multiple_choice",
			Points:        1,
			CreatedAt:     time.Now().Unix(),
			IsActive:      true,
		}

		// Handle additional language-specific fields
		if arabic, ok := question.AdditionalData["arabic"]; ok {
			q.Arabic = arabic
		}
		if french, ok := question.AdditionalData["french"]; ok {
			q.French = french
		}

		// Insert question
		_, err := tx.NamedExecContext(ctx, `
			INSERT INTO questions (
				id, deck_id, category_id, question_key, 
				question_text, correct_answer, arabic, french,
				question_type, points, created_at, is_active
			) VALUES (
				:id, :deck_id, :category_id, :question_key,
				:question_text, :correct_answer, :arabic, :french,
				:question_type, :points, :created_at, :is_active
			)
		`, q)

		if err != nil {
			return fmt.Errorf("failed to insert question %s: %w", question.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✓ Imported deck '%s': %d categories, %d questions",
		parsedDeck.Title, len(parsedDeck.Categories), len(parsedDeck.Questions))

	return nil
}
