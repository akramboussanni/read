package quiz

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
)

// DeckService manages deck loading and caching strategy
type DeckService struct {
	repos        *repo.Repos
	cache        map[int64]*CachedDeck
	cacheMux     sync.RWMutex
	maxCacheSize int
	cacheTTL     time.Duration
}

// CachedDeck represents a deck in memory cache
type CachedDeck struct {
	Deck        *ParsedDeck
	LoadedAt    time.Time
	AccessCount int
}

// NewDeckService creates a new deck service
func NewDeckService(repos *repo.Repos) *DeckService {
	return &DeckService{
		repos:        repos,
		cache:        make(map[int64]*CachedDeck),
		maxCacheSize: 10,               // Cache up to 10 decks in memory
		cacheTTL:     30 * time.Minute, // Cache for 30 minutes
	}
}

// GetDeck retrieves a deck with intelligent caching strategy
func (s *DeckService) GetDeck(ctx context.Context, deckID int64) (*ParsedDeck, error) {
	// First, try memory cache
	if deck := s.getFromMemoryCache(deckID); deck != nil {
		applog.Info("Deck %d loaded from memory cache", deckID)
		return deck, nil
	}

	// Try database cache
	if deck := s.getFromDatabaseCache(ctx, deckID); deck != nil {
		// Ensure DeckID is set (handle potential cache inconsistencies)
		if deck.DeckID == 0 {
			deck.DeckID = deckID
		}

		// Store in memory cache for faster future access
		s.storeInMemoryCache(deckID, deck)
		applog.Info("Deck %d loaded from database cache", deckID)
		return deck, nil
	}

	// Load from database and cache both places
	deck, err := s.loadFromDatabase(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Cache in both memory and database
	s.storeInMemoryCache(deckID, deck)
	if err := s.storeInDatabaseCache(ctx, deckID, deck); err != nil {
		applog.Warn("Failed to cache deck %d in database: %v", deckID, err)
	}

	applog.Debug("Deck %d loaded from database", deckID)
	return deck, nil
}

// getFromMemoryCache retrieves deck from memory cache if valid
func (s *DeckService) getFromMemoryCache(deckID int64) *ParsedDeck {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	cached, exists := s.cache[deckID]
	if !exists {
		return nil
	}

	// Check if cache is still valid
	if time.Since(cached.LoadedAt) > s.cacheTTL {
		// Remove expired cache
		delete(s.cache, deckID)
		return nil
	}

	// Update access count for LRU-style eviction
	cached.AccessCount++
	return cached.Deck
}

// storeInMemoryCache stores deck in memory cache with LRU eviction
func (s *DeckService) storeInMemoryCache(deckID int64, deck *ParsedDeck) {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	// If cache is full, remove least recently used
	if len(s.cache) >= s.maxCacheSize {
		s.evictLRU()
	}

	s.cache[deckID] = &CachedDeck{
		Deck:        deck,
		LoadedAt:    time.Now(),
		AccessCount: 1,
	}
}

// evictLRU removes the least recently used cache entry
func (s *DeckService) evictLRU() {
	var oldestID int64
	var oldestTime time.Time
	var oldestAccess int

	for id, cached := range s.cache {
		if oldestTime.IsZero() || cached.AccessCount < oldestAccess ||
			(cached.AccessCount == oldestAccess && cached.LoadedAt.Before(oldestTime)) {
			oldestID = id
			oldestTime = cached.LoadedAt
			oldestAccess = cached.AccessCount
		}
	}

	if oldestID != 0 {
		delete(s.cache, oldestID)
		applog.Debug("Evicted deck %d from memory cache (LRU)", oldestID)
	}
}

// getFromDatabaseCache retrieves deck from database cache
func (s *DeckService) getFromDatabaseCache(ctx context.Context, deckID int64) *ParsedDeck {
	cache, err := s.repos.DeckCache.GetByDeckID(ctx, deckID)
	if err != nil {
		if err != sql.ErrNoRows {
			applog.Warn("Failed to load deck cache: %v", err)
		}
		return nil
	}

	// Only use cache if it's less than 1 hour old
	if time.Now().Unix()-cache.LastUpdated > 3600 {
		return nil
	}

	var deck ParsedDeck
	if err := json.Unmarshal([]byte(cache.CachedData), &deck); err != nil {
		applog.Warn("Failed to unmarshal cached deck: %v", err)
		return nil
	}

	return &deck
}

// storeInDatabaseCache stores deck in database cache
func (s *DeckService) storeInDatabaseCache(ctx context.Context, deckID int64, deck *ParsedDeck) error {
	data, err := json.Marshal(deck)
	if err != nil {
		return err
	}

	cache := &model.DeckCache{
		DeckID:       deckID,
		CachedData:   string(data),
		CacheVersion: 1,
		LastUpdated:  time.Now().Unix(),
	}

	return s.repos.DeckCache.Upsert(ctx, cache)
}

// loadFromDatabase loads deck from database tables
func (s *DeckService) loadFromDatabase(ctx context.Context, deckID int64) (*ParsedDeck, error) {
	// Load deck metadata
	deck, err := s.loadDeckMetadata(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Load categories
	categories, err := s.loadCategories(ctx, deckID)
	if err != nil {
		return nil, err
	}
	deck.Categories = categories

	// Load questions/entries
	questions, err := s.loadQuestions(ctx, deckID)
	if err != nil {
		return nil, err
	}
	deck.Questions = questions

	return deck, nil
}

// loadDeckMetadata loads basic deck information
func (s *DeckService) loadDeckMetadata(ctx context.Context, deckID int64) (*ParsedDeck, error) {
	deckModel, err := s.repos.Deck.GetByID(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Convert model.Deck to ParsedDeck
	deck := &ParsedDeck{
		DeckID:  deckModel.ID,
		DeckKey: deckModel.DeckKey,
		Title:   deckModel.Title,
		Version: deckModel.Version,
		Metadata: DeckMetadata{
			DeckType: deckModel.DeckType,
		},
		Categories: make(map[string]string),
	}

	// Unmarshal JSON fields from database
	if deckModel.LanguagePair != "" {
		if err := json.Unmarshal([]byte(deckModel.LanguagePair), &deck.Metadata.LanguagePair); err != nil {
			applog.Warn("Failed to unmarshal language pair: %v", err)
		}
	}
	if deckModel.SupportedQuestionTypes != "" {
		if err := json.Unmarshal([]byte(deckModel.SupportedQuestionTypes), &deck.Metadata.SupportedQuestionTypes); err != nil {
			applog.Warn("Failed to unmarshal supported question types: %v", err)
		}
	}

	// Parse additional metadata if present (may override fields)
	if deckModel.DeckMetadata != "" {
		if err := json.Unmarshal([]byte(deckModel.DeckMetadata), &deck.Metadata); err != nil {
			applog.Warn("Failed to unmarshal deck metadata: %v", err)
		}
	}

	return deck, nil
}

// loadCategories loads deck categories
func (s *DeckService) loadCategories(ctx context.Context, deckID int64) (map[string]string, error) {
	categories, err := s.repos.Category.GetByDeckID(ctx, deckID)
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]string)
	for _, cat := range categories {
		categoryMap[cat.CategoryKey] = cat.Title
	}

	return categoryMap, nil
}

// loadQuestions loads deck entries as questions
func (s *DeckService) loadQuestions(ctx context.Context, deckID int64) ([]ParsedQuestion, error) {
	entries, err := s.repos.DeckEntry.GetByDeckID(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Get category models for lookup
	categoryModels, err := s.repos.Category.GetByDeckID(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Create reverse map: category_id -> category_key
	categoryIDToKey := make(map[int64]string)
	for _, cat := range categoryModels {
		categoryIDToKey[cat.ID] = cat.CategoryKey
	}

	var questions []ParsedQuestion
	for _, entry := range entries {
		categoryKey, exists := categoryIDToKey[entry.CategoryID]
		if !exists {
			continue // Skip entries with invalid category
		}

		var entryData map[string]interface{}
		if err := json.Unmarshal([]byte(entry.EntryData), &entryData); err != nil {
			continue // Skip invalid entries
		}

		// Extract source and target text (universal approach)
		sourceText := getStringValue(entryData, "source_text")
		targetText := getStringValue(entryData, "target_text")

		question := ParsedQuestion{
			ID:             entry.EntryKey,
			CategoryKey:    categoryKey,
			QuestionText:   sourceText,
			CorrectAnswer:  targetText,
			AdditionalData: entryData,
		}

		questions = append(questions, question)
	}

	return questions, nil
}

// getStringValue safely extracts string value from map
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ClearCache clears all cached decks
func (s *DeckService) ClearCache() {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()
	s.cache = make(map[int64]*CachedDeck)
	applog.Info("Cleared deck memory cache")
}

// GetCacheStats returns cache statistics
func (s *DeckService) GetCacheStats() map[string]interface{} {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	return map[string]interface{}{
		"cached_decks":      len(s.cache),
		"max_cache_size":    s.maxCacheSize,
		"cache_ttl_minutes": s.cacheTTL.Minutes(),
	}
}
