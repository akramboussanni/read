package admin

import (
	"strconv"

	"github.com/akramboussanni/gocode/internal/model"
)

// UserListResponse is a summary of a user for admin listing
type UserListResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	DisplayName    string `json:"display_name"`
	Role           string `json:"role"`
	IsAdmin        bool   `json:"is_admin"`
	EmailConfirmed bool   `json:"email_confirmed"`
	OnboardingDone bool   `json:"onboarding_completed"`
	ActiveCourseID string `json:"active_course_id,omitempty"`
	CreatedAt      string `json:"created_at"`
}

func ToUserListResponse(user *model.User) UserListResponse {
	activeCourse := ""
	if user.ActiveCourseID != nil {
		activeCourse = *user.ActiveCourseID
	}
	return UserListResponse{
		ID:             strconv.FormatInt(user.ID, 10),
		Username:       user.Username,
		Email:          user.Email,
		DisplayName:    user.DisplayName,
		Role:           user.Role,
		IsAdmin:        user.IsAdmin,
		EmailConfirmed: user.EmailConfirmed,
		OnboardingDone: user.OnboardingCompleted,
		ActiveCourseID: activeCourse,
		CreatedAt:      strconv.FormatInt(user.CreatedAt, 10),
	}
}

// UserDetailResponse is a detailed view of a user for admin
type UserDetailResponse struct {
	User          UserListResponse `json:"user"`
	CoinBalance   int              `json:"coin_balance"`
	CoursesCount  int              `json:"courses_count"`
	AttemptsCount int              `json:"attempts_count"`
}

// ChangePasswordRequest is the request body for changing a password
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password"`
}
