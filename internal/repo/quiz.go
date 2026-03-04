package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

// ============================================================
// DeckRepo
// ============================================================

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

// ============================================================
// CategoryRepo
// ============================================================

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

func (r *CategoryRepo) GetByKeyAndDeckID(ctx context.Context, deckID int64, key string) (*model.Category, error) {
	var category model.Category
	query := fmt.Sprintf("SELECT %s FROM quiz_categories WHERE deck_id = $1 AND category_key = $2", r.AllRaw)
	err := r.db.GetContext(ctx, &category, query, deckID, key)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepo) CountByDeckID(ctx context.Context, deckID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM quiz_categories WHERE deck_id = $1`, deckID)
	return count, err
}

// ============================================================
// DeckEntryRepo
// ============================================================

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

func (r *DeckEntryRepo) GetByCategoryID(ctx context.Context, categoryID int64) ([]*model.DeckEntry, error) {
	var entries []*model.DeckEntry
	query := fmt.Sprintf("SELECT %s FROM deck_entries WHERE category_id = $1 ORDER BY created_at", r.AllRaw)
	err := r.db.SelectContext(ctx, &entries, query, categoryID)
	return entries, err
}

func (r *DeckEntryRepo) GetRandomByCategoryID(ctx context.Context, categoryID int64, limit int) ([]*model.DeckEntry, error) {
	var entries []*model.DeckEntry
	query := fmt.Sprintf("SELECT %s FROM deck_entries WHERE category_id = $1 ORDER BY RANDOM() LIMIT $2", r.AllRaw)
	err := r.db.SelectContext(ctx, &entries, query, categoryID, limit)
	return entries, err
}

func (r *DeckEntryRepo) GetRandomAnswersFromCategory(ctx context.Context, categoryID int64, limit int) ([]string, error) {
	var answers []string
	query := `
		SELECT entry_data->>'target_text' as answer
		FROM deck_entries
		WHERE category_id = $1
		GROUP BY answer
		ORDER BY RANDOM()
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &answers, query, categoryID, limit)
	return answers, err
}

func (r *DeckEntryRepo) GetRandomAnswersFromDeck(ctx context.Context, deckID int64, limit int) ([]string, error) {
	var answers []string
	query := `
		SELECT entry_data->>'target_text' as answer
		FROM deck_entries
		WHERE deck_id = $1
		GROUP BY answer
		ORDER BY RANDOM()
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &answers, query, deckID, limit)
	return answers, err
}

func (r *DeckEntryRepo) CountByCategoryID(ctx context.Context, categoryID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM deck_entries WHERE category_id = $1`, categoryID)
	return count, err
}

// ============================================================
// DeckCacheRepo
// ============================================================

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

// ============================================================
// QuizRepo
// ============================================================

type QuizRepo struct {
	db *sqlx.DB
}

func NewQuizRepo(db *sqlx.DB) *QuizRepo {
	return &QuizRepo{db: db}
}

func (r *QuizRepo) GetDB() *sqlx.DB {
	return r.db
}

func (r *QuizRepo) Create(ctx context.Context, quiz *model.Quiz) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quizzes (
			id, title, description, course_id, node_id, deck_id,
			pass_percentage, shuffle_questions, question_mode,
			gives_coins, coin_reward, is_public, is_system, is_dynamic,
			created_by, created_at, is_active
		) VALUES (
			:id, :title, :description, :course_id, :node_id, :deck_id,
			:pass_percentage, :shuffle_questions, :question_mode,
			:gives_coins, :coin_reward, :is_public, :is_system, :is_dynamic,
			:created_by, :created_at, :is_active
		)
	`, quiz)
	return err
}

func (r *QuizRepo) GetByID(ctx context.Context, quizID int64) (*model.Quiz, error) {
	var quiz model.Quiz
	err := r.db.GetContext(ctx, &quiz, `
		SELECT id, title, description, course_id, node_id, deck_id,
		       pass_percentage, shuffle_questions, question_mode,
		       gives_coins, coin_reward, is_public, is_system, is_dynamic,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE id = $1 AND is_active = TRUE
	`, quizID)
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

func (r *QuizRepo) GetByCourseID(ctx context.Context, courseID string) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, course_id, node_id, deck_id,
		       pass_percentage, shuffle_questions, question_mode,
		       gives_coins, coin_reward, is_public, is_system, is_dynamic,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE course_id = $1 AND is_active = TRUE
		ORDER BY created_at
	`, courseID)
	return quizzes, err
}

