package quiz

import (
	"fmt"
	"math/rand"
)

// QuestionGenerator interface for generating questions from deck entries
type QuestionGenerator interface {
	CanGenerate(questionType QuestionType, direction Direction, deckMetadata DeckMetadata) bool
	Generate(entry DeckEntry, questionType QuestionType, direction Direction, getRandomAnswers func(count int, category string) []string) (*GeneratedQuestion, error)
}

// VocabularyQuestionGenerator generates questions for vocabulary decks
type VocabularyQuestionGenerator struct{}

// CanGenerate checks if this generator can handle the given type and direction
func (v *VocabularyQuestionGenerator) CanGenerate(questionType QuestionType, direction Direction, deckMetadata DeckMetadata) bool {
	// Check if the deck type is vocabulary
	if deckMetadata.DeckType != "vocabulary" {
		return false
	}

	// Check if the question type is supported
	supported := false
	for _, t := range deckMetadata.SupportedQuestionTypes {
		if t == string(questionType) {
			supported = true
			break
		}
	}
	if !supported {
		return false
	}

	// Check if direction is valid for the language pair
	if len(deckMetadata.LanguagePair) < 2 {
		return false
	}

	switch direction {
	case DirectionSourceToTarget, DirectionTargetToSource:
		return true
	default:
		return false
	}
}

// Generate generates a question from a deck entry
func (v *VocabularyQuestionGenerator) Generate(entry DeckEntry, questionType QuestionType, direction Direction, getRandomAnswers func(count int, category string) []string) (*GeneratedQuestion, error) {
	sourceText, ok := entry.Data["source_text"].(string)
	if !ok {
		return nil, fmt.Errorf("entry missing source_text")
	}
	targetText, ok := entry.Data["target_text"].(string)
	if !ok {
		return nil, fmt.Errorf("entry missing target_text")
	}

	var questionText, correctAnswer string

	switch direction {
	case DirectionSourceToTarget:
		questionText = sourceText
		correctAnswer = targetText
	case DirectionTargetToSource:
		questionText = targetText
		correctAnswer = sourceText
	default:
		return nil, fmt.Errorf("unsupported direction: %s", direction)
	}

	generated := &GeneratedQuestion{
		QuestionText:  questionText,
		CorrectAnswer: correctAnswer,
		QuestionType:  questionType,
		Direction:     direction,
	}

	switch questionType {
	case QuestionTypeMCQ:
		// Get 3 random wrong answers from the same category
		wrongAnswers := getRandomAnswers(3, entry.Category)
		// Filter out the correct answer
		filtered := make([]string, 0, len(wrongAnswers))
		for _, ans := range wrongAnswers {
			if ans != correctAnswer {
				filtered = append(filtered, ans)
			}
		}
		// If not enough, add some defaults (this is a simplification)
		for len(filtered) < 3 {
			filtered = append(filtered, fmt.Sprintf("dummy%d", len(filtered)))
		}
		generated.Options = append([]string{correctAnswer}, filtered[:3]...)
		// Shuffle options
		rand.Shuffle(len(generated.Options), func(i, j int) {
			generated.Options[i], generated.Options[j] = generated.Options[j], generated.Options[i]
		})

	case QuestionTypeWriteWord, QuestionTypeTranslate:
		// No options needed
		generated.Options = nil

	case QuestionTypeCustom:
		// Custom questions are handled separately
		return nil, fmt.Errorf("custom questions should not be generated from entries")

	default:
		return nil, fmt.Errorf("unsupported question type: %s", questionType)
	}

	return generated, nil
}

// QuestionGeneratorRegistry manages question generators
type QuestionGeneratorRegistry struct {
	generators []QuestionGenerator
}

// NewQuestionGeneratorRegistry creates a new registry
func NewQuestionGeneratorRegistry() *QuestionGeneratorRegistry {
	return &QuestionGeneratorRegistry{
		generators: []QuestionGenerator{
			&VocabularyQuestionGenerator{},
		},
	}
}

// GetGeneratorForDeck returns the appropriate generator for a deck
func (r *QuestionGeneratorRegistry) GetGeneratorForDeck(deckMetadata DeckMetadata) QuestionGenerator {
	for _, gen := range r.generators {
		// Check if any supported type can be generated
		for _, qt := range deckMetadata.SupportedQuestionTypes {
			if gen.CanGenerate(QuestionType(qt), DirectionSourceToTarget, deckMetadata) {
				return gen
			}
		}
	}
	return nil
}
