package classroom

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (cr *ClassroomRouter) CreateClassroom(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	// Limit one class per user
	teacherClasses, _ := cr.repos.Classroom.ListByTeacher(r.Context(), user.ID)
	studentClasses, _ := cr.repos.Classroom.ListForStudent(r.Context(), user.ID)
	if len(teacherClasses) > 0 || len(studentClasses) > 0 {
		http.Error(w, "Vous ne pouvez avoir qu'une seule classe pour le moment.", http.StatusForbidden)
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	class := &model.Classroom{
		ID:          utils.GenerateID(),
		Name:        body.Name,
		Description: body.Description,
		TeacherID:   user.ID,
		JoinCode:    utils.GenerateJoinCode(), // Implement this in utils
		CreatedAt:   time.Now().Unix(),
	}

	if err := cr.repos.Classroom.Create(r.Context(), class); err != nil {
		http.Error(w, "Failed to create classroom", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(class)
}

func (cr *ClassroomRouter) JoinByCode(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	code := chi.URLParam(r, "code")

	class, err := cr.repos.Classroom.GetByJoinCode(r.Context(), code)
	if err != nil {
		http.Error(w, "Classroom not found", http.StatusNotFound)
		return
	}

	// Teachers cannot join classes
	if user.Role == "teacher" {
		http.Error(w, "Les enseignants ne peuvent pas rejoindre de classes.", http.StatusForbidden)
		return
	}

	// Limit one class per user
	teacherClasses, _ := cr.repos.Classroom.ListByTeacher(r.Context(), user.ID)
	studentClasses, _ := cr.repos.Classroom.ListForStudent(r.Context(), user.ID)
	if len(teacherClasses) > 0 || len(studentClasses) > 0 {
		http.Error(w, "Vous ne pouvez avoir qu'une seule classe pour le moment.", http.StatusForbidden)
		return
	}

	if class.IsLocked {
		http.Error(w, "This classroom is locked", http.StatusForbidden)
		return
	}

	if err := cr.repos.Classroom.EnrollStudent(r.Context(), class.ID, user.ID); err != nil {
		http.Error(w, "Failed to join", http.StatusInternalServerError)
		return
	}

	// Auto-enroll the student in all class courses
	courses, _ := cr.repos.Classroom.GetCourses(r.Context(), class.ID)
	for _, c := range courses {
		_, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), user.ID, c.ID)
		if err != nil {
			cr.repos.Enrollment.Create(r.Context(), &model.UserEnrollment{
				ID:             utils.GenerateUUID(),
				UserID:         user.ID,
				CourseID:       c.ID,
				Status:         "active",
				EnrolledAt:     time.Now().Unix(),
				LastAccessedAt: time.Now().Unix(),
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "joined", "name": class.Name, "id": strconv.FormatInt(class.ID, 10)})
}

func (cr *ClassroomRouter) ListMyClasses(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	// If teacher, show classes they lead
	// If student, show classes they're in
	// For now, let's return both in a single response or separate

	teacherClasses, _ := cr.repos.Classroom.ListByTeacher(r.Context(), user.ID)
	if teacherClasses == nil {
		teacherClasses = []*model.Classroom{}
	}
	studentClasses, _ := cr.repos.Classroom.ListForStudent(r.Context(), user.ID)
	if studentClasses == nil {
		studentClasses = []*model.Classroom{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"teaching": teacherClasses,
		"enrolled": studentClasses,
	})
}

func (cr *ClassroomRouter) GetClassDetails(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	class, err := cr.repos.Classroom.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	students, _ := cr.repos.Classroom.GetStudents(r.Context(), id)
	if students == nil {
		students = []*model.User{}
	}
	assignments, _ := cr.repos.Classroom.ListAssignments(r.Context(), id)
	if assignments == nil {
		assignments = []*model.ClassroomAssignment{}
	}

	type AssignmentDto struct {
		*model.ClassroomAssignment
		IsCompleted    bool    `json:"is_completed"`
		Score          float64 `json:"score,omitempty"`
		MaxScore       float64 `json:"max_score,omitempty"`
		ScorePercent   float64 `json:"score_percent,omitempty"`
		CompletedAt    int64   `json:"completed_at,omitempty"`
		CourseName     string  `json:"course_name,omitempty"`
		NodeTitle      string  `json:"node_title,omitempty"`
		CompletedCount int     `json:"completed_count"`
		TotalStudents  int     `json:"total_students"`
	}

	var asgnDtos []AssignmentDto
	user, _ := utils.UserFromContext(r.Context())

	for _, a := range assignments {
		dto := AssignmentDto{
			ClassroomAssignment: a,
			TotalStudents:       len(students),
		}

		// Enrich with course and node names
		if course, err := cr.repos.Course.GetByID(r.Context(), a.CourseID); err == nil {
			dto.CourseName = course.Title
		}
		if node, err := cr.repos.CourseNode.GetByID(r.Context(), a.NodeID); err == nil {
			dto.NodeTitle = node.Title
		}

		if a.AssignmentType == "path_progress" {
			// Count how many students have that node in their completed_nodes
			count := 0
			for _, st := range students {
				enrollment, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), st.ID, a.CourseID)
				if err == nil && nodeInCompleted(enrollment.CompletedNodes, a.NodeID) {
					count++
				}
			}
			dto.CompletedCount = count

			// Check current user
			if user != nil {
				enrollment, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), user.ID, a.CourseID)
				if err == nil && nodeInCompleted(enrollment.CompletedNodes, a.NodeID) {
					dto.IsCompleted = true
					dto.Score = 100
					dto.MaxScore = 100
					dto.ScorePercent = 100
				}
			}
		} else {
			// Count total completions for this assignment
			completions, _ := cr.repos.AssignmentCompletion.ListByAssignment(r.Context(), a.ID)
			dto.CompletedCount = len(completions)

			// Check current user's completion status
			if user != nil {
				completion, err := cr.repos.AssignmentCompletion.GetByAssignmentAndStudent(r.Context(), a.ID, user.ID)
				if err == nil && completion != nil {
					dto.IsCompleted = true
					dto.Score = completion.Score
					dto.MaxScore = completion.MaxScore
					dto.ScorePercent = completion.Percentage
					dto.CompletedAt = completion.CompletedAt
				}
			}
		}
		asgnDtos = append(asgnDtos, dto)
	}

	if asgnDtos == nil {
		asgnDtos = []AssignmentDto{}
	}

	courses, _ := cr.repos.Classroom.GetCourses(r.Context(), id)
	if courses == nil {
		courses = []*model.Course{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"classroom":   class,
		"students":    students,
		"assignments": asgnDtos,
		"courses":     courses,
	})
}

