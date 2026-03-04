package course

import (
	"encoding/json"
	"net/http"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

// ListCourses returns all published courses
func (cr *CourseRouter) ListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := cr.repos.Course.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		return
	}
	if courses == nil {
		courses = []*model.Course{}
	}

	// Populate nodes and edges for each course
	for _, course := range courses {
		nodes, _ := cr.repos.CourseNode.GetByCourseID(r.Context(), course.ID)
		edges, _ := cr.repos.CourseEdge.GetByCourseID(r.Context(), course.ID)
		course.Nodes = nodes
		course.Edges = edges
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

// GetCourse returns a single course with its full tree
func (cr *CourseRouter) GetCourse(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	course, err := cr.repos.Course.GetByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	nodes, _ := cr.repos.CourseNode.GetByCourseID(r.Context(), courseID)
	edges, _ := cr.repos.CourseEdge.GetByCourseID(r.Context(), courseID)
	course.Nodes = nodes
	course.Edges = edges

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// Enroll enrolls the authenticated user in a course
func (cr *CourseRouter) Enroll(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	courseID := chi.URLParam(r, "courseID")

	// Check if already enrolled
	_, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), user.ID, courseID)
	if err == nil {
		http.Error(w, "Already enrolled", http.StatusConflict)
		return
	}

	err = cr.quizService.CompleteOnboarding(r.Context(), user.ID, []string{courseID})
	if err != nil {
		http.Error(w, "Failed to enroll", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "enrolled"})
}

// MyEnrollments returns the user's enrolled courses
func (cr *CourseRouter) MyEnrollments(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	enrollments, err := cr.repos.Enrollment.GetByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to get enrollments", http.StatusInternalServerError)
		return
	}

	// Enrich with course data
	type EnrichedEnrollment struct {
		*model.UserEnrollment
		Course *model.Course `json:"course"`
	}

	var result []EnrichedEnrollment
	for _, e := range enrollments {
		course, _ := cr.repos.Course.GetByID(r.Context(), e.CourseID)
		result = append(result, EnrichedEnrollment{
			UserEnrollment: e,
			Course:         course,
		})
	}

	if result == nil {
		result = []EnrichedEnrollment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// SetActiveCourse sets the user's active course
func (cr *CourseRouter) SetActiveCourse(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		CourseID string `json:"course_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := cr.repos.User.SetActiveCourse(r.Context(), user.ID, &body.CourseID)
	if err != nil {
		http.Error(w, "Failed to set active course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetCourseStatus returns the detailed status of each node in the course for the authenticated user
func (cr *CourseRouter) GetCourseStatus(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	courseID := chi.URLParam(r, "courseID")

	statuses, err := cr.quizService.GetNodeStatuses(r.Context(), user.ID, courseID)
	if err != nil {
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	// Get enrollment
	enrollment, err := cr.repos.Enrollment.GetByUserAndCourse(r.Context(), user.ID, courseID)

	response := map[string]interface{}{
		"node_statuses": statuses,
	}
	if err == nil {
		response["enrollment"] = enrollment
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
