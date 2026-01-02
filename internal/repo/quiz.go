package repo

import (
	"context"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type DeckRepo struct {
	db *sqlx.DB
}

func NewDeckRepo(db *sqlx.DB) *DeckRepo {
	return &DeckRepo{db: db}
}

func (r *DeckRepo) GetByKey(ctx context.Context, deckKey string) (*model.Deck, error) {
	var deck model.Deck
	err := r.db.GetContext(ctx, &deck, `
		SELECT id, deck_key, title, version, source_file, is_system, created_at
		FROM quiz_decks
		WHERE deck_key = $1
	`, deckKey)
	if err != nil {
		return nil, err
	}
	return &deck, nil
}

func (r *DeckRepo) Create(ctx context.Context, deck *model.Deck) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_decks (id, deck_key, title, version, source_file, is_system, created_at)
		VALUES (:id, :deck_key, :title, :version, :source_file, :is_system, :created_at)
	`, deck)
	return err
}

type CategoryRepo struct {
	db *sqlx.DB
}

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, category *model.Category) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_categories (id, deck_id, category_key, title, display_order, created_at)
		VALUES (:id, :deck_id, :category_key, :title, :display_order, :created_at)
	`, category)
	return err
}

type QuestionRepo struct {
	db *sqlx.DB
}

func NewQuestionRepo(db *sqlx.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) Create(ctx context.Context, question *model.Question) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO questions (
			id, deck_id, category_id, question_key,
			question_text, correct_answer, arabic, french,
			question_type, points, created_at, is_active
		) VALUES (
			:id, :deck_id, :category_id, :question_key,
			:question_text, :correct_answer, :arabic, :french,
			:question_type, :points, :created_at, :is_active
		)
	`, question)
	return err
}

func (r *QuestionRepo) GetRandomAnswersFromDeck(ctx context.Context, deckID, excludeQuestionID int64, limit int) ([]string, error) {
	var answers []string
	err := r.db.SelectContext(ctx, &answers, `
		SELECT DISTINCT correct_answer
		FROM questions
		WHERE deck_id = $1
		  AND id != $2
		  AND is_active = 1
		ORDER BY RANDOM()
		LIMIT $3
	`, deckID, excludeQuestionID, limit)
	return answers, err
}
