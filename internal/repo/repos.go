package repo

import (
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repos struct {
	User                  *UserRepo
	Token                 *TokenRepo
	Lockout               *LockoutRepo
	Deck                  *DeckRepo
	Category              *CategoryRepo
	Question              *QuestionRepo
	QuestionOption        *QuestionOptionRepo
	Quiz                  *QuizRepo
	QuizCategorySelection *QuizCategorySelectionRepo
	QuizAttempt           *QuizAttemptRepo
	UserAnswer            *UserAnswerRepo
	Progression           *ProgressionRepo
	Coin                  *CoinRepo
}

type Columns struct {
	allColumns   []string
	AllRaw       string
	AllPrefixed  string
	safeColumns  []string
	SafeRaw      string
	SafePrefixed string
}

func NewRepos(db *sqlx.DB) *Repos {
	return &Repos{
		User:                  NewUserRepo(db),
		Token:                 NewTokenRepo(db),
		Lockout:               NewLockoutRepo(db),
		Deck:                  NewDeckRepo(db),
		Category:              NewCategoryRepo(db),
		Question:              NewQuestionRepo(db),
		QuestionOption:        NewQuestionOptionRepo(db),
		Quiz:                  NewQuizRepo(db),
		QuizCategorySelection: NewQuizCategorySelectionRepo(db),
		QuizAttempt:           NewQuizAttemptRepo(db),
		UserAnswer:            NewUserAnswerRepo(db),
		Progression:           NewProgressionRepo(db),
		Coin:                  NewCoinRepo(db),
	}
}

func ExtractColumns[T any]() Columns {
	var allCols, safeCols []string

	t := reflect.TypeOf((*T)(nil)).Elem()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag, ok := field.Tag.Lookup("db")
		if !ok || dbTag == "-" {
			dbTag = strings.ToLower(field.Name)
		}
		allCols = append(allCols, dbTag)

		if safeTag, ok := field.Tag.Lookup("safe"); ok && safeTag == "true" {
			safeCols = append(safeCols, dbTag)
		}
	}

	allInsert := strings.Join(allCols, ", ")
	allSelect := ":" + strings.Join(allCols, ", :")

	safeInsert := strings.Join(safeCols, ", ")
	safeSelect := ":" + strings.Join(safeCols, ", :")

	return Columns{
		allColumns:   allCols,
		AllRaw:       allInsert,
		AllPrefixed:  allSelect,
		safeColumns:  safeCols,
		SafeRaw:      safeInsert,
		SafePrefixed: safeSelect,
	}
}
