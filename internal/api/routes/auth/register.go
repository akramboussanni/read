package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/config"
	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
)

// @Summary Register new user account
// @Description Register a new user account with username and password. Email is optional - only needed for quiz creation. If email is provided, a verification email will be sent.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Recaptcha-Token header string false "reCAPTCHA verification token (optional if reCAPTCHA is not configured)"
// @Param request body RegisterRequest true "User registration credentials (email optional)"
// @Success 200 {object} api.SuccessResponse "User account created successfully"
// @Failure 400 {object} api.ErrorResponse "Invalid credentials, duplicate username/email, or validation errors"
// @Failure 429 {object} api.ErrorResponse "Rate limit exceeded (2 requests per minute)"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (ar *AuthRouter) HandleRegister(w http.ResponseWriter, r *http.Request) {
	applog.Info("HandleRegister called", "remoteAddr:", utils.GetClientIP(r))
	req, err := api.DecodeJSON[RegisterRequest](w, r)
	if err != nil {
		applog.Error("Failed to decode register request:", err)
		return
	}

	if req.Username == "" || req.Password == "" {
		applog.Warn("Missing registration fields", "username:", req.Username)
		http.Error(w, "identifiants invalides", http.StatusBadRequest)
		return
	}

	// Trim email whitespace
	req.Email = strings.TrimSpace(req.Email)

	// Validate email only if provided
	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		applog.Warn("Invalid email format", "email:", req.Email)
		http.Error(w, "format d'email invalide", http.StatusBadRequest)
		return
	}

	if strings.Contains(req.Username, "@") || !utils.IsValidPassword(req.Password) {
		applog.Warn("Invalid registration credentials", "username:", req.Username)
		http.Error(w, "identifiants invalides", http.StatusBadRequest)
		return
	}

	// Role validation
	if req.Role == "" {
		req.Role = "student"
	}
	if req.Role != "student" && req.Role != "teacher" {
		http.Error(w, "rôle invalide", http.StatusBadRequest)
		return
	}

	// Teachers must have an email
	if req.Role == "teacher" && req.Email == "" {
		http.Error(w, "les enseignants doivent fournir un email pour vérification", http.StatusBadRequest)
		return
	}

	duplicate, err := ar.UserRepo.DuplicateName(r.Context(), req.Username)
	if err != nil {
		applog.Error("Failed to check duplicate username:", err)
		api.WriteInternalError(w)
		return
	}

	if duplicate {
		applog.Warn("Duplicate username registration attempt", "username:", req.Username)
		http.Error(w, "identifiants invalides", http.StatusBadRequest)
		return
	}

	// Check duplicate email only if provided
	if req.Email != "" {
		emailDuplicate, err := ar.UserRepo.DuplicateEmail(r.Context(), req.Email)
		if err != nil {
			applog.Error("Failed to check duplicate email:", err)
			api.WriteInternalError(w)
			return
		}
		if emailDuplicate {
			applog.Warn("Duplicate email registration attempt", "email:", req.Email)
			http.Error(w, "email déjà utilisé", http.StatusBadRequest)
			return
		}
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		applog.Error("Failed to hash password:", err)
		api.WriteInternalError(w)
		return
	}

	// If no email provided, mark as confirmed (no verification needed)
	// EXCEPT for teachers who MUST verify their email anyway
	emailConfirmed := req.Email == "" && req.Role != "teacher"
	user := &model.User{ID: utils.GenerateSnowflakeID(), Username: req.Username, PasswordHash: hash, Email: req.Email, CreatedAt: time.Now().UTC().Unix(), Role: req.Role, EmailConfirmed: emailConfirmed}

	if err := ar.UserRepo.CreateUser(r.Context(), user); err != nil {
		applog.Error("Failed to create user:", err)
		api.WriteInternalError(w)
		return
	}

	// Only send confirmation email if email was provided
	if req.Email != "" {
		expiryStr := utils.ExpiryToString(24 * 3600)
		confirmUrl := req.Url
		if confirmUrl == "" {
			confirmUrl = config.App.FrontendCors + "/confirm-email"
		}
		token, err := GenerateTokenAndSendEmail(user.Email, "confirmregister", "Confirmation d'email", confirmUrl, map[string]any{"Expiry": expiryStr, "Url": confirmUrl})
		if err != nil {
			applog.Error("Failed to send confirmation email:", err)
			// Don't fail registration if email fails, just warn
			applog.Warn("User created but email sending failed", "userID:", user.ID)
		} else {
			if err := ar.UserRepo.AssignUserConfirmToken(r.Context(), token.Hash, time.Now().UTC().Unix(), user.ID); err != nil {
				applog.Error("Failed to assign confirmation token:", err)
			}
		}
		applog.Info("User registered successfully with email", "userID:", user.ID, "email:", user.Email)
		api.WriteMessage(w, 200, "message", "compte créé - email de confirmation envoyé")
	} else {
		applog.Info("User registered successfully without email", "userID:", user.ID)
		api.WriteMessage(w, 200, "message", "compte créé")
	}
}
