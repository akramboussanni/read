package repo

import (
	"context"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type QuizAttemptRepo struct {
	db *sqlx.DB
}

func NewQuizAttemptRepo(db *sqlx.DB) *QuizAttemptRepo {
	return &QuizAttemptRepo{db: db}
}

// Create creates a new quiz attempt
func (r *QuizAttemptRepo) Create(ctx context.Context, attempt *model.QuizAttempt) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_attempts (
			id, user_id, quiz_id, started_at, completed_at,
			score, max_score, percentage, passed, time_taken, coins_earned
		) VALUES (
			:id, :user_id, :quiz_id, :started_at, :completed_at,
			:score, :max_score, :percentage, :passed, :time_taken, :coins_earned
		)
	`, attempt)
	return err
}

// GetByID retrieves a quiz attempt by ID
func (r *QuizAttemptRepo) GetByID(ctx context.Context, attemptID int64) (*model.QuizAttempt, error) {
	var attempt model.QuizAttempt
	err := r.db.GetContext(ctx, &attempt, `
		SELECT id, user_id, quiz_id, started_at, completed_at,
		       score, max_score, percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE id = $1
	`, attemptID)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

// GetByUserAndQuiz retrieves the latest attempt by a user for a specific quiz
func (r *QuizAttemptRepo) GetByUserAndQuiz(ctx context.Context, userID, quizID int64) (*model.QuizAttempt, error) {
	var attempt model.QuizAttempt
	err := r.db.GetContext(ctx, &attempt, `
		SELECT id, user_id, quiz_id, started_at, completed_at,
		       score, max_score, percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE user_id = $1 AND quiz_id = $2
		ORDER BY started_at DESC
		LIMIT 1
	`, userID, quizID)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

// Update updates a quiz attempt (typically when completing)
func (r *QuizAttemptRepo) Update(ctx context.Context, attempt *model.QuizAttempt) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE quiz_attempts SET
			completed_at = :completed_at,
			score = :score,
			max_score = :max_score,
			percentage = :percentage,
			passed = :passed,
			time_taken = :time_taken,
			coins_earned = :coins_earned
		WHERE id = :id
	`, attempt)
	return err
}

// GetBestAttemptsByUser retrieves user's best scores for each quiz
func (r *QuizAttemptRepo) GetBestAttemptsByUser(ctx context.Context, userID int64) (map[int64]float64, error) {
	type Result struct {
		QuizID int64   `db:"quiz_id"`
		Score  float64 `db:"best_score"`
	}

	var results []Result
	err := r.db.SelectContext(ctx, &results, `
		SELECT quiz_id, MAX(score) as best_score
		FROM quiz_attempts
		WHERE user_id = $1 AND completed_at IS NOT NULL
		GROUP BY quiz_id
	`, userID)

	if err != nil {
		return nil, err
	}

	scores := make(map[int64]float64)
	for _, r := range results {
		scores[r.QuizID] = r.Score
	}

	return scores, nil
}

// GetUserAttempts retrieves all attempts by a user
func (r *QuizAttemptRepo) GetUserAttempts(ctx context.Context, userID int64, limit, offset int) ([]*model.QuizAttempt, error) {
	var attempts []*model.QuizAttempt
	err := r.db.SelectContext(ctx, &attempts, `
		SELECT id, user_id, quiz_id, started_at, completed_at,
		       score, max_score, percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE user_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return attempts, err
}

// GetQuizAttempts retrieves all attempts for a specific quiz
func (r *QuizAttemptRepo) GetQuizAttempts(ctx context.Context, quizID int64, limit, offset int) ([]*model.QuizAttempt, error) {
	var attempts []*model.QuizAttempt
	err := r.db.SelectContext(ctx, &attempts, `
		SELECT id, user_id, quiz_id, started_at, completed_at,
		       score, max_score, percentage, passed, time_taken, coins_earned
		FROM quiz_attempts
		WHERE quiz_id = $1 AND completed_at IS NOT NULL
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`, quizID, limit, offset)
	return attempts, err
}

// CountAttemptsByQuiz counts total attempts for a quiz
func (r *QuizAttemptRepo) CountAttemptsByQuiz(ctx context.Context, quizID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quiz_attempts
		WHERE quiz_id = $1 AND completed_at IS NOT NULL
	`, quizID)
	return count, err
}

