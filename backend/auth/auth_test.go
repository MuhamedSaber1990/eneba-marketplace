package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// Hashing at the real cost takes a few hundred milliseconds per call. The
// tests use the lowest cost bcrypt allows so they stay fast. What is being
// tested is the logic around the hash, not its strength.
func init() {
	bcryptCost = bcrypt.MinCost
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"already normal", "you@mail.com", "you@mail.com"},
		{"upper case", "You@Mail.COM", "you@mail.com"},
		{"spaces around", "  you@mail.com \n", "you@mail.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeEmail(tt.in); got != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		ok    bool
	}{
		{"plain address", "you@mail.com", true},
		{"dots and plus in the name", "first.last+tag@mail.co.uk", true},
		{"no at sign", "you.mail.com", false},
		{"nothing before the at sign", "@mail.com", false},
		{"no dot in the domain", "you@localhost", false},
		{"domain starts with a dot", "you@.com", false},
		{"domain ends with a dot", "you@mail.", false},
		{"two at signs", "you@mail@mail.com", false},
		{"a space inside", "you @mail.com", false},
		{"display-name form", "You <you@mail.com>", false},
		{"empty", "", false},
		{"too long", strings.Repeat("a", 250) + "@mail.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err == nil) != tt.ok {
				t.Errorf("ValidateEmail(%q) error = %v, want ok = %v", tt.email, err, tt.ok)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		ok       bool
	}{
		{"eight characters is enough", "12345678", true},
		{"seven is too short", "1234567", false},
		{"empty", "", false},
		{"a long passphrase", "correct horse battery staple", true},
		{"no symbol or digit is required", "onlyletters", true},
		{"exactly 72 bytes", strings.Repeat("a", 72), true},
		{"73 bytes is refused, not cut", strings.Repeat("a", 73), false},
		// Eight characters, but more than eight bytes: counted as characters.
		{"eight non-English letters", "пароль12", true},
		// 37 characters of two bytes each is 74 bytes: over the byte limit.
		{"many two-byte letters", strings.Repeat("й", 37), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err == nil) != tt.ok {
				t.Errorf("ValidatePassword error = %v, want ok = %v", err, tt.ok)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	const password = "correct horse battery staple"

	first, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if strings.Contains(first, password) {
		t.Error("the hash contains the password")
	}
	if first == second {
		t.Error("two hashes of the same password are equal, so no salt was used")
	}
	if !passwordMatches(first, password) || !passwordMatches(second, password) {
		t.Error("the right password does not match its hash")
	}
	if passwordMatches(first, "correct horse battery stapler") {
		t.Error("a wrong password matches")
	}
	if passwordMatches("not a hash at all", password) {
		t.Error("a damaged hash counts as a match")
	}
}

func TestDummyHashIsARealHash(t *testing.T) {
	// If the constant were mistyped, comparing against it would fail
	// instantly and the timing protection would be gone without any error.
	cost, err := bcrypt.Cost([]byte(dummyHash))
	if err != nil {
		t.Fatalf("dummyHash is not a bcrypt hash: %v", err)
	}
	if cost != 12 {
		t.Errorf("dummyHash has cost %d, want 12, the cost real accounts are hashed with", cost)
	}
	if passwordMatches(dummyHash, "") || passwordMatches(dummyHash, "password") {
		t.Error("an obvious password matches dummyHash")
	}
}

func TestTokens(t *testing.T) {
	first, err := newToken()
	if err != nil {
		t.Fatalf("newToken: %v", err)
	}
	second, _ := newToken()

	if first == second {
		t.Error("two tokens are equal")
	}
	if len(first) != 43 { // 32 bytes written in base64 without padding
		t.Errorf("token has %d characters, want 43", len(first))
	}
	if hashToken(first) != hashToken(first) {
		t.Error("hashing the same token twice gives different results")
	}
	if hashToken(first) == hashToken(second) {
		t.Error("different tokens hash to the same value")
	}
	if hashToken(first) == first || strings.Contains(hashToken(first), first) {
		t.Error("the stored hash reveals the token")
	}
}
