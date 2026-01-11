package repo

import (
	"context"
	"fmt"

	"github.com/akramboussanni/gocode/internal/model"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	Columns
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	repo := &UserRepo{db: db}
	repo.Columns = ExtractColumns[model.User]()
	return repo
}

func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	query := fmt.Sprintf(
		"INSERT INTO users (%s) VALUES (%s)",
		r.AllRaw,
		r.AllPrefixed,
	)
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &user, query, id)
	return &user, err
}

func (r *UserRepo) GetUserByIDSafe(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", r.SafeRaw)
	err := r.db.GetContext(ctx, &user, query, id)
	return &user, err
}

func (r *UserRepo) DuplicateName(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username)
	return exists, err
}

func (r *UserRepo) DuplicateEmail(ctx context.Context, email string) (bool, error) {
	// Empty emails are not considered duplicates
	if email == "" {
		return false, nil
	}
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email)
	return exists, err
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT %s FROM users WHERE email=$1", r.AllRaw)
	err := r.db.GetContext(ctx, &user, query, email)
	return &user, err
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT %s FROM users WHERE username=$1", r.AllRaw)
	err := r.db.GetContext(ctx, &user, query, username)
	return &user, err
}

func (r *UserRepo) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepo) GetUserByConfirmationToken(ctx context.Context, tokenHash string) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT %s FROM users WHERE email_confirm_token = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &user, query, tokenHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) AssignUserConfirmToken(ctx context.Context, token string, iat int64, userID int64) error {
	query := `
		UPDATE users
		SET email_confirm_token = $1,
		    email_confirm_issuedat = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, token, iat, userID)
	return err
}

func (r *UserRepo) MarkUserConfirmed(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET email = CASE 
			WHEN pending_email != '' THEN pending_email 
			ELSE email 
		END,
		    email_confirmed = TRUE,
		    pending_email = '',
		    email_confirm_token = '',
		    email_confirm_issuedat = 0
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *UserRepo) AssignUserResetToken(ctx context.Context, token string, iat int64, userID int64) error {
	query := `
		UPDATE users
		SET password_reset_token = $1,
		    password_reset_issuedat = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, token, iat, userID)
	return err
}

func (r *UserRepo) GetUserByResetToken(ctx context.Context, tokenHash string) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT %s FROM users WHERE password_reset_token = $1", r.AllRaw)
	err := r.db.GetContext(ctx, &user, query, tokenHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) ChangeUserPassword(ctx context.Context, newPasswordHash string, userID int64) error {
	query := `
		UPDATE users
		SET password_hash = $1,
			password_reset_token = '',
		    password_reset_issuedat = 0
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, newPasswordHash, userID)
	return err
}

func (r *UserRepo) ChangeJwtSessionID(ctx context.Context, userID int64, newID int64) error {
	query := `
		UPDATE users
		SET jwt_session_id = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, newID, userID)
	return err
}

// GetAllUsers retrieves all users with pagination
func (r *UserRepo) GetAllUsers(ctx context.Context, limit, offset int) ([]*model.User, error) {
	var users []*model.User
	query := fmt.Sprintf("SELECT %s FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2", r.AllRaw)
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	return users, err
}

// UpdatePassword updates user password and session ID
func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string, sessionID int64) error {
	query := `
		UPDATE users
		SET password_hash = $1,
		    jwt_session_id = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, passwordHash, sessionID, userID)
	return err
}

// CountTotalUsers counts all users in the system
func (r *UserRepo) CountTotalUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users")
	return count, err
}

// CountActiveUsers counts users active within the last N days
func (r *UserRepo) CountActiveUsers(ctx context.Context, days int) (int, error) {
	var count int
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM user_progression
		WHERE last_activity_date >= DATE('now', '-' || $1 || ' days')
	`
	err := r.db.GetContext(ctx, &count, query, days)
	return count, err
}

// UpdateEmail updates user's email and resets email confirmation
func (r *UserRepo) UpdateEmail(ctx context.Context, userID int64, email string) error {
	query := `
		UPDATE users
		SET email = $1,
		    email_confirmed = FALSE,
		    email_confirm_token = '',
		    email_confirm_issuedat = 0
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, email, userID)
	return err
}

// UpdatePendingEmail stores a pending email change (for verified emails)
func (r *UserRepo) UpdatePendingEmail(ctx context.Context, userID int64, pendingEmail string) error {
	query := `
		UPDATE users
		SET pending_email = $1,
		    email_confirm_token = '',
		    email_confirm_issuedat = 0
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, pendingEmail, userID)
	return err
}