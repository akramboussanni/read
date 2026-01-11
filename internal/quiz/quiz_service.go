package quiz

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/repo"
)

// SeedAllDecks seeds all universal deck files into the database
func SeedAllDecks(ctx context.Context, repos *repo.Repos) error {
	applog.Info("Starting universal deck seeding...")

	// Use the data directory for deck files
	baseDir := filepath.Join("internal", "quiz", "data")

	seeder := NewUniversalDeckSeeder(repos, baseDir)

	if err := seeder.SeedDecks(ctx); err != nil {
		return err
	}

	applog.Info("Universal deck seeding completed successfully")
	return nil
}

// QuizService provides the main quiz functionality
type QuizService struct {
	importer      *UniversalImporter
	quizGenerator *QuizGenerator
}

// NewQuizService creates a new quiz service
func NewQuizService(repos *repo.Repos) *QuizService {
	return &QuizService{
		importer:      NewUniversalImporter(),
		quizGenerator: NewQuizGenerator(),
	}
}

// LoadDeck loads a deck from file
func (s *QuizService) LoadDeck(filePath string) (*ParsedDeck, error) {
	return s.importer.ImportFromFile(filePath)
}

// CreateQuiz creates a quiz from multiple loaded decks
func (s *QuizService) CreateQuiz(ctx context.Context, config QuizConfig, decks []*ParsedDeck) (*Quiz, error) {
	return s.quizGenerator.GenerateQuiz(ctx, config, decks)
}

// ValidateAnswer validates a user's answer for a question
func (s *QuizService) ValidateAnswer(question QuizQuestion, userAnswer string) bool {
	switch question.QuestionType {
	case QuestionTypeMCQ:
		return userAnswer == question.CorrectAnswer
	case QuestionTypeWriteWord, QuestionTypeTranslate:
		// Simple string comparison, could be enhanced with fuzzy matching
		return strings.ToLower(strings.TrimSpace(userAnswer)) == strings.ToLower(strings.TrimSpace(question.CorrectAnswer))
	case QuestionTypeCustom:
		return userAnswer == question.CorrectAnswer
	default:
		return false
	}
}

// GetQuizProgress returns the progress of a quiz
func (s *QuizService) GetQuizProgress(quiz *Quiz) (answered int, total int, percentage float64) {
	total = len(quiz.Questions)
	answered = 0
	for _, q := range quiz.Questions {
		if q.UserAnswer != nil {
			answered++
		}
	}
	if total > 0 {
		percentage = float64(answered) / float64(total) * 100
	}
	return
}

// CompleteQuiz marks a quiz as completed
func (s *QuizService) CompleteQuiz(quiz *Quiz) {
	now := time.Now()
	quiz.CompletedAt = &now
}
