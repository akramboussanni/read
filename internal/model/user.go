package model

// @Description User model with profile information
type User struct {
	ID                    int64   `db:"id" safe:"true" json:"id,string" example:"123456789"`
	Username              string  `db:"username" safe:"true" json:"username" example:"johndoe"`
	Email                 string  `db:"email" safe:"true" json:"email" example:"john@example.com"`
	PasswordHash          string  `db:"password_hash" json:"-"`
	DisplayName           string  `db:"display_name" safe:"true" json:"display_name"`
	AvatarURL             string  `db:"avatar_url" safe:"true" json:"avatar_url"`
	OnboardingCompleted   bool    `db:"onboarding_completed" safe:"true" json:"onboarding_completed"`
	ActiveCourseID        *string `db:"active_course_id" safe:"true" json:"active_course_id,omitempty"`
	CreatedAt             int64   `db:"created_at" safe:"true" json:"created_at,string" example:"1640995200"`
	Role                  string  `db:"user_role" safe:"true" json:"role" example:"user"`
	EmailConfirmed        bool    `db:"email_confirmed" safe:"true" json:"email_confirmed" example:"false"`
	PendingEmail          string  `db:"pending_email" json:"-"`
	EmailConfirmToken     string  `db:"email_confirm_token" json:"-"`
	EmailConfirmIssuedAt  int64   `db:"email_confirm_issuedat" json:"-"`
	PasswordResetToken    string  `db:"password_reset_token" json:"-"`
	PasswordResetIssuedAt int64   `db:"password_reset_issuedat" json:"-"`
	JwtSessionID          int64   `db:"jwt_session_id" json:"-"`
	IsAdmin               bool    `db:"is_admin" safe:"true" json:"is_admin" example:"false"`
	Coins                 int     `db:"balance" safe:"true" json:"coins" example:"100"`
}
