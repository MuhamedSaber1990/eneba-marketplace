// Package auth handles accounts: signing up, signing in, sessions, and the
// limits on how often either can be tried. It knows nothing about HTTP or
// cookies; the handlers in httpapi deal with those.
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Errors the handlers translate into responses.
var (
	// ErrEmailTaken: sign-up with an address that already has an account.
	ErrEmailTaken = errors.New("an account with that email already exists")
	// ErrInvalidCredentials: sign-in failed. It deliberately does not say
	// whether the email was unknown or the password wrong.
	ErrInvalidCredentials = errors.New("wrong email or password")
	// ErrTooManyAttempts: a rate limit was reached.
	ErrTooManyAttempts = errors.New("too many attempts")
)

// User is a signed-in person as the rest of the program sees them. The
// password hash is deliberately not part of it, so it cannot leak into a
// response by accident.
type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// Store runs the account queries.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// SignUp creates an account and signs it in. It returns the user and the
// session token for the cookie. ip is the address the request came from,
// used only for the sign-up limit.
func (s *Store) SignUp(ctx context.Context, email, password, ip string) (User, string, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return User{}, "", err
	}
	if err := ValidatePassword(password); err != nil {
		return User{}, "", err
	}

	// Every sign-up from an address counts, successful or not, so the limit
	// also slows down someone testing which emails are registered.
	recent, err := s.countAttempts(ctx, kindSignUpIP, ip, signUpWindow)
	if err != nil {
		return User{}, "", err
	}
	if recent >= maxSignUpsPerIP {
		return User{}, "", ErrTooManyAttempts
	}
	if err := s.recordAttempt(ctx, kindSignUpIP, ip); err != nil {
		return User{}, "", err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return User{}, "", err
	}

	user := User{Email: email}
	err = s.db.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, hash,
	).Scan(&user.ID)
	if isUniqueViolation(err) {
		return User{}, "", ErrEmailTaken
	}
	if err != nil {
		return User{}, "", fmt.Errorf("creating the user: %w", err)
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return User{}, "", err
	}
	return user, token, nil
}

// dummyHash is a real bcrypt hash (cost 12) of a password nobody has. When
// someone tries to sign in with an email that has no account, the password
// is still compared against this hash. Without that, a sign-in for an
// unknown email would answer much faster than one for a known email, and the
// response time alone would reveal which addresses are registered.
//
// It is written out here, not computed at startup, because computing it
// would add a few hundred milliseconds to every cold start of the function.
const dummyHash = "$2a$12$WzwuYBHDVx0EU2lCQS86cejNoAmviB3OD4ieyiThhcLwiEsC8LN7q"

// SignIn checks an email and password and, if they match, starts a session
// and returns the user and the token for the cookie.
//
// An unknown email and a wrong password are indistinguishable from outside:
// same error, same work done.
func (s *Store) SignIn(ctx context.Context, email, password, ip string) (User, string, error) {
	email = NormalizeEmail(email)

	// Look the account up first. No row is not an error here.
	var user User
	var hash string
	var isDemo bool
	err := s.db.QueryRow(ctx,
		"SELECT id, email, password_hash, is_demo FROM users WHERE email = $1", email,
	).Scan(&user.ID, &user.Email, &hash, &isDemo)
	found := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", fmt.Errorf("loading the user: %w", err)
	}

	// The limits. The per-email count is kept for any email that is tried,
	// registered or not, so being limited says nothing about whether an
	// account exists.
	//
	// The demo account is left out of the per-email limit. Its password is
	// public, so there is nothing to protect by locking it, and locking it
	// would let anyone keep reviewers out. The per-address limit still
	// applies to it.
	if !isDemo {
		failures, err := s.countAttempts(ctx, kindSignInEmail, email, signInWindow)
		if err != nil {
			return User{}, "", err
		}
		if failures >= maxFailedSignInsPerEmail {
			return User{}, "", ErrTooManyAttempts
		}
	}
	failures, err := s.countAttempts(ctx, kindSignInIP, ip, signInWindow)
	if err != nil {
		return User{}, "", err
	}
	if failures >= maxFailedSignInsPerIP {
		return User{}, "", ErrTooManyAttempts
	}

	// Always compare against some hash, so both cases take the same time.
	if !found {
		hash = dummyHash
	}
	if !passwordMatches(hash, password) || !found {
		if err := s.recordAttempt(ctx, kindSignInEmail, email); err != nil {
			return User{}, "", err
		}
		if err := s.recordAttempt(ctx, kindSignInIP, ip); err != nil {
			return User{}, "", err
		}
		return User{}, "", ErrInvalidCredentials
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return User{}, "", err
	}
	return user, token, nil
}

// isUniqueViolation reports whether err is PostgreSQL refusing a row because
// of a UNIQUE rule (error code 23505). For the users table that means the
// email is taken. Letting the database decide, and not checking first with a
// SELECT, is what makes two sign-ups at the same instant safe.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
