package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

// @Summary List all users
// @Description Get a list of all users in the system (admin only)
// @Tags Admin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {array} UserListResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Get all users
	users, err := ar.Repos.User.GetAllUsers(r.Context(), limit, offset)
	if err != nil {
		applog.Error("Failed to get users:", err)
		api.WriteInternalError(w)
		return
	}

	// Convert to response format
	response := make([]UserListResponse, 0, len(users))
	for _, user := range users {
		response = append(response, ToUserListResponse(user))
	}

	api.WriteJSON(w, 200, response)
}

// @Summary Get user details
// @Description Get detailed information about a specific user including progression (admin only)
// @Tags Admin
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} UserDetailResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 404 {object} api.ErrorResponse "User not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users/{userID} [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleGetUserDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid user ID")
		return
	}

	// Get user
	user, err := ar.Repos.User.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		api.WriteNotFound(w, "User not found")
		return
	}

	// Get coin balance
	coins, err := ar.Repos.Coin.GetUserCoins(r.Context(), userID)
	coinBalance := 0
	if err == nil && coins != nil {
		coinBalance = coins.Balance
	}

	// Get enrollment count
	enrollments, _ := ar.Repos.Enrollment.GetByUserID(r.Context(), userID)
	coursesCount := len(enrollments)

	// Get attempts count
	attempts, _ := ar.Repos.QuizAttempt.GetUserAttempts(r.Context(), userID, 1000)
	attemptsCount := len(attempts)

	response := UserDetailResponse{
		User:          ToUserListResponse(user),
		CoinBalance:   coinBalance,
		CoursesCount:  coursesCount,
		AttemptsCount: attemptsCount,
	}

	api.WriteJSON(w, 200, response)
}

// @Summary Change user password
// @Description Change a user's password (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 404 {object} api.ErrorResponse "User not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users/{userID}/password [post]
// @Security BearerAuth
func (ar *AdminRouter) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid user ID")
		return
	}

	// Decode request
	req, err := api.DecodeJSON[ChangePasswordRequest](w, r)
	if err != nil {
		return
	}

	// Validate password
	if len(req.NewPassword) < 8 {
		api.WriteBadRequest(w, "Password must be at least 8 characters")
		return
	}

	// Get user
	user, err := ar.Repos.User.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		api.WriteNotFound(w, "User not found")
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		applog.Error("Failed to hash password:", err)
		api.WriteInternalError(w)
		return
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.JwtSessionID = time.Now().UnixMicro() // Invalidate all sessions

	if err := ar.Repos.User.UpdatePassword(r.Context(), user.ID, hashedPassword, user.JwtSessionID); err != nil {
		applog.Error("Failed to update password:", err)
		api.WriteInternalError(w)
		return
	}

	api.WriteMessage(w, 200, "message", "Password changed successfully")
}

// @Summary Delete user
// @Description Delete a user account (admin only)
// @Tags Admin
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 404 {object} api.ErrorResponse "User not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users/{userID} [delete]
// @Security BearerAuth
func (ar *AdminRouter) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid user ID")
		return
	}

	// Prevent deleting yourself
	admin, _ := utils.UserFromContext(r.Context())
	if admin.ID == userID {
		api.WriteBadRequest(w, "Cannot delete your own account")
		return
	}

	// Check user exists
	user, err := ar.Repos.User.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		api.WriteNotFound(w, "User not found")
		return
	}

	// Delete user
	if err := ar.Repos.User.DeleteUser(r.Context(), userID); err != nil {
		applog.Error("Failed to delete user:", err)
		api.WriteInternalError(w)
		return
	}

	api.WriteMessage(w, 200, "message", "User deleted successfully")
}
