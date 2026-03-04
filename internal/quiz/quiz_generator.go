package quiz

import (
	"context"
	"fmt"
	"math/rand"
	"os"
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
		fmt.Fprintf(os.Stderr, "DEBUG: Processing deck %d. Questions in deck: %d\n", deck.DeckID, len(deck.Questions))
		// Find the deck selection for this deck
		var selectedCategories []string
		found := false
		for _, selection := range config.DeckSelections {
			if selection.DeckID == deck.DeckID {
				selectedCategories = selection.Categories
				found = true
				break
			}
		}

		// If no selection found for this deck, skip it
		if !found {
			fmt.Fprintf(os.Stderr, "DEBUG: Deck %d not in config selections\n", deck.DeckID)
			continue
		}

		fmt.Fprintf(os.Stderr, "DEBUG: Deck %d selected categories: %v\n", deck.DeckID, selectedCategories)

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

		fmt.Fprintf(os.Stderr, "DEBUG: Deck %d eligible entries found: %d\n", deck.DeckID, len(eligibleEntries))

		// Add entries with their deck reference
		for i := range eligibleEntries {
			allEligibleEntries = append(allEligibleEntries, struct {
				entry *ParsedQuestion
				deck  *ParsedDeck
			}{&eligibleEntries[i], deck})
		}
	}

	if len(allEligibleEntries) == 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: Total eligible entries is 0. Aborting.\n")
		return nil, fmt.Errorf("no eligible entries found for the given deck selections")
	}

	// Function to get random answers from all decks
	getRandomAnswers := func(count int, category string) []string {
		var sameCategoryAnswers []string
		var otherCategoryAnswers []string

		for _, deckEntry := range allEligibleEntries {
			if deckEntry.entry.CorrectAnswer != "" {
				if deckEntry.entry.CategoryKey == category {
					sameCategoryAnswers = append(sameCategoryAnswers, deckEntry.entry.CorrectAnswer)
				} else {
					otherCategoryAnswers = append(otherCategoryAnswers, deckEntry.entry.CorrectAnswer)
				}
			}
		}

		// Shuffle both
		rand.Shuffle(len(sameCategoryAnswers), func(i, j int) {
			sameCategoryAnswers[i], sameCategoryAnswers[j] = sameCategoryAnswers[j], sameCategoryAnswers[i]
		})
		rand.Shuffle(len(otherCategoryAnswers), func(i, j int) {
			otherCategoryAnswers[i], otherCategoryAnswers[j] = otherCategoryAnswers[j], otherCategoryAnswers[i]
		})

		// Combine, prioritizing same category
		answers := append(sameCategoryAnswers, otherCategoryAnswers...)

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

		genQuestion, err := generator.Generate(deckEntry, qt, dir, entryInfo.deck.Metadata, getRandomAnswers)
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
