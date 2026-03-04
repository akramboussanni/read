package repo

import (
	"context"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type AssignmentCompletionRepo struct {
	db *sqlx.DB
}

func NewAssignmentCompletionRepo(db *sqlx.DB) *AssignmentCompletionRepo {
	return &AssignmentCompletionRepo{db: db}
}

func (r *AssignmentCompletionRepo) Create(ctx context.Context, ac *model.AssignmentCompletion) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO assignment_completions (id, assignment_id, student_id, attempt_id, score, max_score, percentage, passed, completed_at)
		VALUES (:id, :assignment_id, :student_id, :attempt_id, :score, :max_score, :percentage, :passed, :completed_at)
		ON CONFLICT (assignment_id, student_id) 
		DO UPDATE SET
			attempt_id = EXCLUDED.attempt_id,
			score = EXCLUDED.score,
			max_score = EXCLUDED.max_score,
			percentage = EXCLUDED.percentage,
			passed = EXCLUDED.passed,
			completed_at = EXCLUDED.completed_at
		WHERE EXCLUDED.percentage > assignment_completions.percentage
	`, ac)
	return err
}

func (r *AssignmentCompletionRepo) GetByAssignmentAndStudent(ctx context.Context, assignmentID int64, studentID int64) (*model.AssignmentCompletion, error) {
	var ac model.AssignmentCompletion
	err := r.db.GetContext(ctx, &ac, "SELECT * FROM assignment_completions WHERE assignment_id = $1 AND student_id = $2", assignmentID, studentID)
	return &ac, err
}

func (r *AssignmentCompletionRepo) ListByAssignment(ctx context.Context, assignmentID int64) ([]*model.AssignmentCompletion, error) {
	var acs []*model.AssignmentCompletion
	err := r.db.SelectContext(ctx, &acs, "SELECT * FROM assignment_completions WHERE assignment_id = $1 ORDER BY completed_at DESC", assignmentID)
	return acs, err
}

func (r *AssignmentCompletionRepo) ListByStudent(ctx context.Context, studentID int64) ([]*model.AssignmentCompletion, error) {
	var acs []*model.AssignmentCompletion
	err := r.db.SelectContext(ctx, &acs, "SELECT * FROM assignment_completions WHERE student_id = $1 ORDER BY completed_at DESC", studentID)
	return acs, err
}
