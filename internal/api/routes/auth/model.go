package auth

// @Description User registration request - email is optional, only needed for quiz creation
type RegisterRequest struct {
	Username string `json:"username" example:"johndoe" binding:"required" minLength:"3" maxLength:"30" pattern:"^[a-zA-Z0-9_-]+$"`
	Email    string `json:"email" example:"john@example.com" format:"email" description:"Optional - only needed for quiz creation"`
	Password string `json:"password" example:"SecurePass123!" binding:"required" minLength:"8"`
	Url      string `json:"url" example:"http://localhost:3000/confirm-email" format:"uri" description:"Optional confirmation URL for email verification"`
	Role     string `json:"role" example:"student" description:"User role: student or teacher"`
}

// @Description User login credentials - uses username, not email
type LoginRequest struct {
	Username string `json:"username" example:"johndoe" binding:"required"`
	Password string `json:"password" example:"SecurePass123!" binding:"required"`
}

// @Description Add or update email address
type AddEmailRequest struct {
	Email    string `json:"email" example:"john@example.com" binding:"required" format:"email"`
	Url      string `json:"url" example:"https://example.com/confirm" format:"uri" description:"Confirmation URL for email verification"`
	Password string `json:"password" example:"SecurePass123!" description:"Required only when changing a verified email address"`
}

// @Description Email-based request for password reset and email confirmation resend
type EmailRequest struct {
	Email string `json:"email" example:"john@example.com" binding:"required" format:"email"`
	Url   string `json:"url" example:"https://example.com/reset" format:"uri" description:"Optional URL for email templates"`
}

// @Description Token-based request for various operations (email confirmation, password reset, token refresh)
type TokenRequest struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." binding:"required" description:"JWT token or base64 encoded token"`
	Url   string `json:"url" example:"https://example.com/reset" format:"uri" description:"Optional URL for email templates"`
}

// @Description Password reset request with token and new password
type PasswordResetRequest struct {
	Token       string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." binding:"required" description:"Password reset token from email"`
	NewPassword string `json:"new_password" example:"NewSecurePass123!" binding:"required" minLength:"8" description:"New password that meets security requirements"`
}

// @Description Password change request requiring current password verification
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" example:"SecurePass123!" binding:"required" description:"Current password for verification"`
	NewPassword string `json:"new_password" example:"NewSecurePass123!" binding:"required" minLength:"8" description:"New password that meets security requirements"`
}