func (cr *ClassroomRouter) RemoveStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	studentIDStr := chi.URLParam(r, "studentID")
	studentID, _ := strconv.ParseInt(studentIDStr, 10, 64)
	user, _ := utils.UserFromContext(r.Context())

	class, _ := cr.repos.Classroom.GetByID(r.Context(), id)
	if class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := cr.repos.Classroom.RemoveStudent(r.Context(), id, studentID); err != nil {
		http.Error(w, "Failed to remove student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cr *ClassroomRouter) GetAssignmentStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	asgnIDStr := chi.URLParam(r, "asgnID")
	asgnID, _ := strconv.ParseInt(asgnIDStr, 10, 64)
	user, _ := utils.UserFromContext(r.Context())

	class, err := cr.repos.Classroom.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	asgn, err := cr.repos.Classroom.GetAssignmentByID(r.Context(), asgnID)
	if err != nil || asgn.ClassroomID != id {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	students, _ := cr.repos.Classroom.GetStudents(r.Context(), id)

	type StudentStat struct {
		StudentID   int64   `json:"student_id,string"`
		Username    string  `json:"username"`
		IsCompleted bool    `json:"is_completed"`
		Score       float64 `json:"score,omitempty"`
		MaxScore    float64 `json:"max_score,omitempty"`
		Percentage  float64 `json:"percentage,omitempty"`
		CompletedAt int64   `json:"completed_at,omitempty"`
	}

	stats := make([]StudentStat, 0, len(students))

	if asgn.AssignmentType == "path_progress" {
		// Completion = node appears in enrollment's completed_nodes
		for _, st := range students {
			stat := StudentStat{StudentID: st.ID, Username: st.Username}
			enrollment, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), st.ID, asgn.CourseID)
			if err == nil && nodeInCompleted(enrollment.CompletedNodes, asgn.NodeID) {
				stat.IsCompleted = true
				stat.Score = 100
				stat.MaxScore = 100
				stat.Percentage = 100
			}
			stats = append(stats, stat)
		}
	} else {
		// Build a map of completions for quick lookup
		completions, _ := cr.repos.AssignmentCompletion.ListByAssignment(r.Context(), asgnID)
		completionMap := make(map[int64]*model.AssignmentCompletion)
		for _, c := range completions {
			completionMap[c.StudentID] = c
		}

		for _, st := range students {
			stat := StudentStat{
				StudentID: st.ID,
				Username:  st.Username,
			}
			if c, ok := completionMap[st.ID]; ok {
				stat.IsCompleted = true
				stat.Score = c.Score
				stat.MaxScore = c.MaxScore
				stat.Percentage = c.Percentage
				stat.CompletedAt = c.CompletedAt
			}
			stats = append(stats, stat)
		}
	}

	json.NewEncoder(w).Encode(stats)
}

