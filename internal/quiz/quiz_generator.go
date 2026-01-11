package quiz

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// QuizGenerator generates quizzes based on configuration
type QuizGenerator struct {
	registry *QuestionGeneratorRegistry
}

// NewQuizGenerator creates a new quiz generator
func NewQuizGenerator() *QuizGenerator {
	return &QuizGenerator{
		registry: NewQuestionGeneratorRegistry(),
	}
}

// GenerateQuiz generates a quiz based on the config
func (g *QuizGenerator) GenerateQuiz(ctx context.Context, config QuizConfig, parsedDecks []*ParsedDeck) (*Quiz, error) {
	if len(parsedDecks) == 0 {
		return nil, fmt.Errorf("at least one parsed deck is required")
	}

	questions := make([]QuizQuestion, 0, config.QuestionCount)

	// Collect all eligible entries from all decks
	var allEligibleEntries []struct {
		entry *ParsedQuestion
		deck  *ParsedDeck
	}

	for _, deck := range parsedDecks {
		// Find the deck selection for this deck
		var selectedCategories []string
		for _, selection := range config.DeckSelections {
			if selection.DeckID == deck.DeckID {
				selectedCategories = selection.Categories
				break
			}
		}

		// If no selection found for this deck, skip it
		if selectedCategories == nil {
			continue
		}

		// Filter entries by categories if specified
		var eligibleEntries []ParsedQuestion
		if len(selectedCategories) > 0 {
			for _, q := range deck.Questions {
				for _, cat := range selectedCategories {
					if q.CategoryKey == cat {
						eligibleEntries = append(eligibleEntries, q)
						break
					}
				}
			}
		} else {
			eligibleEntries = deck.Questions
		}

		// Add entries with their deck reference
		for i := range eligibleEntries {
			allEligibleEntries = append(allEligibleEntries, struct {
				entry *ParsedQuestion
				deck  *ParsedDeck
			}{&eligibleEntries[i], deck})
		}
	}

	if len(allEligibleEntries) == 0 {
		return nil, fmt.Errorf("no eligible entries found for the given deck selections")
	}

	// Function to get random answers from all decks
	getRandomAnswers := func(count int, category string) []string {
		var answers []string
		for _, deckEntry := range allEligibleEntries {
			if deckEntry.entry.CategoryKey == category && deckEntry.entry.CorrectAnswer != "" {
				answers = append(answers, deckEntry.entry.CorrectAnswer)
			}
		}
		// Shuffle and take count
		rand.Shuffle(len(answers), func(i, j int) {
			answers[i], answers[j] = answers[j], answers[i]
		})
		if len(answers) > count {
			answers = answers[:count]
		}
		return answers
	}

	// Generate questions from entries
	entryCount := len(allEligibleEntries)
	generatedCount := 0

	for generatedCount < config.QuestionCount && entryCount > 0 {
		// Randomly select an entry
		idx := rand.Intn(entryCount)
		entryInfo := allEligibleEntries[idx]

		// Get the generator for this deck
		generator := g.registry.GetGeneratorForDeck(entryInfo.deck.Metadata)
		if generator == nil {
			continue // Skip this entry
		}

		// Convert ParsedQuestion to DeckEntry (simplified)
		deckEntry := DeckEntry{
			ID:       entryInfo.entry.ID,
			Category: entryInfo.entry.CategoryKey,
			Data:     entryInfo.entry.AdditionalData,
		}

		// Randomly select question type and direction
		qt := config.QuestionTypes[rand.Intn(len(config.QuestionTypes))]
		dir := config.Directions[rand.Intn(len(config.Directions))]

		if !generator.CanGenerate(qt, dir, entryInfo.deck.Metadata) {
			continue // Try another
		}

		genQuestion, err := generator.Generate(deckEntry, qt, dir, getRandomAnswers)
		if err != nil {
			continue // Skip this entry
		}

		quizQuestion := QuizQuestion{
			ID:            int64(generatedCount + 1), // Simple ID
			QuestionText:  genQuestion.QuestionText,
			CorrectAnswer: genQuestion.CorrectAnswer,
			Options:       genQuestion.Options,
			QuestionType:  genQuestion.QuestionType,
			Direction:     genQuestion.Direction,
		}

		questions = append(questions, quizQuestion)
		generatedCount++

		// Remove this entry to avoid duplicates
		allEligibleEntries = append(allEligibleEntries[:idx], allEligibleEntries[idx+1:]...)
		entryCount--
	}

	// Add custom questions
	for _, custom := range config.CustomQuestions {
		quizQuestion := QuizQuestion{
			ID:            int64(len(questions) + 1),
			QuestionText:  custom.QuestionText,
			CorrectAnswer: custom.CorrectAnswer,
			QuestionType:  custom.QuestionType,
		}

		if custom.QuestionType == QuestionTypeMCQ {
			quizQuestion.Options = append([]string{custom.CorrectAnswer}, custom.WrongAnswers...)
			rand.Shuffle(len(quizQuestion.Options), func(i, j int) {
				quizQuestion.Options[i], quizQuestion.Options[j] = quizQuestion.Options[j], quizQuestion.Options[i]
			})
		}

		questions = append(questions, quizQuestion)
	}

	// Shuffle the questions
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	// Trim to QuestionCount if we have more
	if len(questions) > config.QuestionCount {
		questions = questions[:config.QuestionCount]
	}

	quiz := &Quiz{
		Config:       config,
		Questions:    questions,
		CurrentIndex: 0,
		StartedAt:    time.Now(),
	}

	return quiz, nil
}
