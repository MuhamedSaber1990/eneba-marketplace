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
	Name  string `json:"name"`
	// MarketingOptIn is the "send me news about deals" choice. Nothing is
	// ever sent: it is a preference the account page shows and can change.
	MarketingOptIn bool `json:"marketing_opt_in"`
}

// userColumns are the columns that make up a User, in the order scanUser
// reads them. Every query that returns a user selects exactly these.
const userColumns = "u.id, u.email, u.name, u.marketing_opt_in"

// scanUser reads one row selected with userColumns, plus any extra columns
// selected after them.
func scanUser(row pgx.Row, extra ...any) (User, error) {
	var user User
	dest := append([]any{&user.ID, &user.Email, &user.Name, &user.MarketingOptIn}, extra...)
	err := row.Scan(dest...)
	return user, err
}

// SignUpInput is everything the sign-up form sends.
type SignUpInput struct {
	Name            string
	Email           string
	Password        string
	PasswordConfirm string
	AcceptTerms     bool
	MarketingOptIn  bool
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
func (s *Store) SignUp(ctx context.Context, input SignUpInput, ip string) (User, string, error) {
	// Every check the form makes is made again here. The form can be
	// skipped by sending the request directly; this cannot.
	name := NormalizeName(input.Name)
	if err := ValidateName(name); err != nil {
		return User{}, "", err
	}
	email := NormalizeEmail(input.Email)
	if err := ValidateEmail(email); err != nil {
		return User{}, "", err
	}
	if err := ValidateNewPassword(email, input.Password, input.PasswordConfirm); err != nil {
		return User{}, "", err
	}
	if !input.AcceptTerms {
		return User{}, "", ValidationError{"You need to agree to the Terms and the Privacy Policy to create an account."}
	}
	password := input.Password

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

	// terms_accepted_at is the moment of signing up: the lines above refused
	// anyone who did not tick the box. now() is the database's clock.
	user := User{Email: email, Name: name, MarketingOptIn: input.MarketingOptIn}
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, terms_accepted_at, marketing_opt_in)
		VALUES ($1, $2, $3, now(), $4)
		RETURNING id`,
		email, hash, name, input.MarketingOptIn,
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
	var hash string
	var isDemo bool
	user, err := scanUser(s.db.QueryRow(ctx,
		"SELECT "+userColumns+", u.password_hash, u.is_demo FROM users u WHERE u.email = $1", email,
	), &hash, &isDemo)
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

// UpdateProfile changes the two things a person can change about their
// account: the name and the marketing choice. It returns the account as it
// now is.
//
// userID comes from the session, never from the request body, so a person
// can only ever change their own account.
func (s *Store) UpdateProfile(ctx context.Context, userID int64, name string, marketingOptIn bool) (User, error) {
	name = NormalizeName(name)
	if err := ValidateName(name); err != nil {
		return User{}, err
	}
	user, err := scanUser(s.db.QueryRow(ctx, `
		UPDATE users u
		   SET name = $2, marketing_opt_in = $3
		 WHERE u.id = $1
		RETURNING `+userColumns,
		userID, name, marketingOptIn))
	if err != nil {
		return User{}, fmt.Errorf("updating the profile: %w", err)
	}
	return user, nil
}

// isUniqueViolation reports whether err is PostgreSQL refusing a row because
// of a UNIQUE rule (error code 23505). For the users table that means the
// email is taken. Letting the database decide, and not checking first with a
// SELECT, is what makes two sign-ups at the same instant safe.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