func (r *QuizRepo) GetByNodeID(ctx context.Context, nodeID string) (*model.Quiz, error) {
	var quiz model.Quiz
	err := r.db.GetContext(ctx, &quiz, `
		SELECT id, title, description, course_id, node_id, deck_id,
		       pass_percentage, shuffle_questions, question_mode,
		       gives_coins, coin_reward, is_public, is_system, is_dynamic,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE node_id = $1 AND is_active = TRUE
		LIMIT 1
	`, nodeID)
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

func (r *QuizRepo) GetPublicQuizzes(ctx context.Context, limit, offset int) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, course_id, node_id, deck_id,
		       pass_percentage, shuffle_questions, question_mode,
		       gives_coins, coin_reward, is_public, is_system, is_dynamic,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE is_public = TRUE AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return quizzes, err
}

func (r *QuizRepo) GetQuizzesByCreator(ctx context.Context, userID int64) ([]*model.Quiz, error) {
	var quizzes []*model.Quiz
	err := r.db.SelectContext(ctx, &quizzes, `
		SELECT id, title, description, course_id, node_id, deck_id,
		       pass_percentage, shuffle_questions, question_mode,
		       gives_coins, coin_reward, is_public, is_system, is_dynamic,
		       created_by, created_at, updated_at, is_active
		FROM quizzes
		WHERE created_by = $1 AND is_active = TRUE
		ORDER BY created_at DESC
	`, userID)
	return quizzes, err
}

func (r *QuizRepo) DeactivateQuiz(ctx context.Context, quizID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quizzes SET is_active = FALSE, updated_at = $1 WHERE id = $2
	`, time.Now().Unix(), quizID)
	return err
}

func (r *QuizRepo) UpdateQuiz(ctx context.Context, quiz *model.Quiz) error {
	now := time.Now().Unix()
	quiz.UpdatedAt = &now
	_, err := r.db.ExecContext(ctx, `
		UPDATE quizzes SET
			title = $1, description = $2, pass_percentage = $3,
			shuffle_questions = $4, gives_coins = $5, coin_reward = $6,
			is_public = $7, updated_at = $8
		WHERE id = $9
	`, quiz.Title, quiz.Description, quiz.PassPercentage, quiz.ShuffleQuestions,
		quiz.GivesCoins, quiz.CoinReward, quiz.IsPublic, quiz.UpdatedAt, quiz.ID)
	return err
}

// ============================================================
// QuestionTemplateRepo
// ============================================================

type QuestionTemplateRepo struct {
	db *sqlx.DB
}

func NewQuestionTemplateRepo(db *sqlx.DB) *QuestionTemplateRepo {
	return &QuestionTemplateRepo{db: db}
}

func (r *QuestionTemplateRepo) Create(ctx context.Context, t *model.QuestionTemplate) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO question_templates (
			id, quiz_id, deck_id, category_id,
			question_types, directions, generation_mode,
			llm_prompt, manual_data, question_count, created_at
		) VALUES (
			:id, :quiz_id, :deck_id, :category_id,
			:question_types, :directions, :generation_mode,
			:llm_prompt, :manual_data, :question_count, :created_at
		)
	`, t)
	return err
}

func (r *QuestionTemplateRepo) GetByQuizID(ctx context.Context, quizID int64) ([]*model.QuestionTemplate, error) {
	var templates []*model.QuestionTemplate
	err := r.db.SelectContext(ctx, &templates, `
		SELECT id, quiz_id, deck_id, category_id,
		       question_types, directions, generation_mode,
		       llm_prompt, manual_data, question_count, created_at
		FROM question_templates
		WHERE quiz_id = $1
	`, quizID)
	return templates, err
}

