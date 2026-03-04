package classroom

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/akramboussanni/gocode/internal/middleware"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

type ClassroomRouter struct {
	repos *repo.Repos
}

func NewClassroomRouter(repos *repo.Repos) http.Handler {
	cr := &ClassroomRouter{repos: repos}
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, repos.User, repos.Token)
		r.Use(middleware.RequireEmailConfirmed)

		r.Get("/my", cr.ListMyClasses)
		r.Post("/join/{code}", cr.JoinByCode)
		r.Get("/{id}", cr.GetClassDetails)

		// Teachers only:
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireTeacher)

			r.Post("/create", cr.CreateClassroom)
			r.Put("/{id}", cr.UpdateClassroom)
			r.Post("/{id}/assignments", cr.CreateAssignment)
			r.Put("/{id}/assignments/{asgnID}", cr.UpdateAssignment)
			r.Delete("/{id}/assignments/{asgnID}", cr.DeleteAssignment)
			r.Get("/{id}/assignments/{asgnID}/stats", cr.GetAssignmentStats)
			r.Get("/{id}/assignments/{asgnID}/students/{studentID}", cr.GetStudentAssignmentDetail)

			r.Delete("/{id}/students/{studentID}", cr.RemoveStudent)
			r.Post("/{id}/courses", cr.AddCourse)
			r.Delete("/{id}/courses/{courseID}", cr.RemoveCourse)
		})
	})

	return r
}

func (cr *ClassroomRouter) UpdateClassroom(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	class, err := cr.repos.Classroom.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsLocked    bool   `json:"is_locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	class.Name = body.Name
	class.Description = body.Description
	class.IsLocked = body.IsLocked

	if err := cr.repos.Classroom.Update(r.Context(), class); err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(class)
}

func (cr *ClassroomRouter) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	classID, _ := strconv.ParseInt(idStr, 10, 64)

	class, err := cr.repos.Classroom.GetByID(r.Context(), classID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		CourseID     string `json:"course_id"`
		NodeID       string `json:"node_id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		DueDate      int64  `json:"due_date"`
		PassingGrade int    `json:"passing_grade"` // 0-100
		MaxRetakes   int    `json:"max_retakes"`   // -1=unlimited
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Default passing grade to 70% if not specified
	if body.PassingGrade == 0 {
		body.PassingGrade = 70
	}
	// Default max retakes to unlimited
	if body.MaxRetakes == 0 {
		body.MaxRetakes = -1
	}

	asgn := &model.ClassroomAssignment{
		ID:           utils.GenerateID(),
		ClassroomID:  classID,
		CourseID:     body.CourseID,
		NodeID:       body.NodeID,
		Title:        body.Title,
		Description:  body.Description,
		DueDate:      body.DueDate,
		PassingGrade: body.PassingGrade,
		MaxRetakes:   body.MaxRetakes,
		CreatedAt:    time.Now().Unix(),
	}

	if err := cr.repos.Classroom.CreateAssignment(r.Context(), asgn); err != nil {
		http.Error(w, "Failed to create assignment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(asgn)
}

// GetStudentAssignmentDetail returns a student's answers for a given assignment (teacher use).
func (cr *ClassroomRouter) GetStudentAssignmentDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	asgnIDStr := chi.URLParam(r, "asgnID")
	asgnID, _ := strconv.ParseInt(asgnIDStr, 10, 64)
	studentIDStr := chi.URLParam(r, "studentID")
	studentID, _ := strconv.ParseInt(studentIDStr, 10, 64)
	user, _ := utils.UserFromContext(r.Context())

	class, err := cr.repos.Classroom.GetByID(r.Context(), id)
	if err != nil || class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	asgn, err := cr.repos.Classroom.GetAssignmentByID(r.Context(), asgnID)
	if err != nil || asgn.ClassroomID != id {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	// Count total attempts
	attemptCount, _ := cr.repos.QuizAttempt.CountCompletedByAssignmentAndUser(r.Context(), asgnID, studentID)

	// Get the most recent completed attempt
	attempt, err := cr.repos.QuizAttempt.GetLastAttemptByAssignmentAndUser(r.Context(), asgnID, studentID)
	if err != nil {
		// Student hasn't done it yet
		json.NewEncoder(w).Encode(map[string]interface{}{
			"attempt_count": 0,
			"attempt":       nil,
			"questions":     []interface{}{},
			"answers":       []interface{}{},
		})
		return
	}

	questions, _ := cr.repos.AttemptQuestion.GetByAttemptID(r.Context(), attempt.ID)
	answers, _ := cr.repos.UserAnswer.GetByAttemptID(r.Context(), attempt.ID)

	// Compute passed based on assignment's passing_grade
	passedAssignment := attempt.Percentage != nil && *attempt.Percentage >= float64(asgn.PassingGrade)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"attempt_count":     attemptCount,
		"attempt":           attempt,
		"questions":         questions,
		"answers":           answers,
		"passed_assignment": passedAssignment,
		"passing_grade":     asgn.PassingGrade,
	})
}
