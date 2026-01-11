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
// @Description Add or update the user's email address and send verification email. If the current email is verified, the new email is stored as pending until confirmed. If unverified or no email, it updates immediately. Required for quiz creation.
// @Tags Account
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body AddEmailRequest true "Email, confirmation URL, and password (password required only when changing verified email)"
// @Success 200 {object} api.SuccessResponse "Email update initiated - verification email sent"
// @Failure 400 {object} api.ErrorResponse "Invalid email format, email already in use, or missing required password"
// @Failure 401 {object} api.ErrorResponse "Unauthorized or invalid password"
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

	// Security check: if user has a verified email, require password confirmation
	if user.Email != "" && user.EmailConfirmed {
		if req.Password == "" {
			applog.Warn("Password required to change verified email", "userID:", user.ID)
			api.WriteBadRequest(w, "Password confirmation required to change verified email")
			return
		}

		// Verify the password
		if !utils.ComparePassword(user.PasswordHash, req.Password) {
			applog.Warn("Invalid password during email change attempt", "userID:", user.ID)
			api.WriteUnauthorized(w)
			return
		}

		applog.Info("Password verified for email change", "userID:", user.ID)
	}

	// Check if email is already in use by another user
	existingUser, err := ar.UserRepo.GetUserByEmail(r.Context(), req.Email)
	if err == nil && existingUser != nil && existingUser.ID != user.ID {
		api.WriteBadRequest(w, "Email already in use")
		return
	}

	// If current email is verified, store new email as pending
	// If unverified or no email, update directly
	if user.Email != "" && user.EmailConfirmed {
		// Store as pending email - will be applied upon confirmation
		if err := ar.UserRepo.UpdatePendingEmail(r.Context(), user.ID, req.Email); err != nil {
			applog.Error("Failed to update pending email:", err)
			api.WriteInternalError(w)
			return
		}
	} else {
		// Update email directly for unverified/no email cases
		if err := ar.UserRepo.UpdateEmail(r.Context(), user.ID, req.Email); err != nil {
			applog.Error("Failed to update email:", err)
			api.WriteInternalError(w)
			return
		}
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

	applog.Info("Email update initiated successfully", "userID:", user.ID, "email:", req.Email)
	api.WriteMessage(w, 200, "message", "Verification email sent - confirm to complete email change")
}
