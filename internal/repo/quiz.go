package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type DeckRepo struct {
	Columns
	db *sqlx.DB
}

func NewDeckRepo(db *sqlx.DB) *DeckRepo {
	repo := &DeckRepo{db: db}
	repo.Columns = ExtractColumns[model.Deck]()
	return repo
}

func (r *DeckRepo) GetByKey(ctx context.Context, deckKey string) (*model.Deck, error) {
	var deck model.Deck
	query := fmt.Sprintf("SELECT %s FROM quiz_decks WHERE deck_key = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &deck, query, deckKey)
	if err != nil {
		return nil, err
	}
	return &deck, nil
}

func (r *DeckRepo) Create(ctx context.Context, deck *model.Deck) error {
	query := fmt.Sprintf(
		"INSERT INTO quiz_decks (%s) VALUES (%s)",
		r.AllRaw,
		r.AllPrefixed,
	)
	_, err := r.db.NamedExecContext(ctx, query, deck)
	return err
}

func (r *DeckRepo) GetAll(ctx context.Context) ([]*model.Deck, error) {
	var decks []*model.Deck
	query := fmt.Sprintf("SELECT %s FROM quiz_decks ORDER BY title", r.AllRaw)
	err := r.db.SelectContext(ctx, &decks, query)
	return decks, err
}

func (r *DeckRepo) GetByKeyAndVersion(ctx context.Context, deckKey string, version int) (*model.Deck, error) {
	var deck model.Deck
	query := fmt.Sprintf("SELECT %s FROM quiz_decks WHERE deck_key = $1 AND version = $2", r.AllRaw)
	err := r.db.GetContext(ctx, &deck, query, deckKey, version)
	if err != nil {
		return nil, err
	}
	return &deck, nil
}

func (r *DeckRepo) GetByID(ctx context.Context, id int64) (*model.Deck, error) {
	var deck model.Deck
	query := fmt.Sprintf("SELECT %s FROM quiz_decks WHERE id = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &deck, query, id)
	if err != nil {
		return nil, err
	}
	return &deck, nil
}

type CategoryRepo struct {
	Columns
	db *sqlx.DB
}

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo {
	repo := &CategoryRepo{db: db}
	repo.Columns = ExtractColumns[model.Category]()
	return repo
}

func (r *CategoryRepo) Create(ctx context.Context, category *model.Category) error {
	query := fmt.Sprintf(
		"INSERT INTO quiz_categories (%s) VALUES (%s)",
		r.AllRaw,
		r.AllPrefixed,
	)
	_, err := r.db.NamedExecContext(ctx, query, category)
	return err
}

func (r *CategoryRepo) GetByDeckID(ctx context.Context, deckID int64) ([]*model.Category, error) {
	var categories []*model.Category
	query := fmt.Sprintf("SELECT %s FROM quiz_categories WHERE deck_id = $1 ORDER BY display_order", r.AllRaw)
	err := r.db.SelectContext(ctx, &categories, query, deckID)
	return categories, err
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*model.Category, error) {
	var category model.Category
	query := fmt.Sprintf("SELECT %s FROM quiz_categories WHERE id = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepo) CountByDeckID(ctx context.Context, deckID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quiz_categories WHERE deck_id = $1
	`, deckID)
	return count, err
}

type DeckEntryRepo struct {
	Columns
	db *sqlx.DB
}

func NewDeckEntryRepo(db *sqlx.DB) *DeckEntryRepo {
	repo := &DeckEntryRepo{db: db}
	repo.Columns = ExtractColumns[model.DeckEntry]()
	return repo
}

func (r *DeckEntryRepo) Create(ctx context.Context, entry *model.DeckEntry) error {
	query := fmt.Sprintf(
		"INSERT INTO deck_entries (%s) VALUES (%s)",
		r.AllRaw,
		r.AllPrefixed,
	)
	_, err := r.db.NamedExecContext(ctx, query, entry)
	return err
}

func (r *DeckEntryRepo) GetByDeckID(ctx context.Context, deckID int64) ([]*model.DeckEntry, error) {
	var entries []*model.DeckEntry
	query := fmt.Sprintf("SELECT %s FROM deck_entries WHERE deck_id = $1 ORDER BY created_at", r.AllRaw)
	err := r.db.SelectContext(ctx, &entries, query, deckID)
	return entries, err
}

func (r *DeckEntryRepo) GetByDeckAndCategoryID(ctx context.Context, deckID, categoryID int64) ([]*model.DeckEntry, error) {
	var entries []*model.DeckEntry
	query := fmt.Sprintf("SELECT %s FROM deck_entries WHERE deck_id = $1 AND category_id = $2 ORDER BY created_at", r.AllRaw)
	err := r.db.SelectContext(ctx, &entries, query, deckID, categoryID)
	return entries, err
}

type DeckCacheRepo struct {
	Columns
	db *sqlx.DB
}

func NewDeckCacheRepo(db *sqlx.DB) *DeckCacheRepo {
	repo := &DeckCacheRepo{db: db}
	repo.Columns = ExtractColumns[model.DeckCache]()
	return repo
}

func (r *DeckCacheRepo) Upsert(ctx context.Context, cache *model.DeckCache) error {
	query := fmt.Sprintf(`
		INSERT INTO deck_cache (%s) VALUES (%s)
		ON CONFLICT (deck_id) DO UPDATE SET
			cached_data = EXCLUDED.cached_data,
			cache_version = EXCLUDED.cache_version,
			last_updated = EXCLUDED.last_updated
	`, r.AllRaw, r.AllPrefixed)
	_, err := r.db.NamedExecContext(ctx, query, cache)
	return err
}

func (r *DeckCacheRepo) GetByDeckID(ctx context.Context, deckID int64) (*model.DeckCache, error) {
	var cache model.DeckCache
	query := fmt.Sprintf("SELECT %s FROM deck_cache WHERE deck_id = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &cache, query, deckID)
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

func (r *DeckCacheRepo) DeleteByDeckID(ctx context.Context, deckID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM deck_cache WHERE deck_id = $1", deckID)
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

func (r *QuestionRepo) CountByDeckID(ctx context.Context, deckID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM questions WHERE deck_id = $1 AND is_active = 1
	`, deckID)
	return count, err
}

func (r *QuestionRepo) CountByCategoryID(ctx context.Context, categoryID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM questions WHERE category_id = $1 AND is_active = 1
	`, categoryID)
	return count, err
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

// BeginTx begins a transaction
func (r *QuizRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
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

// CreateWithTx creates a new quiz within a transaction
func (r *QuizRepo) CreateWithTx(ctx context.Context, tx *sqlx.Tx, quiz *model.Quiz) error {
	_, err := tx.NamedExecContext(ctx, `
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
		UPDATE quizzes SET is_active = 0, updated_at = $1 WHERE id = $2
	`, time.Now().Unix(), quizID)
	return err
}

// UpdateQuiz updates a quiz's metadata and increments version
func (r *QuizRepo) UpdateQuiz(ctx context.Context, quiz *model.Quiz) error {
	quiz.UpdatedAt = new(int64)
	*quiz.UpdatedAt = time.Now().Unix()
	quiz.Version++ // Increment version on each update

	_, err := r.db.ExecContext(ctx, `
		UPDATE quizzes SET
			title = $1,
			description = $2,
			version = $3,
			pass_percentage = $4,
			shuffle_questions = $5,
			gives_coins = $6,
			coin_reward = $7,
			level_order = $8,
			prerequisite_quiz_id = $9,
			is_public = $10,
			updated_at = $11
		WHERE id = $12
	`, quiz.Title, quiz.Description, quiz.Version, quiz.PassPercentage, quiz.ShuffleQuestions,
		quiz.GivesCoins, quiz.CoinReward, quiz.LevelOrder, quiz.PrerequisiteQuizID,
		quiz.IsPublic, quiz.UpdatedAt, quiz.ID)
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

func (r *QuizCategorySelectionRepo) CreateWithTx(ctx context.Context, tx *sqlx.Tx, selection *model.QuizCategorySelection) error {
	_, err := tx.NamedExecContext(ctx, `
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

type QuizQuestionRepo struct {
	db *sqlx.DB
}

func NewQuizQuestionRepo(db *sqlx.DB) *QuizQuestionRepo {
	return &QuizQuestionRepo{db: db}
}

func (r *QuizQuestionRepo) Create(ctx context.Context, question *model.QuizQuestion) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_questions (
			id, quiz_id, question_id, question_text, correct_answer,
			options, question_type, direction, display_order, created_at
		) VALUES (
			:id, :quiz_id, :question_id, :question_text, :correct_answer,
			:options, :question_type, :direction, :display_order, :created_at
		)
	`, question)
	return err
}

func (r *QuizQuestionRepo) CreateWithTx(ctx context.Context, tx *sqlx.Tx, question *model.QuizQuestion) error {
	_, err := tx.NamedExecContext(ctx, `
		INSERT INTO quiz_questions (
			id, quiz_id, question_id, question_text, correct_answer,
			options, question_type, direction, display_order, created_at
		) VALUES (
			:id, :quiz_id, :question_id, :question_text, :correct_answer,
			:options, :question_type, :direction, :display_order, :created_at
		)
	`, question)
	return err
}

func (r *QuizQuestionRepo) GetByQuizID(ctx context.Context, quizID int64) ([]*model.QuizQuestion, error) {
	var questions []*model.QuizQuestion
	err := r.db.SelectContext(ctx, &questions, `
		SELECT id, quiz_id, question_id, question_text, correct_answer,
		       options, question_type, direction, display_order, created_at
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY display_order
	`, quizID)
	return questions, err
}

func (r *QuizQuestionRepo) GetByID(ctx context.Context, questionID int64) (*model.QuizQuestion, error) {
	var question model.QuizQuestion
	err := r.db.GetContext(ctx, &question, `
		SELECT id, quiz_id, question_id, question_text, correct_answer,
		       options, question_type, direction, display_order, created_at
		FROM quiz_questions
		WHERE id = $1
	`, questionID)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuizQuestionRepo) CountByQuizID(ctx context.Context, quizID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quiz_questions WHERE quiz_id = $1
	`, quizID)
	return count, err
}

type QuizSessionRepo struct {
	db *sqlx.DB
}

func NewQuizSessionRepo(db *sqlx.DB) *QuizSessionRepo {
	return &QuizSessionRepo{db: db}
}

func (r *QuizSessionRepo) Create(ctx context.Context, session *model.QuizSession) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_sessions (
			id, quiz_id, user_id, current_question_index, started_at, last_activity, is_completed
		) VALUES (
			:id, :quiz_id, :user_id, :current_question_index, :started_at, :last_activity, :is_completed
		)
	`, session)
	return err
}

func (r *QuizSessionRepo) GetByID(ctx context.Context, sessionID string) (*model.QuizSession, error) {
	var session model.QuizSession
	err := r.db.GetContext(ctx, &session, `
		SELECT id, quiz_id, user_id, current_question_index, started_at, last_activity, is_completed
		FROM quiz_sessions
		WHERE id = $1
	`, sessionID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *QuizSessionRepo) UpdateActivity(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quiz_sessions SET last_activity = $1 WHERE id = $2
	`, time.Now().Unix(), sessionID)
	return err
}

func (r *QuizSessionRepo) UpdateProgress(ctx context.Context, sessionID string, currentIndex int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quiz_sessions SET current_question_index = $1, last_activity = $2 WHERE id = $3
	`, currentIndex, time.Now().Unix(), sessionID)
	return err
}

func (r *QuizSessionRepo) Complete(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quiz_sessions SET is_completed = 1, last_activity = $1 WHERE id = $2
	`, time.Now().Unix(), sessionID)
	return err
}