func (r *QuestionTemplateRepo) DeleteByQuizID(ctx context.Context, quizID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM question_templates WHERE quiz_id = $1`, quizID)
	return err
}

func (r *QuestionTemplateRepo) Update(ctx context.Context, t *model.QuestionTemplate) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE question_templates
		SET
			deck_id = :deck_id,
			category_id = :category_id,
			question_types = :question_types,
			directions = :directions,
			generation_mode = :generation_mode,
			llm_prompt = :llm_prompt,
			manual_data = :manual_data,
			question_count = :question_count
		WHERE id = :id
	`, t)
	return err
}

// ============================================================
// QuizAttemptRepo
// ============================================================

type QuizAttemptRepo struct {
	db *sqlx.DB
}

func NewQuizAttemptRepo(db *sqlx.DB) *QuizAttemptRepo {
	return &QuizAttemptRepo{db: db}
}

func (r *QuizAttemptRepo) Create(ctx context.Context, attempt *model.QuizAttempt) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_attempts (
			id, user_id, quiz_id, course_id, node_id, assignment_id,
			started_at, coins_earned
		) VALUES (
			:id, :user_id, :quiz_id, :course_id, :node_id, :assignment_id,
			:started_at, :coins_earned
		)
	`, attempt)
	return err
}

func (r *QuizAttemptRepo) GetByID(ctx context.Context, attemptID int64) (*model.QuizAttempt, error) {
	var attempt model.QuizAttempt
	err := r.db.GetContext(ctx, &attempt, `
		SELECT id, user_id, quiz_id, course_id, node_id, assignment_id,
		       started_at, completed_at, score, max_score,
		       percentage, passed, time_taken, coins_earned
		FROM quiz_attempts WHERE id = $1
	`, attemptID)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *QuizAttemptRepo) GetActiveByUserAndQuiz(ctx context.Context, userID, quizID int64) (*model.QuizAttempt, error) {
	var attempt model.QuizAttempt
	err := r.db.GetContext(ctx, &attempt, `
		SELECT id, user_id, quiz_id, course_id, node_id, assignment_id,
		       started_at, completed_at, score, max_score,
		       percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE user_id = $1 AND quiz_id = $2 AND completed_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, userID, quizID)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *QuizAttemptRepo) CompleteAttempt(ctx context.Context, attemptID int64, score, maxScore, percentage float64, passed bool, timeTaken, coinsEarned int) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `
		UPDATE quiz_attempts SET
			completed_at = $1, score = $2, max_score = $3,
			percentage = $4, passed = $5, time_taken = $6, coins_earned = $7
		WHERE id = $8
	`, now, score, maxScore, percentage, passed, timeTaken, coinsEarned, attemptID)
	return err
}

func (r *QuizAttemptRepo) GetUserAttempts(ctx context.Context, userID int64, limit int) ([]*model.QuizAttempt, error) {
	var attempts []*model.QuizAttempt
	err := r.db.SelectContext(ctx, &attempts, `
		SELECT id, user_id, quiz_id, course_id, node_id, assignment_id,
		       started_at, completed_at, score, max_score,
		       percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE user_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, userID, limit)
	return attempts, err
}

func (r *QuizAttemptRepo) HasPassedQuiz(ctx context.Context, userID, quizID int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM quiz_attempts
			WHERE user_id = $1 AND quiz_id = $2 AND passed = TRUE
		)
	`, userID, quizID)
	return exists, err
}

