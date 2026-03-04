package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

// ============================================================
// CourseRepo
// ============================================================

type CourseRepo struct {
	db *sqlx.DB
}

func NewCourseRepo(db *sqlx.DB) *CourseRepo {
	return &CourseRepo{db: db}
}

func (r *CourseRepo) Create(ctx context.Context, course *model.Course) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO courses (
			id, title, description, icon, color,
			deck_id, is_default, is_published,
			created_by, created_at, updated_at
		) VALUES (
			:id, :title, :description, :icon, :color,
			:deck_id, :is_default, :is_published,
			:created_by, :created_at, :updated_at
		)
	`, course)
	return err
}

func (r *CourseRepo) GetByID(ctx context.Context, id string) (*model.Course, error) {
	var course model.Course
	err := r.db.GetContext(ctx, &course, `
		SELECT id, title, description, icon, color,
		       deck_id, is_default, is_published,
		       created_by, created_at, updated_at
		FROM courses WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepo) GetAll(ctx context.Context) ([]*model.Course, error) {
	var courses []*model.Course
	err := r.db.SelectContext(ctx, &courses, `
		SELECT id, title, description, icon, color,
		       deck_id, is_default, is_published,
		       created_by, created_at, updated_at
		FROM courses
		WHERE is_published = TRUE
		ORDER BY is_default DESC, created_at
	`)
	return courses, err
}

func (r *CourseRepo) GetDefaults(ctx context.Context) ([]*model.Course, error) {
	var courses []*model.Course
	err := r.db.SelectContext(ctx, &courses, `
		SELECT id, title, description, icon, color,
		       deck_id, is_default, is_published,
		       created_by, created_at, updated_at
		FROM courses
		WHERE is_default = TRUE AND is_published = TRUE
		ORDER BY created_at
	`)
	return courses, err
}

func (r *CourseRepo) Update(ctx context.Context, course *model.Course) error {
	course.UpdatedAt = time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `
		UPDATE courses SET
			title=$1, description=$2, icon=$3, color=$4,
			is_default=$5, is_published=$6, updated_at=$7
		WHERE id=$8
	`, course.Title, course.Description, course.Icon, course.Color,
		course.IsDefault, course.IsPublished, course.UpdatedAt, course.ID)
	return err
}

func (r *CourseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM courses WHERE id = $1`, id)
	return err
}

func (r *CourseRepo) GetByDeckID(ctx context.Context, deckID int64) ([]*model.Course, error) {
	var courses []*model.Course
	err := r.db.SelectContext(ctx, &courses, `
		SELECT id, title, description, icon, color,
		       deck_id, is_default, is_published,
		       created_by, created_at, updated_at
		FROM courses
		WHERE deck_id = $1
		ORDER BY created_at
	`, deckID)
	return courses, err
}

func (r *CourseRepo) GetAllAdmin(ctx context.Context) ([]*model.Course, error) {
	var courses []*model.Course
	err := r.db.SelectContext(ctx, &courses, `
		SELECT id, title, description, icon, color,
		       deck_id, is_default, is_published,
		       created_by, created_at, updated_at
		FROM courses
		ORDER BY created_at DESC
	`)
	return courses, err
}

// ============================================================
// CourseNodeRepo
// ============================================================

type CourseNodeRepo struct {
	db *sqlx.DB
}

func NewCourseNodeRepo(db *sqlx.DB) *CourseNodeRepo {
	return &CourseNodeRepo{db: db}
}

func (r *CourseNodeRepo) Create(ctx context.Context, node *model.CourseNode) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO course_nodes (
			id, course_id, node_type, title, description, icon,
			position_x, position_y, sort_order,
			quiz_config, lesson_content, metadata, created_at
		) VALUES (
			:id, :course_id, :node_type, :title, :description, :icon,
			:position_x, :position_y, :sort_order,
			:quiz_config, :lesson_content, :metadata, :created_at
		)
	`, node)
	return err
}

func (r *CourseNodeRepo) GetByCourseID(ctx context.Context, courseID string) ([]*model.CourseNode, error) {
	var nodes []*model.CourseNode
	err := r.db.SelectContext(ctx, &nodes, `
		SELECT id, course_id, node_type, title, description, icon,
		       position_x, position_y, sort_order,
		       quiz_config, lesson_content, metadata, created_at
		FROM course_nodes
		WHERE course_id = $1
		ORDER BY sort_order
	`, courseID)
	return nodes, err
}

func (r *CourseNodeRepo) GetByID(ctx context.Context, nodeID string) (*model.CourseNode, error) {
	var node model.CourseNode
	err := r.db.GetContext(ctx, &node, `
		SELECT id, course_id, node_type, title, description, icon,
		       position_x, position_y, sort_order,
		       quiz_config, lesson_content, metadata, created_at
		FROM course_nodes WHERE id = $1
	`, nodeID)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *CourseNodeRepo) Update(ctx context.Context, node *model.CourseNode) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE course_nodes SET
			title=$1, description=$2, icon=$3,
			position_x=$4, position_y=$5, sort_order=$6,
			quiz_config=$7, lesson_content=$8, metadata=$9
		WHERE id=$10
	`, node.Title, node.Description, node.Icon,
		node.PositionX, node.PositionY, node.SortOrder,
		node.QuizConfig, node.LessonContent, node.Metadata, node.ID)
	return err
}

func (r *CourseNodeRepo) Delete(ctx context.Context, nodeID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM course_nodes WHERE id = $1`, nodeID)
	return err
}

