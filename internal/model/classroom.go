package model

type Classroom struct {
	ID          int64  `json:"id,string" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	TeacherID   int64  `json:"teacher_id,string" db:"teacher_id"`
	JoinCode    string `json:"join_code" db:"join_code"`
	IsLocked    bool   `json:"is_locked" db:"is_locked"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
}

type ClassroomStudent struct {
	ClassroomID int64 `json:"classroom_id,string" db:"classroom_id"`
	StudentID   int64 `json:"student_id,string" db:"student_id"`
	JoinedAt    int64 `json:"joined_at" db:"joined_at"`
}

type ClassroomAssignment struct {
	ID           int64  `json:"id,string" db:"id"`
	ClassroomID  int64  `json:"classroom_id,string" db:"classroom_id"`
	CourseID     string `json:"course_id" db:"course_id"`
	NodeID       string `json:"node_id" db:"node_id"`
	Title        string `json:"title" db:"title"`
	Description  string `json:"description" db:"description"`
	DueDate      int64  `json:"due_date" db:"due_date"`
	PassingGrade int    `json:"passing_grade" db:"passing_grade"` // 0-100, percentage required to pass
	MaxRetakes   int    `json:"max_retakes" db:"max_retakes"`     // -1=unlimited, 0=no retakes, N=N retakes allowed
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

type ClassroomCourse struct {
	ClassroomID int64  `json:"classroom_id,string" db:"classroom_id"`
	CourseID    string `json:"course_id" db:"course_id"`
	AddedAt     int64  `json:"added_at" db:"added_at"`
}

type AssignmentCompletion struct {
	ID           int64   `json:"id,string" db:"id"`
	AssignmentID int64   `json:"assignment_id,string" db:"assignment_id"`
	StudentID    int64   `json:"student_id,string" db:"student_id"`
	AttemptID    *int64  `json:"attempt_id,string,omitempty" db:"attempt_id"`
	Score        float64 `json:"score" db:"score"`
	MaxScore     float64 `json:"max_score" db:"max_score"`
	Percentage   float64 `json:"percentage" db:"percentage"`
	Passed       bool    `json:"passed" db:"passed"`
	CompletedAt  int64   `json:"completed_at" db:"completed_at"`
}
