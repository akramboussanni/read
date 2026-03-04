package course

import (
	"encoding/json"
	"net/http"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
)

// AvailableCourses returns courses available for onboarding
func (or *OnboardingRouter) AvailableCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := or.repos.Course.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to get courses", http.StatusInternalServerError)
		return
	}

	// Count nodes per course for a summary
	type CourseSummary struct {
		*model.Course
		NodeCount int `json:"node_count"`
	}

	var summaries []CourseSummary
	for _, c := range courses {
		nodes, _ := or.repos.CourseNode.GetByCourseID(r.Context(), c.ID)
		summaries = append(summaries, CourseSummary{
			Course:    c,
			NodeCount: len(nodes),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// CompleteOnboarding completes onboarding by enrolling the user in selected courses
func (or *OnboardingRouter) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		CourseIDs   []string `json:"course_ids"`
		DisplayName string   `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(body.CourseIDs) == 0 {
		http.Error(w, "Select at least one course", http.StatusBadRequest)
		return
	}

	// Update display name if provided
	if body.DisplayName != "" {
		_ = or.repos.User.UpdateDisplayName(r.Context(), user.ID, body.DisplayName)
	}

	if err := or.quizService.CompleteOnboarding(r.Context(), user.ID, body.CourseIDs); err != nil {
		http.Error(w, "Failed to complete onboarding", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "completed"})
}
