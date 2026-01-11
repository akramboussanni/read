package quiz

import "time"

// DeckMetadata represents the metadata of a quiz deck
type DeckMetadata struct {
	Title                  string   `json:"title"`
	Version                int      `json:"version"`
	DeckType               string   `json:"deck_type"`
	LanguagePair           []string `json:"language_pair"`
	Description            string   `json:"description"`
	SupportedQuestionTypes []string `json:"supported_question_types"`
	// Removed: DefaultDirection - now handled per question
}

// Category represents a category in the deck
type Category struct {
	Title      string                 `json:"title"`
	Difficulty string                 `json:"difficulty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// DeckEntry represents an entry in the deck
type DeckEntry struct {
	ID       string                 `json:"id"`
	Category string                 `json:"category"`
	Data     map[string]interface{} `json:"data"`
	Tags     []string               `json:"tags,omitempty"`
	// Data can include:
	// "source_text": "كَتَبَ"
	// "target_text": "écrire"
	// "preferred_directions": ["source_to_target", "target_to_source"]
	// "direction_difficulty": {"source_to_target": "easy", "target_to_source": "hard"}
}

// UniversalDeck represents the full universal JSON structure
type UniversalDeck struct {
	Metadata   DeckMetadata        `json:"metadata"`
	Categories map[string]Category `json:"categories"`
	Entries    []DeckEntry         `json:"entries"`
}

// ParsedDeck is the internal representation after parsing
type ParsedDeck struct {
	DeckID    int64
	DeckKey   string
	Title     string
	Version   int
	Categories map[string]string // category key -> title
	Questions  []ParsedQuestion
	Metadata   DeckMetadata
}

// ParsedQuestion represents a parsed question
type ParsedQuestion struct {
	ID             string
	CategoryKey    string
	QuestionText   string
	CorrectAnswer  string
	AdditionalData map[string]interface{}
}

// QuestionType represents the type of question
type QuestionType string

const (
	QuestionTypeMCQ       QuestionType = "mcq"
	QuestionTypeWriteWord QuestionType = "write_word"
	QuestionTypeTranslate QuestionType = "translate"
	QuestionTypeCustom    QuestionType = "custom"
)

// Direction represents the translation direction
type Direction string

const (
	DirectionSourceToTarget Direction = "source_to_target"
	DirectionTargetToSource Direction = "target_to_source"
)

// GeneratedQuestion represents a generated question
type GeneratedQuestion struct {
	QuestionText  string
	CorrectAnswer string
	Options       []string // For MCQ
	QuestionType  QuestionType
	Direction     Direction
}

// CustomQuestion represents a user-defined custom question
type CustomQuestion struct {
	QuestionText  string
	CorrectAnswer string
	WrongAnswers  []string // For MCQ
	QuestionType  QuestionType
}

// QuizConfig represents the configuration for a quiz
type QuizConfig struct {
	DeckSelections   []DeckSelection       // Multiple deck-category selections
	QuestionTypes    []QuestionType
	Directions       []Direction
	QuestionCount    int
	Difficulty       string
	CustomQuestions  []CustomQuestion
}

// DeckSelection represents a selection of categories from a specific deck
type DeckSelection struct {
	DeckID     int64
	Categories []string // category keys to include, empty means all
}

// Quiz represents a quiz session
type Quiz struct {
	ID           int64
	Config       QuizConfig
	Questions    []QuizQuestion
	CurrentIndex int
	StartedAt    time.Time
	CompletedAt  *time.Time
}

// QuizQuestion represents a question in a quiz
type QuizQuestion struct {
	ID            int64
	QuestionText  string
	CorrectAnswer string
	Options       []string
	QuestionType  QuestionType
	Direction     Direction
	UserAnswer    *string
	IsCorrect     *bool
	TimeSpent     *time.Duration
}
