package repo

import (
	"context"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type ClassroomRepo struct {
	db *sqlx.DB
}

func NewClassroomRepo(db *sqlx.DB) *ClassroomRepo {
	return &ClassroomRepo{db: db}
}

func (r *ClassroomRepo) Create(ctx context.Context, class *model.Classroom) error {
	_, err := r.db.NamedExecContext(ctx, `
        INSERT INTO classrooms (id, name, description, teacher_id, join_code, is_locked, created_at)
        VALUES (:id, :name, :description, :teacher_id, :join_code, :is_locked, :created_at)
    `, class)
	return err
}

func (r *ClassroomRepo) GetByID(ctx context.Context, id int64) (*model.Classroom, error) {
	var class model.Classroom
	err := r.db.GetContext(ctx, &class, "SELECT * FROM classrooms WHERE id = $1", id)
	return &class, err
}

func (r *ClassroomRepo) GetByJoinCode(ctx context.Context, code string) (*model.Classroom, error) {
	var class model.Classroom
	err := r.db.GetContext(ctx, &class, "SELECT * FROM classrooms WHERE join_code = $1", code)
	return &class, err
}

func (r *ClassroomRepo) ListByTeacher(ctx context.Context, teacherID int64) ([]*model.Classroom, error) {
	var classes []*model.Classroom
	err := r.db.SelectContext(ctx, &classes, "SELECT * FROM classrooms WHERE teacher_id = $1 ORDER BY created_at DESC", teacherID)
	return classes, err
}

func (r *ClassroomRepo) EnrollStudent(ctx context.Context, classroomID int64, studentID int64) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO classroom_students (classroom_id, student_id, joined_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (classroom_id, student_id) DO NOTHING
    `, classroomID, studentID, time.Now().Unix())
	return err
}

func (r *ClassroomRepo) ListForStudent(ctx context.Context, studentID int64) ([]*model.Classroom, error) {
	var classes []*model.Classroom
	err := r.db.SelectContext(ctx, &classes, `
        SELECT c.* FROM classrooms c
        JOIN classroom_students cs ON c.id = cs.classroom_id
        WHERE cs.student_id = $1
        ORDER BY cs.joined_at DESC
    `, studentID)
	return classes, err
}

func (r *ClassroomRepo) GetStudents(ctx context.Context, classroomID int64) ([]*model.User, error) {
	var students []*model.User
	err := r.db.SelectContext(ctx, &students, `
        SELECT u.* FROM users u
        JOIN classroom_students cs ON u.id = cs.student_id
        WHERE cs.classroom_id = $1
    `, classroomID)
	return students, err
}

func (r *ClassroomRepo) Update(ctx context.Context, class *model.Classroom) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE classrooms SET name = $1, description = $2, is_locked = $3 WHERE id = $4
    `, class.Name, class.Description, class.IsLocked, class.ID)
	return err
}

func (r *ClassroomRepo) CreateAssignment(ctx context.Context, asgn *model.ClassroomAssignment) error {
	_, err := r.db.NamedExecContext(ctx, `
        INSERT INTO classroom_assignments (id, classroom_id, course_id, node_id, title, description, due_date, passing_grade, max_retakes, created_at)
        VALUES (:id, :classroom_id, :course_id, :node_id, :title, :description, :due_date, :passing_grade, :max_retakes, :created_at)
    `, asgn)
	return err
}

func (r *ClassroomRepo) ListAssignments(ctx context.Context, classroomID int64) ([]*model.ClassroomAssignment, error) {
	var asgns []*model.ClassroomAssignment
	err := r.db.SelectContext(ctx, &asgns, "SELECT * FROM classroom_assignments WHERE classroom_id = $1 ORDER BY created_at DESC", classroomID)
	return asgns, err
}

func (r *ClassroomRepo) GetAssignmentByID(ctx context.Context, id int64) (*model.ClassroomAssignment, error) {
	var asgn model.ClassroomAssignment
	err := r.db.GetContext(ctx, &asgn, "SELECT * FROM classroom_assignments WHERE id = $1", id)
	return &asgn, err
}

func (r *ClassroomRepo) UpdateAssignment(ctx context.Context, asgn *model.ClassroomAssignment) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE classroom_assignments
        SET title = $1, description = $2, due_date = $3, passing_grade = $4, max_retakes = $5
        WHERE id = $6
    `, asgn.Title, asgn.Description, asgn.DueDate, asgn.PassingGrade, asgn.MaxRetakes, asgn.ID)
	return err
}

func (r *ClassroomRepo) DeleteAssignment(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM classroom_assignments WHERE id = $1", id)
	return err
}

func (r *ClassroomRepo) RemoveStudent(ctx context.Context, classroomID int64, studentID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM classroom_students WHERE classroom_id = $1 AND student_id = $2", classroomID, studentID)
	return err
}

func (r *ClassroomRepo) AddCourse(ctx context.Context, classroomID int64, courseID string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO classroom_courses (classroom_id, course_id, added_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (classroom_id, course_id) DO NOTHING
    `, classroomID, courseID, time.Now().Unix())
	return err
}

func (r *ClassroomRepo) RemoveCourse(ctx context.Context, classroomID int64, courseID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM classroom_courses WHERE classroom_id = $1 AND course_id = $2", classroomID, courseID)
	return err
}

func (r *ClassroomRepo) GetCourses(ctx context.Context, classroomID int64) ([]*model.Course, error) {
	var courses []*model.Course
	err := r.db.SelectContext(ctx, &courses, `
        SELECT c.* FROM courses c
        JOIN classroom_courses cc ON c.id = cc.course_id
        WHERE cc.classroom_id = $1
        ORDER BY cc.added_at DESC
    `, classroomID)
	return courses, err
}
