package quiz

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/akramboussanni/gocode/internal/utils"
)

//go:embed data/*.json
var embeddedDecks embed.FS

// SeedAllDecks is a convenience function to seed all embedded decks
func SeedAllDecks(ctx context.Context, repos *repo.Repos) error {
	fsys, err := fs.Sub(embeddedDecks, "data")
	if err != nil {
		return fmt.Errorf("failed to get decks sub-fs: %w", err)
	}

	seeder := NewUniversalDeckSeeder(repos, fsys)
	return seeder.SeedDecks(ctx)
}

// UniversalDeckSeeder handles seeding universal decks into the database
type UniversalDeckSeeder struct {
	repos  *repo.Repos
	seeder *UniversalImporter
	dataFS fs.FS
}

// NewUniversalDeckSeeder creates a new seeder
func NewUniversalDeckSeeder(repos *repo.Repos, dataFS fs.FS) *UniversalDeckSeeder {
	return &UniversalDeckSeeder{
		repos:  repos,
		seeder: NewUniversalImporter(),
		dataFS: dataFS,
	}
}

// SeedDecks loads all universal deck files and seeds them into the database
func (s *UniversalDeckSeeder) SeedDecks(ctx context.Context) error {
	files, err := s.findDeckFiles()
	if err != nil {
		return fmt.Errorf("failed to find deck files: %w", err)
	}

	seeded := 0
	for _, file := range files {
		if err := s.seedDeckFile(ctx, file); err != nil {
			applog.Errorf("Failed to seed deck file %s: %v", file, err)
			continue
		}
		seeded++
	}

	applog.Infof("Successfully seeded %d deck files", seeded)
	return nil
}