func (cr *ClassroomRouter) AddCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	user, _ := utils.UserFromContext(r.Context())

	class, _ := cr.repos.Classroom.GetByID(r.Context(), id)
	if class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		CourseID string `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := cr.repos.Classroom.AddCourse(r.Context(), id, body.CourseID); err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		return
	}

	// Auto-enroll teacher and all students in course
	_, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), user.ID, body.CourseID)
	if err != nil {
		cr.repos.Enrollment.Create(r.Context(), &model.UserEnrollment{
			ID:             utils.GenerateUUID(),
			UserID:         user.ID,
			CourseID:       body.CourseID,
			Status:         "active",
			EnrolledAt:     time.Now().Unix(),
			LastAccessedAt: time.Now().Unix(),
		})
	}

	students, _ := cr.repos.Classroom.GetStudents(r.Context(), id)
	for _, st := range students {
		_, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), st.ID, body.CourseID)
		if err != nil {
			cr.repos.Enrollment.Create(r.Context(), &model.UserEnrollment{
				ID:             utils.GenerateUUID(),
				UserID:         st.ID,
				CourseID:       body.CourseID,
				Status:         "active",
				EnrolledAt:     time.Now().Unix(),
				LastAccessedAt: time.Now().Unix(),
			})
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cr *ClassroomRouter) RemoveCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	courseID := chi.URLParam(r, "courseID")
	user, _ := utils.UserFromContext(r.Context())

	class, _ := cr.repos.Classroom.GetByID(r.Context(), id)
	if class.TeacherID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := cr.repos.Classroom.RemoveCourse(r.Context(), id, courseID); err != nil {
		http.Error(w, "Failed to remove course", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cr *ClassroomRouter) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	asgnIDStr := chi.URLParam(r, "asgnID")
	asgnID, _ := strconv.ParseInt(asgnIDStr, 10, 64)
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

	var body struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		DueDate      int64  `json:"due_date"`
		PassingGrade int    `json:"passing_grade"`
		MaxRetakes   int    `json:"max_retakes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	asgn.Title = body.Title
	asgn.Description = body.Description
	asgn.DueDate = body.DueDate
	asgn.PassingGrade = body.PassingGrade
	asgn.MaxRetakes = body.MaxRetakes

	if err := cr.repos.Classroom.UpdateAssignment(r.Context(), asgn); err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(asgn)
}

func (cr *ClassroomRouter) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	asgnIDStr := chi.URLParam(r, "asgnID")
	asgnID, _ := strconv.ParseInt(asgnIDStr, 10, 64)
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

	if err := cr.repos.Classroom.DeleteAssignment(r.Context(), asgnID); err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// nodeInCompleted checks whether nodeID is present in a JSON-encoded array of node IDs
// (the completed_nodes field of user_enrollments).
func nodeInCompleted(completedNodesJSON, nodeID string) bool {
	if completedNodesJSON == "" || completedNodesJSON == "null" || completedNodesJSON == "[]" {
		return false
	}
	var ids []string
	if err := json.Unmarshal([]byte(completedNodesJSON), &ids); err != nil {
		return false
	}
	for _, id := range ids {
		if id == nodeID {
			return true
		}
	}
	return false
}
