package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"k8s-monitoring-app/internal/core"

	"golang.org/x/oauth2"
)

type Session struct {
	ID           string
	UserEmail    string
	UserName     string
	UserPicture  string
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session in the database
func CreateSession(ctx context.Context, email, name, picture string, token *oauth2.Token) (*Session, error) {
	db, err := core.ConnectDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // Session valid for 24 hours

	session := &Session{
		ID:           sessionID,
		UserEmail:    email,
		UserName:     name,
		UserPicture:  picture,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
	}

    query := `
        INSERT INTO sessions (id, user_email, user_name, user_picture, access_token, refresh_token, token_expiry, created_at, expires_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err = db.ExecContext(ctx, query,
		session.ID,
		session.UserEmail,
		session.UserName,
		session.UserPicture,
		session.AccessToken,
		session.RefreshToken,
		session.TokenExpiry,
		session.CreatedAt,
		session.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func GetSession(ctx context.Context, sessionID string) (*Session, error) {
	db, err := core.ConnectDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

    query := `
        SELECT id, user_email, user_name, user_picture, access_token, refresh_token, token_expiry, created_at, expires_at
        FROM sessions
        WHERE id = ? AND expires_at > DATETIME('now')
    `

	var session Session
	var userPicture sql.NullString
	err = db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserEmail,
		&session.UserName,
		&userPicture,
		&session.AccessToken,
		&session.RefreshToken,
		&session.TokenExpiry,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if userPicture.Valid {
		session.UserPicture = userPicture.String
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session by ID
func DeleteSession(ctx context.Context, sessionID string) error {
	db, err := core.ConnectDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

    query := `DELETE FROM sessions WHERE id = ?`
	_, err = db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func CleanupExpiredSessions(ctx context.Context) error {
	db, err := core.ConnectDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

    query := `DELETE FROM sessions WHERE expires_at <= DATETIME('now')`
	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired sessions\n", rowsAffected)
	}

	return nil
}

// UpdateSessionExpiry extends the session expiry time
func UpdateSessionExpiry(ctx context.Context, sessionID string) error {
	db, err := core.ConnectDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
    query := `UPDATE sessions SET expires_at = ? WHERE id = ?`

	_, err = db.ExecContext(ctx, query, expiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session expiry: %w", err)
	}

	return nil
}