// findDeckFiles finds all .json files in the decks directory
func (s *UniversalDeckSeeder) findDeckFiles() ([]string, error) {
	var files []string

	err := fs.WalkDir(s.dataFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".json" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// seedDeckFile loads a single deck file and stores it in the database
func (s *UniversalDeckSeeder) seedDeckFile(ctx context.Context, filePath string) error {
	applog.Infof("Seeding deck file: %s", filePath)

	// Load and parse the deck from embedded FS
	rawData, err := fs.ReadFile(s.dataFS, filePath)
	if err != nil {
		return fmt.Errorf("failed to read deck file: %w", err)
	}

	deck, err := s.seeder.ImportFromBytes(filePath, rawData)
	if err != nil {
		return fmt.Errorf("failed to parse deck file: %w", err)
	}

	// Check if deck already exists
	existing, err := s.repos.Deck.GetByKeyAndVersion(ctx, deck.DeckKey, deck.Version)
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no rows") {
		return fmt.Errorf("failed to check existing deck: %w", err)
	}

	if existing != nil {
		applog.Infof("Deck %s v%d already exists, skipping", deck.DeckKey, deck.Version)
		return nil
	}

	// Convert to database format and save
	return s.saveDeckToDatabase(ctx, deck, filePath)
}

// saveDeckToDatabase saves a parsed deck to the database
func (s *UniversalDeckSeeder) saveDeckToDatabase(ctx context.Context, deck *ParsedDeck, filePath string) error {
	// Convert parsed deck to model
	deckModel := s.convertToDeckModel(deck, filePath)

	// Create deck
	if err := s.repos.Deck.Create(ctx, deckModel); err != nil {
		return fmt.Errorf("failed to create deck: %w", err)
	}

	// Insert categories
	categoryIDMap, err := s.insertCategories(ctx, deckModel.ID, deck)
	if err != nil {
		return fmt.Errorf("failed to insert categories: %w", err)
	}

	// Insert entries
	if err := s.insertEntries(ctx, deckModel.ID, categoryIDMap, deck); err != nil {
		return fmt.Errorf("failed to insert entries: %w", err)
	}

	// Cache the deck data for performance
	if err := s.cacheDeckData(ctx, deckModel.ID, deck); err != nil {
		applog.Warn("Failed to cache deck data: %v", err)
		// Don't fail the seeding for cache issues
	}

	return nil
}

// convertToDeckModel converts a ParsedDeck to a model.Deck
func (s *UniversalDeckSeeder) convertToDeckModel(deck *ParsedDeck, filePath string) *model.Deck {
	metadataJSON, _ := json.Marshal(deck.Metadata)
	langPairJSON, _ := json.Marshal(deck.Metadata.LanguagePair)
	supportedTypesJSON, _ := json.Marshal(deck.Metadata.SupportedQuestionTypes)

	return &model.Deck{
		ID:                     utils.GenerateSnowflakeID(),
		DeckKey:                deck.DeckKey,
		Title:                  deck.Title,
		Version:                deck.Version,
		DeckType:               deck.Metadata.DeckType,
		LanguagePair:           string(langPairJSON),
		SupportedQuestionTypes: string(supportedTypesJSON),
		DefaultDirection:       "", // Removed: now handled per question
		DeckMetadata:           string(metadataJSON),
		SourceFile:             filepath.Base(filePath),
		IsSystem:               true,
		CreatedAt:              time.Now().Unix(),
	}
}

// insertCategories inserts all categories for a deck
func (s *UniversalDeckSeeder) insertCategories(ctx context.Context, deckID int64, deck *ParsedDeck) (map[string]int64, error) {
	categoryIDMap := make(map[string]int64)

	for key, title := range deck.Categories {
		// For now, set default difficulty - could be enhanced to parse from metadata
		difficulty := "intermediate"

		category := &model.Category{
			ID:               utils.GenerateSnowflakeID(),
			DeckID:           deckID,
			CategoryKey:      key,
			Title:            title,
			Difficulty:       difficulty,
			CategoryMetadata: "{}", // Ensure valid JSON for PostgreSQL JSONB
			DisplayOrder:     0,
			CreatedAt:        time.Now().Unix(),
		}

		if err := s.repos.Category.Create(ctx, category); err != nil {
			return nil, err
		}

		categoryIDMap[key] = category.ID
	}

	return categoryIDMap, nil
}

// insertEntries inserts all deck entries
func (s *UniversalDeckSeeder) insertEntries(ctx context.Context, deckID int64, categoryIDMap map[string]int64, deck *ParsedDeck) error {
	for _, entry := range deck.Questions {
		categoryID, exists := categoryIDMap[entry.CategoryKey]
		if !exists {
			return fmt.Errorf("category %s not found for entry %s", entry.CategoryKey, entry.ID)
		}

		// Convert additional data to JSON
		dataJSON, err := json.Marshal(entry.AdditionalData)
		if err != nil {
			return fmt.Errorf("failed to marshal entry data: %w", err)
		}

		// For now, empty tags array - could be enhanced
		tagsJSON := "[]"

		deckEntry := &model.DeckEntry{
			ID:         utils.GenerateSnowflakeID(),
			DeckID:     deckID,
			CategoryID: categoryID,
			EntryKey:   entry.ID,
			EntryData:  string(dataJSON),
			Tags:       tagsJSON,
			CreatedAt:  time.Now().Unix(),
		}

		if err := s.repos.DeckEntry.Create(ctx, deckEntry); err != nil {
			return err
		}
	}

	return nil
}

// cacheDeckData caches the complete parsed deck for performance
func (s *UniversalDeckSeeder) cacheDeckData(ctx context.Context, deckID int64, deck *ParsedDeck) error {
	// Convert entire deck to JSON for caching
	deckJSON, err := json.Marshal(deck)
	if err != nil {
		return err
	}

	cache := &model.DeckCache{
		DeckID:       deckID,
		CachedData:   string(deckJSON),
		CacheVersion: 1,
		LastUpdated:  time.Now().Unix(),
	}

	return s.repos.DeckCache.Upsert(ctx, cache)
}
