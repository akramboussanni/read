package quiz

import (
	"fmt"
	"math/rand"
	"strings"
)

// UniversalQuestionGenerator handles question generation for any deck type
type UniversalQuestionGenerator struct{}

// CanGenerate checks if this generator can handle the given question type and direction
func (v *VocabularyQuestionGenerator) CanGenerate(questionType QuestionType, direction Direction, metadata DeckMetadata) bool {
	switch questionType {
	case QuestionTypeMCQ, QuestionTypeWriteWord, QuestionTypeTranslate, QuestionTypeCustom:
		return true
	default:
		return false
	}
}

// Generate creates a question from a deck entry
func (v *VocabularyQuestionGenerator) Generate(entry DeckEntry, questionType QuestionType, direction Direction) (*GeneratedQuestion, error) {
	switch questionType {
	case QuestionTypeMCQ:
		return v.generateMCQ(entry, direction)
	case QuestionTypeWriteWord:
		return v.generateWriteWord(entry, direction)
	case QuestionTypeTranslate:
		return v.generateTranslate(entry, direction)
	default:
		return nil, fmt.Errorf("unsupported question type: %s", questionType)
	}
}

func (v *VocabularyQuestionGenerator) generateMCQ(entry DeckEntry, direction Direction) (*GeneratedQuestion, error) {
	var questionText, correctAnswer string

	switch direction {
	case DirectionSourceToTarget:
		if source, ok := entry.Data["arabic"].(string); ok {
			questionText = fmt.Sprintf("What is the French translation of: %s", source)
			if target, ok := entry.Data["french"].(string); ok {
				correctAnswer = target
			}
		}
	case DirectionTargetToSource:
		if target, ok := entry.Data["french"].(string); ok {
			questionText = fmt.Sprintf("What is the Arabic translation of: %s", target)
			if source, ok := entry.Data["arabic"].(string); ok {
				correctAnswer = source
			}
		}
	}

	if questionText == "" || correctAnswer == "" {
		return nil, fmt.Errorf("insufficient data for MCQ generation")
	}

	return &GeneratedQuestion{
		QuestionText:  questionText,
		CorrectAnswer: correctAnswer,
		Options:       []string{}, // Will be populated with wrong answers from other entries
		QuestionType:  QuestionTypeMCQ,
		Direction:     direction,
	}, nil
}

func (v *VocabularyQuestionGenerator) generateWriteWord(entry DeckEntry, direction Direction) (*GeneratedQuestion, error) {
	var questionText, correctAnswer string

	switch direction {
	case DirectionSourceToTarget:
		if source, ok := entry.Data["arabic"].(string); ok {
			questionText = fmt.Sprintf("Translate to French: %s", source)
			if target, ok := entry.Data["french"].(string); ok {
				correctAnswer = target
			}
		}
	case DirectionTargetToSource:
		if target, ok := entry.Data["french"].(string); ok {
			questionText = fmt.Sprintf("Translate to Arabic: %s", target)
			if source, ok := entry.Data["arabic"].(string); ok {
				correctAnswer = source
			}
		}
	}

	if questionText == "" || correctAnswer == "" {
		return nil, fmt.Errorf("insufficient data for write word generation")
	}

	return &GeneratedQuestion{
		QuestionText:  questionText,
		CorrectAnswer: correctAnswer,
		Options:       []string{},
		QuestionType:  QuestionTypeWriteWord,
		Direction:     direction,
	}, nil
}

func (v *VocabularyQuestionGenerator) generateTranslate(entry DeckEntry, direction Direction) (*GeneratedQuestion, error) {
	// For translate type, we expect the user to provide the full translation
	var questionText, correctAnswer string

	switch direction {
	case DirectionSourceToTarget:
		if source, ok := entry.Data["arabic"].(string); ok {
			questionText = fmt.Sprintf("Translate this Arabic text to French: %s", source)
			if target, ok := entry.Data["french"].(string); ok {
				correctAnswer = target
			}
		}
	case DirectionTargetToSource:
		if target, ok := entry.Data["french"].(string); ok {
			questionText = fmt.Sprintf("Translate this French text to Arabic: %s", target)
			if source, ok := entry.Data["arabic"].(string); ok {
				correctAnswer = source
			}
		}
	}

	if questionText == "" || correctAnswer == "" {
		return nil, fmt.Errorf("insufficient data for translate generation")
	}

	return &GeneratedQuestion{
		QuestionText:  questionText,
		CorrectAnswer: correctAnswer,
		Options:       []string{},
		QuestionType:  QuestionTypeTranslate,
		Direction:     direction,
	}, nil
}

// PopulateMCQOptions fills in wrong answers for MCQ questions from other entries in the same categories
func (v *VocabularyQuestionGenerator) PopulateMCQOptions(question *GeneratedQuestion, allEntries []DeckEntry, category string, count int) {
	if question.QuestionType != QuestionTypeMCQ {
		return
	}

	var wrongAnswers []string

	// Collect answers from entries in the same category
	for _, entry := range allEntries {
		if entry.Category == category && len(entry.Data) > 0 {
			var answer string
			switch question.Direction {
			case DirectionSourceToTarget:
				if target, ok := entry.Data["french"].(string); ok {
					answer = target
				}
			case DirectionTargetToSource:
				if source, ok := entry.Data["arabic"].(string); ok {
					answer = source
				}
			}

			// Don't include the correct answer
			if answer != "" && answer != question.CorrectAnswer && !contains(wrongAnswers, answer) {
				wrongAnswers = append(wrongAnswers, answer)
			}
		}
	}

	// Shuffle and take the required number
	rand.Shuffle(len(wrongAnswers), func(i, j int) {
		wrongAnswers[i], wrongAnswers[j] = wrongAnswers[j], wrongAnswers[i]
	})

	if len(wrongAnswers) > count {
		wrongAnswers = wrongAnswers[:count]
	}

	question.Options = wrongAnswers
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}