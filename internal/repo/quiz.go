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

func (r *CategoryRepo) GetByDeckID(ctx context.Context, deckID int64) ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.SelectContext(ctx, &categories, `
		SELECT id, deck_id, category_key, title, display_order, created_at
		FROM quiz_categories
		WHERE deck_id = $1
		ORDER BY display_order
	`, deckID)
	return categories, err
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

// GetRandomAnswersFromCategory gets random answers from a specific category
func (r *QuestionRepo) GetRandomAnswersFromCategory(ctx context.Context, categoryID, excludeQuestionID int64, limit int) ([]string, error) {
	var answers []string
	err := r.db.SelectContext(ctx, &answers, `
		SELECT DISTINCT correct_answer
		FROM questions
		WHERE category_id = $1
		  AND id != $2
		  AND is_active = 1
		ORDER BY RANDOM()
		LIMIT $3
	`, categoryID, excludeQuestionID, limit)
	return answers, err
}

// GetRandomAnswersFromCategories gets random answers from multiple categories
func (r *QuestionRepo) GetRandomAnswersFromCategories(ctx context.Context, categoryIDs []int64, excludeQuestionID int64, limit int) ([]string, error) {
	if len(categoryIDs) == 0 {
		return []string{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT DISTINCT correct_answer
		FROM questions
		WHERE category_id IN (?)
		  AND id != ?
		  AND is_active = 1
		ORDER BY RANDOM()
		LIMIT ?
	`, categoryIDs, excludeQuestionID, limit)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var answers []string
	err = r.db.SelectContext(ctx, &answers, query, args...)
	return answers, err
}

// GetRandomAnswersFromDeckExcludingCategory gets random answers from deck but excludes a category
func (r *QuestionRepo) GetRandomAnswersFromDeckExcludingCategory(ctx context.Context, deckID, excludeCategoryID, excludeQuestionID int64, limit int) ([]string, error) {
	var answers []string
	err := r.db.SelectContext(ctx, &answers, `
		SELECT DISTINCT correct_answer
		FROM questions
		WHERE deck_id = $1
		  AND category_id != $2
		  AND id != $3
		  AND is_active = 1
		ORDER BY RANDOM()
		LIMIT $4
	`, deckID, excludeCategoryID, excludeQuestionID, limit)
	return answers, err
}

// GetRandomAnswers gets random answers from any question
func (r *QuestionRepo) GetRandomAnswers(ctx context.Context, excludeQuestionID int64, limit int) ([]string, error) {
	var answers []string
	err := r.db.SelectContext(ctx, &answers, `
		SELECT DISTINCT correct_answer
		FROM questions
		WHERE id != $1
		  AND is_active = 1
		ORDER BY RANDOM()
		LIMIT $2
	`, excludeQuestionID, limit)
	return answers, err
}

// CountByQuizID counts questions for a quiz
func (r *QuestionRepo) CountByQuizID(ctx context.Context, quizID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM questions q
		INNER JOIN quiz_category_selections qcs ON q.category_id = qcs.category_id
		WHERE qcs.quiz_id = $1 AND q.is_active = 1
	`, quizID)
	return count, err
}

// GetQuestionsByQuizID retrieves all questions for a quiz based on category selections
func (r *QuestionRepo) GetQuestionsByQuizID(ctx context.Context, quizID int64) ([]*model.Question, error) {
	var questions []*model.Question
	err := r.db.SelectContext(ctx, &questions, `
		SELECT q.id, q.deck_id, q.category_id, q.question_key,
		       q.question_text, q.correct_answer, q.arabic, q.french,
		       q.question_type, q.difficulty, q.points, q.hint,
		       q.explanation, q.created_at, q.updated_at, q.is_active
		FROM questions q
		INNER JOIN quiz_category_selections qcs ON q.category_id = qcs.category_id
		WHERE qcs.quiz_id = $1 AND q.is_active = 1
		ORDER BY RANDOM()
	`, quizID)
	return questions, err
}

// QuizRepo handles quiz database operations
type QuizRepo struct {
	db *sqlx.DB
}

func NewQuizRepo(db *sqlx.DB) *QuizRepo {
	return &QuizRepo{db: db}
}

// GetByID retrieves a quiz by ID
func (r *QuizRepo) GetByID(ctx context.Context, quizID int64) (*model.Quiz, error) {
	var quiz model.Quiz
	err := r.db.GetContext(ctx, &quiz, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE id = $1 AND is_active = 1
	`, quizID)
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

// GetSystemQuizzes retrieves all system quizzes ordered by level
func (r *QuizRepo) GetSystemQuizzes(ctx context.Context) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE is_system = 1 AND is_active = 1
		ORDER BY level_order
	`)
	return quizzes, err
}

// GetNextUnlockedQuiz finds the next quiz to unlock for a user
func (r *QuizRepo) GetNextUnlockedQuiz(ctx context.Context, userID int64, currentLevel int) (*model.Quiz, error) {
	var quiz model.Quiz
	err := r.db.GetContext(ctx, &quiz, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE is_system = 1 AND is_active = 1
		  AND level_order >= $1
		ORDER BY level_order
		LIMIT 1
	`, currentLevel)
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

// GetPublicQuizzes retrieves public user-created quizzes
func (r *QuizRepo) GetPublicQuizzes(ctx context.Context, limit, offset int) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE is_public = 1 AND is_active = 1
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return quizzes, err
}

// GetQuizzesByCreator retrieves quizzes created by a specific user
func (r *QuizRepo) GetQuizzesByCreator(ctx context.Context, userID int64) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE created_by = $1 AND is_active = 1
		ORDER BY created_at DESC
	`, userID)
	return quizzes, err
}

// Create creates a new quiz
func (r *QuizRepo) Create(ctx context.Context, quiz *model.Quiz) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quizzes (
			id, title, description, deck_id, time_limit, pass_percentage,
			shuffle_questions, question_mode, gives_coins, coin_reward,
			level_order, prerequisite_quiz_id, is_public, is_system,
			created_by, created_at, is_active
		) VALUES (
			:id, :title, :description, :deck_id, :time_limit, :pass_percentage,
			:shuffle_questions, :question_mode, :gives_coins, :coin_reward,
			:level_order, :prerequisite_quiz_id, :is_public, :is_system,
			:created_by, :created_at, :is_active
		)
	`, quiz)
	return err
}

// GetAllQuizzes retrieves all quizzes with pagination
func (r *QuizRepo) GetAllQuizzes(ctx context.Context, limit, offset int) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE is_active = 1
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return quizzes, err
}

// GetUserGeneratedQuizzes retrieves quizzes created by users (not system)
func (r *QuizRepo) GetUserGeneratedQuizzes(ctx context.Context, limit, offset int) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, deck_id, time_limit, pass_percentage,
		       shuffle_questions, question_mode, gives_coins, coin_reward,
		       level_order, prerequisite_quiz_id, is_public, is_system,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE is_system = 0 AND is_active = 1 AND created_by IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return quizzes, err
}

// DeactivateQuiz soft deletes a quiz
func (r *QuizRepo) DeactivateQuiz(ctx context.Context, quizID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quizzes SET is_active = 0 WHERE id = $1
	`, quizID)
	return err
}

