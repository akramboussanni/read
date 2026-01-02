package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
)

// GenerateRandomPassword generates a cryptographically secure random password
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 16
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Use base64 encoding for readable password
	password := base64.URLEncoding.EncodeToString(bytes)[:length]
	return password, nil
}

// InitializeDefaultAdmin creates a default admin user if no admin exists
// Returns the generated password if a new admin was created, empty string otherwise
func InitializeDefaultAdmin(ctx context.Context, userRepo *repo.UserRepo) (string, error) {
	// Check if admin already exists
	admin, err := userRepo.GetUserByEmail(ctx, "admin@localhost")
	if err == nil && admin != nil {
		// Admin already exists
		return "", nil
	}

	// Generate random password
	password, err := GenerateRandomPassword(24)
	if err != nil {
		return "", fmt.Errorf("failed to generate admin password: %w", err)
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Generate snowflake ID
	adminID := GenerateSnowflakeID()

	// Create admin user
	adminUser := &model.User{
		ID:                    adminID,
		Username:              "admin",
		Email:                 "admin@localhost",
		PasswordHash:          hashedPassword,
		CreatedAt:             time.Now().Unix(),
		Role:                  "admin",
		EmailConfirmed:        true, // Pre-confirmed for first login
		EmailConfirmToken:     "",
		EmailConfirmIssuedAt:  0,
		PasswordResetToken:    "",
		PasswordResetIssuedAt: 0,
		JwtSessionID:          time.Now().UnixMicro(),
		IsAdmin:               true,
	}

	if err := userRepo.CreateUser(ctx, adminUser); err != nil {
		return "", fmt.Errorf("failed to create admin user: %w", err)
	}

	return password, nil
}
