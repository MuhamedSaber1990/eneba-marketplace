package auth

import (
	"context"
	"fmt"
	"time"
)

// The limits. A "window" is how far back attempts are counted.
const (
	maxFailedSignInsPerEmail = 5
	maxFailedSignInsPerIP    = 30
	signInWindow             = 15 * time.Minute

	maxSignUpsPerIP = 10
	signUpWindow    = time.Hour

	// Rows older than this are no longer counted by any limit and are
	// deleted whenever a new attempt is recorded.
	attemptRetention = 24 * time.Hour
)

// The kinds of event that are counted. They match the CHECK in the
// auth_attempts table.
const (
	kindSignInEmail = "signin_email"
	kindSignInIP    = "signin_ip"
	kindSignUpIP    = "signup_ip"
)

// countAttempts returns how many events of this kind and key happened within
// the window.
func (s *Store) countAttempts(ctx context.Context, kind, key string, window time.Duration) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM auth_attempts
		 WHERE kind = $1 AND key = $2 AND attempted_at > $3`,
		kind, key, time.Now().Add(-window),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting %s attempts: %w", kind, err)
	}
	return count, nil
}

// recordAttempt stores one event, and takes the chance to delete rows too
// old to matter, so the table does not grow forever.
func (s *Store) recordAttempt(ctx context.Context, kind, key string) error {
	_, err := s.db.Exec(ctx, "INSERT INTO auth_attempts (kind, key) VALUES ($1, $2)", kind, key)
	if err != nil {
		return fmt.Errorf("recording a %s attempt: %w", kind, err)
	}
	_, err = s.db.Exec(ctx, "DELETE FROM auth_attempts WHERE attempted_at < $1", time.Now().Add(-attemptRetention))
	if err != nil {
		return fmt.Errorf("deleting old attempts: %w", err)
	}
	return nil
}
