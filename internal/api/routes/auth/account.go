package auth

import (
	"net/http"
	"time"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/utils"
)

// @Summary Get current user profile
// @Description Retrieve the current authenticated user's profile information. Returns safe user data (excluding sensitive fields like password hash).
// @Tags Account
// @Accept json
// @Produce json
// @Security CookieAuth
// @Success 200 {object} model.User "User profile information (safe fields only)"
// @Failure 401 {object} api.ErrorResponse "Unauthorized - invalid or missing session cookie"
// @Failure 429 {object} api.ErrorResponse "Rate limit exceeded (30 requests per minute)"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /auth/me [get]
func (ar *AuthRouter) HandleProfile(w http.ResponseWriter, r *http.Request) {
	applog.Info("HandleProfile called")
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		applog.Error("Failed to get user from context")
		api.WriteInternalError(w)
		return
	}

	utils.StripUnsafeFields(user)
	applog.Info("Profile retrieved", "userID:", user.ID)
	api.WriteJSON(w, 200, user)
}

// @Summary Add or update email address
// @Description Add or update the user's email address and send verification email. Required for quiz creation.
// @Tags Account
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body AddEmailRequest true "Email and confirmation URL"
// @Success 200 {object} api.SuccessResponse "Email updated - verification email sent"
// @Failure 400 {object} api.ErrorResponse "Invalid email format or email already in use"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 429 {object} api.ErrorResponse "Rate limit exceeded (30 requests per minute)"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /auth/me/email [post]
func (ar *AuthRouter) HandleAddEmail(w http.ResponseWriter, r *http.Request) {
	applog.Info("HandleAddEmail called")
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}

	req, err := api.DecodeJSON[AddEmailRequest](w, r)
	if err != nil {
		return
	}

	// Validate email format
	if !utils.IsValidEmail(req.Email) {
		api.WriteBadRequest(w, "Invalid email format")
		return
	}

	// Check if email is already in use by another user
	existingUser, err := ar.UserRepo.GetUserByEmail(r.Context(), req.Email)
	if err == nil && existingUser != nil && existingUser.ID != user.ID {
		api.WriteBadRequest(w, "Email already in use")
		return
	}

	// Update email and reset confirmation status
	if err := ar.UserRepo.UpdateEmail(r.Context(), user.ID, req.Email); err != nil {
		applog.Error("Failed to update email:", err)
		api.WriteInternalError(w)
		return
	}

	// Send confirmation email
	expiryStr := utils.ExpiryToString(24 * 3600)
	token, err := GenerateTokenAndSendEmail(req.Email, "confirmregister", "Email confirmation", req.Url, map[string]any{"Expiry": expiryStr, "Url": req.Url})
	if err != nil {
		applog.Error("Failed to send confirmation email:", err)
		api.WriteMessage(w, 200, "message", "Email updated but confirmation email failed to send")
		return
	}

	if err := ar.UserRepo.AssignUserConfirmToken(r.Context(), token.Hash, time.Now().UTC().Unix(), user.ID); err != nil {
		applog.Error("Failed to assign confirmation token:", err)
		api.WriteInternalError(w)
		return
	}

	applog.Info("Email updated successfully", "userID:", user.ID, "email:", req.Email)
	api.WriteMessage(w, 200, "message", "Email updated - verification email sent")
}