// GetAverageScoreByQuiz calculates average score for a quiz
func (r *QuizAttemptRepo) GetAverageScoreByQuiz(ctx context.Context, quizID int64) (*float64, error) {
	var avg *float64
	err := r.db.GetContext(ctx, &avg, `
		SELECT AVG(percentage) FROM quiz_attempts
		WHERE quiz_id = $1 AND completed_at IS NOT NULL
	`, quizID)
	return avg, err
}

// CountPassedAttempts counts attempts that passed
func (r *QuizAttemptRepo) CountPassedAttempts(ctx context.Context, quizID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quiz_attempts
		WHERE quiz_id = $1 AND completed_at IS NOT NULL AND passed = 1
	`, quizID)
	return count, err
}

// CountTotalAttempts counts all completed attempts
func (r *QuizAttemptRepo) CountTotalAttempts(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM quiz_attempts WHERE completed_at IS NOT NULL
	`)
	return count, err
}

// UserAnswerRepo handles user answer operations
type UserAnswerRepo struct {
	db *sqlx.DB
}

func NewUserAnswerRepo(db *sqlx.DB) *UserAnswerRepo {
	return &UserAnswerRepo{db: db}
}

// Create creates a new user answer
func (r *UserAnswerRepo) Create(ctx context.Context, answer *model.UserAnswer) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO user_answers (
			id, attempt_id, question_id, user_answer,
			is_correct, points_earned, answered_at
		) VALUES (
			:id, :attempt_id, :question_id, :user_answer,
			:is_correct, :points_earned, :answered_at
		)
	`, answer)
	return err
}

// GetByAttemptID retrieves all answers for an attempt
func (r *UserAnswerRepo) GetByAttemptID(ctx context.Context, attemptID int64) ([]*model.UserAnswer, error) {
	var answers []*model.UserAnswer
	err := r.db.SelectContext(ctx, &answers, `
		SELECT id, attempt_id, question_id, user_answer,
		       is_correct, points_earned, answered_at
		FROM user_answers
		WHERE attempt_id = $1
		ORDER BY answered_at
	`, attemptID)
	return answers, err
}

// GetByAttemptAndQuestion retrieves a specific answer
func (r *UserAnswerRepo) GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID int64) (*model.UserAnswer, error) {
	var answer model.UserAnswer
	err := r.db.GetContext(ctx, &answer, `
		SELECT id, attempt_id, question_id, user_answer,
		       is_correct, points_earned, answered_at
		FROM user_answers
		WHERE attempt_id = $1 AND question_id = $2
	`, attemptID, questionID)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// CountByAttemptID counts answers for an attempt
func (r *UserAnswerRepo) CountByAttemptID(ctx context.Context, attemptID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM user_answers WHERE attempt_id = $1
	`, attemptID)
	return count, err
}

// CountCorrectByAttemptID counts correct answers for an attempt
func (r *UserAnswerRepo) CountCorrectByAttemptID(ctx context.Context, attemptID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM user_answers WHERE attempt_id = $1 AND is_correct = 1
	`, attemptID)
	return count, err
}

// BatchCreate creates multiple user answers in a transaction
func (r *UserAnswerRepo) BatchCreate(ctx context.Context, answers []*model.UserAnswer) error {
	if len(answers) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, answer := range answers {
		if answer.AnsweredAt == 0 {
			answer.AnsweredAt = time.Now().Unix()
		}

		_, err := tx.NamedExecContext(ctx, `
			INSERT INTO user_answers (
				id, attempt_id, question_id, user_answer,
				is_correct, points_earned, answered_at
			) VALUES (
				:id, :attempt_id, :question_id, :user_answer,
				:is_correct, :points_earned, :answered_at
			)
		`, answer)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