// CountCompletedByAssignmentAndUser returns how many completed attempts exist for a given assignment+user.
// Used to enforce max_retakes limits.
func (r *QuizAttemptRepo) CountCompletedByAssignmentAndUser(ctx context.Context, assignmentID int64, userID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quiz_attempts
		WHERE assignment_id = $1 AND user_id = $2 AND completed_at IS NOT NULL
	`, assignmentID, userID)
	return count, err
}

// GetLastAttemptByAssignmentAndUser returns the most recent completed attempt for a student's assignment.
func (r *QuizAttemptRepo) GetLastAttemptByAssignmentAndUser(ctx context.Context, assignmentID int64, userID int64) (*model.QuizAttempt, error) {
	var attempt model.QuizAttempt
	err := r.db.GetContext(ctx, &attempt, `
		SELECT id, user_id, quiz_id, course_id, node_id, assignment_id,
		       started_at, completed_at, score, max_score,
		       percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE assignment_id = $1 AND user_id = $2 AND completed_at IS NOT NULL
		ORDER BY completed_at DESC LIMIT 1
	`, assignmentID, userID)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

// ============================================================
// AttemptQuestionRepo
// ============================================================

type AttemptQuestionRepo struct {
	db *sqlx.DB
}

func NewAttemptQuestionRepo(db *sqlx.DB) *AttemptQuestionRepo {
	return &AttemptQuestionRepo{db: db}
}

func (r *AttemptQuestionRepo) Create(ctx context.Context, q *model.AttemptQuestion) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO attempt_questions (
			id, attempt_id, quiz_id, question_text, correct_answer,
			options, question_type, direction, display_order,
			source_entry_id, generation_mode, created_at
		) VALUES (
			:id, :attempt_id, :quiz_id, :question_text, :correct_answer,
			:options, :question_type, :direction, :display_order,
			:source_entry_id, :generation_mode, :created_at
		)
	`, q)
	return err
}

func (r *AttemptQuestionRepo) GetByAttemptID(ctx context.Context, attemptID int64) ([]*model.AttemptQuestion, error) {
	var questions []*model.AttemptQuestion
	err := r.db.SelectContext(ctx, &questions, `
		SELECT id, attempt_id, quiz_id, question_text, correct_answer,
		       options, question_type, direction, display_order,
		       source_entry_id, generation_mode, created_at
		FROM attempt_questions
		WHERE attempt_id = $1
		ORDER BY display_order
	`, attemptID)
	return questions, err
}

func (r *AttemptQuestionRepo) GetByID(ctx context.Context, questionID int64) (*model.AttemptQuestion, error) {
	var q model.AttemptQuestion
	err := r.db.GetContext(ctx, &q, `
		SELECT id, attempt_id, quiz_id, question_text, correct_answer,
		       options, question_type, direction, display_order,
		       source_entry_id, generation_mode, created_at
		FROM attempt_questions WHERE id = $1
	`, questionID)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// ============================================================
// UserAnswerRepo
// ============================================================

type UserAnswerRepo struct {
	db *sqlx.DB
}

func NewUserAnswerRepo(db *sqlx.DB) *UserAnswerRepo {
	return &UserAnswerRepo{db: db}
}

func (r *UserAnswerRepo) Create(ctx context.Context, answer *model.UserAnswer) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO user_answers (
			id, attempt_id, question_id, user_answer,
			is_correct, points_earned, ai_explanation, answered_at
		) VALUES (
			:id, :attempt_id, :question_id, :user_answer,
			:is_correct, :points_earned, :ai_explanation, :answered_at
		)
		ON CONFLICT (attempt_id, question_id) DO UPDATE SET
			user_answer    = EXCLUDED.user_answer,
			is_correct     = EXCLUDED.is_correct,
			points_earned  = EXCLUDED.points_earned,
			ai_explanation = EXCLUDED.ai_explanation,
			answered_at    = EXCLUDED.answered_at
	`, answer)
	return err
}

func (r *UserAnswerRepo) GetByAttemptID(ctx context.Context, attemptID int64) ([]*model.UserAnswer, error) {
	var answers []*model.UserAnswer
	err := r.db.SelectContext(ctx, &answers, `
		SELECT DISTINCT ON (question_id)
		       id, attempt_id, question_id, user_answer,
		       is_correct, points_earned, ai_explanation, answered_at
		FROM user_answers
		WHERE attempt_id = $1
		ORDER BY question_id, answered_at DESC
	`, attemptID)
	return answers, err
}

// ============================================================
// CoinRepo
// ============================================================

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
		FROM user_coins WHERE user_id = $1
	`, userID)
	if err != nil {
		return &model.UserCoins{
			UserID:         userID,
			Balance:        0,
			LifetimeEarned: 0,
			LastUpdated:    time.Now().Unix(),
		}, nil
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