func (r *CourseNodeRepo) BatchCreate(ctx context.Context, nodes []*model.CourseNode) error {
	for _, node := range nodes {
		if err := r.Create(ctx, node); err != nil {
			return fmt.Errorf("batch create node %s: %w", node.ID, err)
		}
	}
	return nil
}

// ============================================================
// CourseEdgeRepo
// ============================================================

type CourseEdgeRepo struct {
	db *sqlx.DB
}

func NewCourseEdgeRepo(db *sqlx.DB) *CourseEdgeRepo {
	return &CourseEdgeRepo{db: db}
}

func (r *CourseEdgeRepo) Create(ctx context.Context, edge *model.CourseEdge) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO course_edges (
			id, course_id, source_node_id, target_node_id,
			label, edge_type, created_at
		) VALUES (
			:id, :course_id, :source_node_id, :target_node_id,
			:label, :edge_type, :created_at
		)
	`, edge)
	return err
}

func (r *CourseEdgeRepo) GetByCourseID(ctx context.Context, courseID string) ([]*model.CourseEdge, error) {
	var edges []*model.CourseEdge
	err := r.db.SelectContext(ctx, &edges, `
		SELECT id, course_id, source_node_id, target_node_id,
		       label, edge_type, created_at
		FROM course_edges WHERE course_id = $1
	`, courseID)
	return edges, err
}

func (r *CourseEdgeRepo) Delete(ctx context.Context, edgeID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM course_edges WHERE id = $1`, edgeID)
	return err
}

func (r *CourseEdgeRepo) DeleteByCourseID(ctx context.Context, courseID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM course_edges WHERE course_id = $1`, courseID)
	return err
}

func (r *CourseEdgeRepo) BatchCreate(ctx context.Context, edges []*model.CourseEdge) error {
	for _, edge := range edges {
		if err := r.Create(ctx, edge); err != nil {
			return fmt.Errorf("batch create edge %s: %w", edge.ID, err)
		}
	}
	return nil
}

// ============================================================
// EnrollmentRepo
// ============================================================

type EnrollmentRepo struct {
	db *sqlx.DB
}

func NewEnrollmentRepo(db *sqlx.DB) *EnrollmentRepo {
	return &EnrollmentRepo{db: db}
}

func (r *EnrollmentRepo) Create(ctx context.Context, enrollment *model.UserEnrollment) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO user_enrollments (
			id, user_id, course_id, status, progress,
			current_node_id, completed_nodes,
			enrolled_at, last_accessed_at
		) VALUES (
			:id, :user_id, :course_id, :status, :progress,
			:current_node_id, :completed_nodes,
			:enrolled_at, :last_accessed_at
		)
	`, enrollment)
	return err
}

func (r *EnrollmentRepo) GetByUserID(ctx context.Context, userID int64) ([]*model.UserEnrollment, error) {
	var enrollments []*model.UserEnrollment
	err := r.db.SelectContext(ctx, &enrollments, `
		SELECT id, user_id, course_id, status, progress,
		       current_node_id, completed_nodes,
		       enrolled_at, completed_at, last_accessed_at
		FROM user_enrollments
		WHERE user_id = $1
		ORDER BY last_accessed_at DESC
	`, userID)
	return enrollments, err
}

func (r *EnrollmentRepo) GetByUserAndCourse(ctx context.Context, userID int64, courseID string) (*model.UserEnrollment, error) {
	var enrollment model.UserEnrollment
	err := r.db.GetContext(ctx, &enrollment, `
		SELECT id, user_id, course_id, status, progress,
		       current_node_id, completed_nodes,
		       enrolled_at, completed_at, last_accessed_at
		FROM user_enrollments
		WHERE user_id = $1 AND course_id = $2
	`, userID, courseID)
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (r *EnrollmentRepo) UpdateProgress(ctx context.Context, enrollmentID string, progress float64, completedNodes string, currentNodeID *string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_enrollments SET
			progress = $1, completed_nodes = $2,
			current_node_id = $3, last_accessed_at = $4
		WHERE id = $5
	`, progress, completedNodes, currentNodeID, now, enrollmentID)
	return err
}

func (r *EnrollmentRepo) CompleteCourse(ctx context.Context, enrollmentID string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_enrollments SET
			status = 'completed', progress = 1.0,
			completed_at = $1, last_accessed_at = $1
		WHERE id = $2
	`, now, enrollmentID)
	return err
}

func (r *EnrollmentRepo) UpdateLastAccessed(ctx context.Context, enrollmentID string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_enrollments SET last_accessed_at = $1 WHERE id = $2
	`, now, enrollmentID)
	return err
}
