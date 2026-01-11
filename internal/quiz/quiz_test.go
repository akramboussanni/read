package quiz

import (
	"context"
	"testing"
)

func TestUniversalQuizSystem(t *testing.T) {
	service := NewQuizService(nil)

	// Load the vocabulary deck
	deck, err := service.LoadDeck("data/vocabulary.json")
	if err != nil {
		t.Fatalf("Failed to load deck: %v", err)
	}

	// Set mock deck ID for testing
	deck.DeckID = 1

	if deck.Title != "80% des Mots du Qour'ân - Part 2" {
		t.Errorf("Expected title '80%% des Mots du Qour'ân - Part 2', got %s", deck.Title)
	}

	if len(deck.Questions) == 0 {
		t.Error("Expected questions to be loaded")
	}

	// Create a quiz config
	config := QuizConfig{
		DeckSelections: []DeckSelection{
			{
				DeckID:     1, // Assume deck ID 1 for the test
				Categories: []string{"pronouns_suffixes"},
			},
		},
		QuestionTypes: []QuestionType{QuestionTypeMCQ, QuestionTypeWriteWord},
		Directions:    []Direction{DirectionSourceToTarget},
		QuestionCount: 5,
	}

	// Generate quiz
	quiz, err := service.CreateQuiz(context.Background(), config, []*ParsedDeck{deck})
	if err != nil {
		t.Fatalf("Failed to create quiz: %v", err)
	}

	if len(quiz.Questions) != 5 {
		t.Errorf("Expected 5 questions, got %d", len(quiz.Questions))
	}

	// Test answer validation
	question := quiz.Questions[0]
	correct := service.ValidateAnswer(question, question.CorrectAnswer)
	if !correct {
		t.Error("Expected correct answer to be validated as correct")
	}

	incorrect := service.ValidateAnswer(question, "wrong answer")
	if incorrect {
		t.Error("Expected incorrect answer to be validated as incorrect")
	}
}
