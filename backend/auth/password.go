package auth

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordChars = 8
	// bcrypt only reads the first 72 bytes of a password. A longer one is
	// refused, because silently ignoring its end would mean two different
	// passwords open the same account.
	maxPasswordBytes = 72
	maxEmailLength   = 254
)

// bcryptCost sets how slow hashing is. Each step up doubles the work. 12
// takes a few hundred milliseconds, which a person signing in does not
// notice and someone guessing millions of passwords does. It is a variable
// so the tests can lower it and run quickly.
var bcryptCost = 12

// ValidationError is a problem with what the visitor typed. Its message is
// written for the visitor and is safe to show.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

// NormalizeEmail trims and lowercases an address, so the same person always
// maps to the same account however they type it.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateEmail does a plain shape check: something, an @, something with a
// dot. Whether the address really exists could only be proven by sending
// mail to it, which this project does not do.
func ValidateEmail(email string) error {
	local, domain, found := strings.Cut(email, "@")
	switch {
	case len(email) > maxEmailLength:
		return ValidationError{"That email address is too long."}
	case !found, local == "", !strings.Contains(domain, "."),
		strings.HasPrefix(domain, "."), strings.HasSuffix(domain, "."),
		strings.ContainsAny(email, " \t\r\n<>(),;"), strings.Contains(domain, "@"):
		return ValidationError{"Enter a valid email address."}
	}
	return nil
}

// ValidatePassword checks length only. Rules such as "must contain a symbol"
// are left out on purpose: length is what makes a password hard to guess.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < minPasswordChars {
		return ValidationError{fmt.Sprintf("The password needs at least %d characters.", minPasswordChars)}
	}
	if len(password) > maxPasswordBytes {
		return ValidationError{fmt.Sprintf("The password can be at most %d bytes long.", maxPasswordBytes)}
	}
	return nil
}

// HashPassword turns a password into the text stored in the database. bcrypt
// adds a random salt each time, so the same password gives a different hash
// on every call and two users with the same password cannot be spotted.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// passwordMatches reports whether password is the one the hash was made
// from. Any failure, including a damaged hash, counts as "no".
func passwordMatches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
