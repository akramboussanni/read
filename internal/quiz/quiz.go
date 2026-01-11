package quiz

import (
	"context"
	"fmt"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/akramboussanni/gocode/internal/utils"
)

// Service handles quiz business logic
type QuizService struct {
	repos              *repo.Repos
	answerGenerator    *AnswerGenerator
	dataSourceRegistry *DataSourceRegistry
}

// NewQuizService creates a new quiz service
func NewQuizService(repos *repo.Repos) *QuizService {
	// Initialize data source registry
	registry := NewDataSourceRegistry()

	// Register available data sources
	quranWordsDS := NewQuranWordsDataSource(repos)
	registry.Register(quranWordsDS)
	// Future: register more data sources here

	// Initialize answer generator
	answerGen := NewAnswerGenerator(repos, registry)

	return &QuizService{
		repos:              repos,
		answerGenerator:    answerGen,
		dataSourceRegistry: registry,
	}
}

// GetDeck retrieves a deck by its key
func (s *QuizService) GetDeck(ctx context.Context, deckKey string) (*model.Deck, error) {
	return s.repos.Deck.GetByKey(ctx, deckKey)
}

// GetCategories retrieves all categories for a deck
func (s *QuizService) GetCategories(ctx context.Context, deckID int64) ([]*model.Category, error) {
	// TODO: Implement category retrieval
	return nil, nil
}

// CreateQuestionRequest represents a request to create a new question
type CreateQuestionRequest struct {
	DeckID        int64
	CategoryID    int64
	QuestionKey   string
	QuestionText  string
	CorrectAnswer string
	QuestionType  string // "multiple_choice", "written", "true_false"

	// For MCQ
	AnswerMode    string   // "manual" or "auto_generated"
	ManualAnswers []string // Used if AnswerMode is "manual"

	// For auto-generated answers
	GenerationRule   string // "same_category", "same_deck", "different_category", "random"
	DataSourceName   string // Which data source to use (e.g., "quran_words")
	WrongAnswerCount int    // How many wrong answers to generate (default: 3)
}

// CreateQuestion creates a new question with answer options
func (s *QuizService) CreateQuestion(ctx context.Context, req CreateQuestionRequest) (*model.Question, error) {
	// Validate request
	if req.QuestionText == "" || req.CorrectAnswer == "" {
		return nil, fmt.Errorf("question text and correct answer are required")
	}

	if req.QuestionType == "" {
		req.QuestionType = "multiple_choice"
	}

	// Create the question
	now := time.Now().Unix()
	question := &model.Question{
		ID:            utils.GenerateSnowflakeID(),
		DeckID:        req.DeckID,
		CategoryID:    req.CategoryID,
		QuestionKey:   req.QuestionKey,
		QuestionText:  req.QuestionText,
		CorrectAnswer: req.CorrectAnswer,
		QuestionType:  req.QuestionType,
		Points:        1,
		IsActive:      true,
		CreatedAt:     now,
	}

	if err := s.repos.Question.Create(ctx, question); err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	// Create answer options for MCQ
	if req.QuestionType == "multiple_choice" {
		if req.AnswerMode == "manual" && len(req.ManualAnswers) > 0 {
			// Create manual options
			if err := s.answerGenerator.CreateManualOptions(ctx, question, req.ManualAnswers); err != nil {
				return nil, fmt.Errorf("failed to create manual options: %w", err)
			}
		} else if req.AnswerMode == "auto_generated" {
			// Generate automatic options
			count := req.WrongAnswerCount
			if count == 0 {
				count = 3 // Default to 3 wrong answers
			}

			config := GenerationConfig{
				Count:          count,
				Rule:           req.GenerationRule,
				DataSourceName: req.DataSourceName,
			}

			if err := s.answerGenerator.GenerateAndSaveOptions(ctx, question, config); err != nil {
				return nil, fmt.Errorf("failed to generate options: %w", err)
			}
		}
	}

	return question, nil
}

// CreateQuizRequest represents a request to create a quiz
type CreateQuizRequest struct {
	Title              string
	Description        string
	DeckID             int64
	CategorySelections []CategorySelection // Categories and how many questions from each
	QuestionMode       string
	TimeLimit          *int
	PassPercentage     *int
	ShuffleQuestions   bool
	IsSystem           bool
	CreatedBy          *int64
}

type CategorySelection struct {
	CategoryID    int64
	QuestionCount int
}

// CreateQuiz creates a new quiz configuration
func (s *QuizService) CreateQuiz(ctx context.Context, req CreateQuizRequest, isAdmin bool) (*model.Quiz, error) {
	now := time.Now().Unix()

	quiz := &model.Quiz{
		ID:               utils.GenerateSnowflakeID(),
		Title:            req.Title,
		Description:      req.Description,
		DeckID:           req.DeckID,
		TimeLimit:        req.TimeLimit,
		PassPercentage:   req.PassPercentage,
		ShuffleQuestions: req.ShuffleQuestions,
		QuestionMode:     req.QuestionMode,
		IsSystem:         req.IsSystem,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        now,
		IsActive:         true,
		GivesCoins:       req.IsSystem || isAdmin, // Only system quizzes or admin-created give coins
		CoinReward:       0,                       // Can be set separately
	}

	// If quiz gives coins, set default reward
	if quiz.GivesCoins {
		quiz.CoinReward = 10 // Default: 10 coins per correct answer
	}

	// Save quiz to database
	if err := s.repos.Quiz.Create(ctx, quiz); err != nil {
		return nil, fmt.Errorf("failed to create quiz: %w", err)
	}

	// Save category selections
	for _, cs := range req.CategorySelections {
		selection := &model.QuizCategorySelection{
			QuizID:        quiz.ID,
			CategoryID:    cs.CategoryID,
			QuestionCount: cs.QuestionCount,
		}
		if err := s.repos.QuizCategorySelection.Create(ctx, selection); err != nil {
			return nil, fmt.Errorf("failed to create category selection: %w", err)
		}
	}

	return quiz, nil
}

// GenerateQuiz generates a new quiz attempt for a user
func (s *QuizService) GenerateQuiz(ctx context.Context, userID, quizID int64) (*model.QuizAttempt, error) {
	// TODO: Implement quiz generation with random questions and options
	return nil, nil
}

// AwardCoins awards coins to a user for completing a quiz
func (s *QuizService) AwardCoins(ctx context.Context, userID int64, amount int, attemptID int64) error {
	if amount <= 0 {
		return nil
	}

	now := time.Now().Unix()

	// Get or create user coins record
	userCoins, err := s.repos.Coin.GetUserCoins(ctx, userID)
	if err != nil {
		// Create new record
		userCoins = &model.UserCoins{
			UserID:         userID,
			Balance:        0,
			LifetimeEarned: 0,
			LastUpdated:    now,
		}
	}

	// Update balance
	userCoins.Balance += amount
	userCoins.LifetimeEarned += amount
	userCoins.LastUpdated = now

	if err := s.repos.Coin.CreateOrUpdateCoins(ctx, userCoins); err != nil {
		return fmt.Errorf("failed to update user coins: %w", err)
	}

	// Create transaction record
	refType := "quiz_attempt"
	transaction := &model.CoinTransaction{
		ID:              utils.GenerateSnowflakeID(),
		UserID:          userID,
		Amount:          amount,
		TransactionType: "quiz_reward",
		ReferenceType:   &refType,
		ReferenceID:     &attemptID,
		CreatedAt:       now,
	}

	if err := s.repos.Coin.AddTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}
