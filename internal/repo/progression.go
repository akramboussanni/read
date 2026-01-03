package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type ProgressionRepo struct {
	db *sqlx.DB
}

func NewProgressionRepo(db *sqlx.DB) *ProgressionRepo {
	return &ProgressionRepo{db: db}
}

// GetByUserID retrieves user's progression or creates a new one
func (r *ProgressionRepo) GetByUserID(ctx context.Context, userID int64) (*model.UserProgression, error) {
	var progression model.UserProgression
	err := r.db.GetContext(ctx, &progression, `
		SELECT user_id, current_level, unlocked_quiz_ids, last_completed_quiz_id,
		       total_quizzes_completed, total_coins_earned, streak_days,
		       last_activity_date, created_at, updated_at
		FROM user_progression
		WHERE user_id = $1
	`, userID)

	if err != nil {
		// Create new progression record
		now := time.Now().Unix()
		progression = model.UserProgression{
			UserID:                userID,
			CurrentLevel:          1,
			UnlockedQuizIDs:       "[]", // Empty JSON array
			TotalQuizzesCompleted: 0,
			TotalCoinsEarned:      0,
			StreakDays:            0,
			CreatedAt:             now,
		}

		if err := r.Create(ctx, &progression); err != nil {
			return nil, err
		}
	}

	return &progression, nil
}

// Create creates a new user progression record
func (r *ProgressionRepo) Create(ctx context.Context, progression *model.UserProgression) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO user_progression (
			user_id, current_level, unlocked_quiz_ids, last_completed_quiz_id,
			total_quizzes_completed, total_coins_earned, streak_days,
			last_activity_date, created_at, updated_at
		) VALUES (
			:user_id, :current_level, :unlocked_quiz_ids, :last_completed_quiz_id,
			:total_quizzes_completed, :total_coins_earned, :streak_days,
			:last_activity_date, :created_at, :updated_at
		)
	`, progression)
	return err
}

// Update updates user progression
func (r *ProgressionRepo) Update(ctx context.Context, progression *model.UserProgression) error {
	now := time.Now().Unix()
	progression.UpdatedAt = &now

	_, err := r.db.NamedExecContext(ctx, `
		UPDATE user_progression SET
			current_level = :current_level,
			unlocked_quiz_ids = :unlocked_quiz_ids,
			last_completed_quiz_id = :last_completed_quiz_id,
			total_quizzes_completed = :total_quizzes_completed,
			total_coins_earned = :total_coins_earned,
			streak_days = :streak_days,
			last_activity_date = :last_activity_date,
			updated_at = :updated_at
		WHERE user_id = :user_id
	`, progression)
	return err
}

// IsQuizUnlocked checks if a quiz is unlocked for the user
func (r *ProgressionRepo) IsQuizUnlocked(ctx context.Context, userID, quizID int64) (bool, error) {
	progression, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	// Parse unlocked quiz IDs
	var unlockedIDs []int64
	if progression.UnlockedQuizIDs != "" {
		if err := json.Unmarshal([]byte(progression.UnlockedQuizIDs), &unlockedIDs); err != nil {
			return false, err
		}
	}

	// Check if quiz is in unlocked list
	for _, id := range unlockedIDs {
		if id == quizID {
			return true, nil
		}
	}

	// Check if it's the first quiz (level_order = 0 or 1)
	var levelOrder int
	err = r.db.GetContext(ctx, &levelOrder, `
		SELECT level_order FROM quizzes WHERE id = $1 AND is_system = 1 AND is_active = 1
	`, quizID)
	if err != nil {
		return false, err
	}

	return levelOrder <= 1, nil
}

// UnlockQuiz adds a quiz to the user's unlocked list
func (r *ProgressionRepo) UnlockQuiz(ctx context.Context, userID, quizID int64) error {
	progression, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Parse existing unlocked IDs
	var unlockedIDs []int64
	if progression.UnlockedQuizIDs != "" {
		if err := json.Unmarshal([]byte(progression.UnlockedQuizIDs), &unlockedIDs); err != nil {
			return err
		}
	}

	// Check if already unlocked
	for _, id := range unlockedIDs {
		if id == quizID {
			return nil // Already unlocked
		}
	}

	// Add new quiz ID
	unlockedIDs = append(unlockedIDs, quizID)

	// Marshal back to JSON
	jsonData, err := json.Marshal(unlockedIDs)
	if err != nil {
		return err
	}

	progression.UnlockedQuizIDs = string(jsonData)
	return r.Update(ctx, progression)
}

// CompleteQuiz updates progression after completing a quiz
func (r *ProgressionRepo) CompleteQuiz(ctx context.Context, userID, quizID int64, coinsEarned int) error {
	progression, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	progression.LastCompletedQuizID = &quizID
	progression.TotalQuizzesCompleted++
	progression.TotalCoinsEarned += coinsEarned

	// Update streak
	today := time.Now().Format("2006-01-02")
	if progression.LastActivityDate != nil && *progression.LastActivityDate != "" {
		_, err := time.Parse("2006-01-02", *progression.LastActivityDate)
		if err == nil {
			yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			if *progression.LastActivityDate == yesterday {
				progression.StreakDays++
			} else if *progression.LastActivityDate != today {
				progression.StreakDays = 1
			}
		}
	} else {
		progression.StreakDays = 1
	}

	progression.LastActivityDate = &today

	// Check if should level up
	var nextLevelQuiz *int64
	err = r.db.GetContext(ctx, &nextLevelQuiz, `
		SELECT id FROM quizzes
		WHERE is_system = 1 AND is_active = 1
		  AND level_order = $1
		ORDER BY level_order
		LIMIT 1
	`, progression.CurrentLevel+1)

	if err == nil && nextLevelQuiz != nil {
		progression.CurrentLevel++
		// Unlock the next level's first quiz
		if err := r.UnlockQuiz(ctx, userID, *nextLevelQuiz); err != nil {
			return err
		}
	}

	return r.Update(ctx, progression)
}