// CountTotalQuizzes counts all active quizzes
func (r *QuizRepo) CountTotalQuizzes(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quizzes WHERE is_active = 1
	`)
	return count, err
}

// QuestionOptionRepo handles question option database operations
type QuestionOptionRepo struct {
	db *sqlx.DB
}

func NewQuestionOptionRepo(db *sqlx.DB) *QuestionOptionRepo {
	return &QuestionOptionRepo{db: db}
}

func (r *QuestionOptionRepo) Create(ctx context.Context, option *model.AnswerOption) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO question_options (
			id, question_id, option_text, is_correct,
			is_auto_generated, generation_rule, display_order, created_at
		) VALUES (
			:id, :question_id, :option_text, :is_correct,
			:is_auto_generated, :generation_rule, :display_order, :created_at
		)
	`, option)
	return err
}

func (r *QuestionOptionRepo) GetByQuestionID(ctx context.Context, questionID int64) ([]*model.AnswerOption, error) {
	var options []*model.AnswerOption
	err := r.db.SelectContext(ctx, &options, `
		SELECT id, question_id, option_text, is_correct,
		       is_auto_generated, generation_rule, display_order, created_at
		FROM question_options
		WHERE question_id = $1
		ORDER BY display_order
	`, questionID)
	return options, err
}

func (r *QuestionOptionRepo) DeleteByQuestionID(ctx context.Context, questionID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM question_options WHERE question_id = $1
	`, questionID)
	return err
}

// CoinRepo handles user coin operations
type CoinRepo struct {
	db *sqlx.DB
}

func NewCoinRepo(db *sqlx.DB) *CoinRepo {
	return &CoinRepo{db: db}
}

func (r *CoinRepo) GetUserCoins(ctx context.Context, userID int64) (*model.UserCoins, error) {
	var coins model.UserCoins
	err := r.db.GetContext(ctx, &coins, `
		SELECT user_id, balance, lifetime_earned, last_updated
		FROM user_coins
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	return &coins, nil
}

func (r *CoinRepo) CreateOrUpdateCoins(ctx context.Context, coins *model.UserCoins) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO user_coins (user_id, balance, lifetime_earned, last_updated)
		VALUES (:user_id, :balance, :lifetime_earned, :last_updated)
		ON CONFLICT (user_id) DO UPDATE SET
			balance = EXCLUDED.balance,
			lifetime_earned = EXCLUDED.lifetime_earned,
			last_updated = EXCLUDED.last_updated
	`, coins)
	return err
}

func (r *CoinRepo) AddTransaction(ctx context.Context, tx *model.CoinTransaction) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO coin_transactions (
			id, user_id, amount, transaction_type,
			reference_type, reference_id, description, created_at
		) VALUES (
			:id, :user_id, :amount, :transaction_type,
			:reference_type, :reference_id, :description, :created_at
		)
	`, tx)
	return err
}

// QuizCategorySelectionRepo handles quiz category selection operations
type QuizCategorySelectionRepo struct {
	db *sqlx.DB
}

func NewQuizCategorySelectionRepo(db *sqlx.DB) *QuizCategorySelectionRepo {
	return &QuizCategorySelectionRepo{db: db}
}

func (r *QuizCategorySelectionRepo) Create(ctx context.Context, selection *model.QuizCategorySelection) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_category_selections (quiz_id, category_id, question_count)
		VALUES (:quiz_id, :category_id, :question_count)
	`, selection)
	return err
}

func (r *QuizCategorySelectionRepo) GetByQuizID(ctx context.Context, quizID int64) ([]*model.QuizCategorySelection, error) {
	var selections []*model.QuizCategorySelection
	err := r.db.SelectContext(ctx, &selections, `
		SELECT quiz_id, category_id, question_count
		FROM quiz_category_selections
		WHERE quiz_id = $1
	`, quizID)
	return selections, err
}
