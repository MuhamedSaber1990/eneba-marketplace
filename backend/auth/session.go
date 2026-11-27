package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SessionLifetime is how long a sign-in lasts. It is fixed from the moment
// of signing in: using the site does not extend it.
const SessionLifetime = 7 * 24 * time.Hour

// ErrNoSession means the request carries no valid session: no token, an
// unknown one, a signed-out one or an expired one. Callers treat all of
// these the same way, as "not signed in".
var ErrNoSession = errors.New("no valid session")

// newToken makes the secret that goes into the browser's cookie: 32 bytes
// from the operating system's secure random source, written as text. There
// is nothing to decode in it and no way to guess the next one.
func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating a session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// hashToken is what the database stores in place of the token. SHA-256
// cannot be reversed, so a copy of the sessions table does not let anyone
// sign in. A fast hash is right here (unlike for passwords) because the
// token is 32 random bytes and cannot be guessed however fast the hash is.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// createSession signs a user in: it stores a new session and returns the
// token for the cookie. It also deletes that user's expired sessions, which
// is the only cleanup there is (no scheduled job).
func (s *Store) createSession(ctx context.Context, userID int64) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1 AND expires_at <= now()", userID)
	if err != nil {
		return "", fmt.Errorf("deleting expired sessions: %w", err)
	}
	_, err = s.db.Exec(ctx,
		"INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)",
		hashToken(token), userID, time.Now().Add(SessionLifetime))
	if err != nil {
		return "", fmt.Errorf("storing the session: %w", err)
	}
	return token, nil
}

// UserForToken returns the user a session token belongs to. The expiry is
// checked here, in the query, on every request: an expired session stops
// working even if the browser still sends its cookie.
func (s *Store) UserForToken(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrNoSession
	}

	user, err := scanUser(s.db.QueryRow(ctx, `
		SELECT `+userColumns+`
		  FROM sessions s
		  JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1
		   AND s.expires_at > now()`,
		hashToken(token),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNoSession
	}
	if err != nil {
		return User{}, fmt.Errorf("loading the session: %w", err)
	}
	return user, nil
}

// SignOut ends a session by deleting its row. From then on the token is
// worthless, including any copy of it. Signing out with an unknown token is
// not an error: the result, "not signed in", is the same.
func (s *Store) SignOut(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.db.Exec(ctx, "DELETE FROM sessions WHERE token_hash = $1", hashToken(token))
	if err != nil {
		return fmt.Errorf("deleting the session: %w", err)
	}
	return nil
}
