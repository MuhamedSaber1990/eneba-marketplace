package auth

import (
	"fmt"
	"strings"
	"unicode"
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
	// The same limit as the CHECK on users.name in the database.
	maxNameChars = 60
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

// ValidateNewPassword checks a password someone is choosing: the length
// rules, that it is not simply their email address, and that they typed the
// same thing twice.
//
// The sign-up form shows these same three rules as the visitor types
// (frontend/src/lib/passwordRules.ts). The form is a convenience. This
// function is the check that counts, because a request can be sent without
// the form.
func ValidateNewPassword(email, password, confirmation string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	// An email is public. Used as the password, it is the first thing
	// anyone would try.
	if strings.EqualFold(strings.TrimSpace(password), NormalizeEmail(email)) {
		return ValidationError{"The password cannot be your email address."}
	}
	if password != confirmation {
		return ValidationError{"The two passwords do not match."}
	}
	return nil
}

// NormalizeName trims a name and turns any run of spaces, tabs or line
// breaks inside it into one space: "  Layla   Hassan " becomes "Layla Hassan".
func NormalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// ValidateName checks a normalized name: something, not too long, and no
// invisible control characters. Letters of any alphabet, apostrophes,
// hyphens and the like are all fine. People's names are not ours to restrict.
func ValidateName(name string) error {
	switch {
	case name == "":
		return ValidationError{"Enter your name."}
	case utf8.RuneCountInString(name) > maxNameChars:
		return ValidationError{fmt.Sprintf("The name can be at most %d characters.", maxNameChars)}
	}
	for _, character := range name {
		if unicode.IsControl(character) {
			return ValidationError{"The name contains a character that cannot be used."}
		}
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
