package apitoken

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	Prefix     = "rpsn_"
	rawBytes   = 32
	tokenBytes = 64 // hex-encoded rawBytes
)

// Token is a 32-byte API bearer secret. Text form is rpsn_ plus 64 hex chars.
type Token [rawBytes]byte

// Scope is the privilege of a stored token.
type Scope string

const (
	ScopeRead  Scope = "read"
	ScopeWrite Scope = "write"
)

// Generate returns a new high-entropy API token.
func Generate() (Token, error) {
	var t Token
	if _, err := rand.Read(t[:]); err != nil {
		return Token{}, fmt.Errorf("could not generate API token: %w", err)
	}
	return t, nil
}

// Parse decodes an rpsn_-prefixed hex token.
func Parse(s string) (Token, error) {
	var t Token
	if err := t.UnmarshalText([]byte(strings.TrimSpace(s))); err != nil {
		return Token{}, err
	}
	return t, nil
}

// String returns the rpsn_-prefixed encoding.
func (t Token) String() string {
	b, _ := t.MarshalText()
	return string(b)
}

// MarshalText implements encoding.TextMarshaler.
func (t Token) MarshalText() ([]byte, error) {
	out := make([]byte, len(Prefix)+tokenBytes)
	copy(out, Prefix)
	hex.Encode(out[len(Prefix):], t[:])
	return out, nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Token) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if !strings.HasPrefix(s, Prefix) {
		return fmt.Errorf("API token must start with %s", Prefix)
	}
	body := strings.TrimPrefix(s, Prefix)
	if len(body) != tokenBytes {
		return fmt.Errorf("API token must be %s followed by %d hex characters", Prefix, tokenBytes)
	}
	n, err := hex.Decode(t[:], []byte(body))
	if err != nil || n != rawBytes {
		return fmt.Errorf("API token is not valid hex")
	}
	return nil
}

// Equal compares two tokens in constant time.
func (t Token) Equal(other Token) bool {
	return subtle.ConstantTimeCompare(t[:], other[:]) == 1
}

// Valid reports whether t is a non-zero token.
func (t Token) Valid() bool {
	return !t.IsZero()
}

// IsZero reports whether t is the zero value.
func (t Token) IsZero() bool {
	var z Token
	return subtle.ConstantTimeCompare(t[:], z[:]) == 1
}

// UnmarshalText implements encoding.TextUnmarshaler for Scope.
func (s *Scope) UnmarshalText(text []byte) error {
	switch Scope(strings.TrimSpace(string(text))) {
	case ScopeRead:
		*s = ScopeRead
	case ScopeWrite:
		*s = ScopeWrite
	default:
		return fmt.Errorf("unknown API token scope %q (want %q or %q)", text, ScopeRead, ScopeWrite)
	}
	return nil
}
