package apitoken

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	Prefix     = "rpsn_"
	rawBytes   = 32
	fileMode   = 0600
	dirMode    = 0700
	tokenBytes = 64 // hex-encoded rawBytes
)

// Generate returns a new high-entropy API token of the form rpsn_<64 hex chars>.
func Generate() (string, error) {
	buf := make([]byte, rawBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("could not generate API token: %w", err)
	}
	return Prefix + hex.EncodeToString(buf), nil
}

// Equal compares two tokens in constant time. Length mismatches still run a
// dummy compare so the timing does not leak the expected length.
func Equal(a, b string) bool {
	ab := []byte(a)
	bb := []byte(b)
	if len(ab) != len(bb) {
		dummy := make([]byte, len(ab))
		subtle.ConstantTimeCompare(ab, dummy)
		return false
	}
	return subtle.ConstantTimeCompare(ab, bb) == 1
}

// Valid reports whether token has the expected generated format.
func Valid(token string) bool {
	if !strings.HasPrefix(token, Prefix) {
		return false
	}
	body := strings.TrimPrefix(token, Prefix)
	if len(body) != tokenBytes {
		return false
	}
	_, err := hex.DecodeString(body)
	return err == nil
}

// ReadFile returns the trimmed token stored at path. Missing or empty files
// return ("", nil).
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("could not read API token file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteFile writes token to path with 0600 permissions, creating the parent
// directory if needed.
func WriteFile(path, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("API token cannot be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return fmt.Errorf("could not create API token directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-api-token-*")
	if err != nil {
		return fmt.Errorf("could not create API token temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(fileMode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not set API token file permissions: %w", err)
	}
	if _, err := tmp.WriteString(token); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not write API token file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("could not close API token file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("could not replace API token file: %w", err)
	}
	if err := os.Chmod(path, fileMode); err != nil {
		return fmt.Errorf("could not set API token file permissions: %w", err)
	}
	return nil
}

// EnsureFile returns the token stored at path, creating a new one if the file
// is missing or empty. Creation uses O_EXCL so concurrent callers converge on
// a single token.
func EnsureFile(path string) (string, error) {
	existing, err := ReadFile(path)
	if err != nil {
		return "", err
	}
	if existing != "" {
		return existing, nil
	}

	token, err := Generate()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return "", fmt.Errorf("could not create API token directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		if os.IsExist(err) {
			existing, readErr := ReadFile(path)
			if readErr != nil {
				return "", readErr
			}
			if existing != "" {
				return existing, nil
			}
			if writeErr := WriteFile(path, token); writeErr != nil {
				return "", writeErr
			}
			return token, nil
		}
		return "", fmt.Errorf("could not create API token file: %w", err)
	}
	if _, err := f.WriteString(token); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("could not write API token file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("could not close API token file: %w", err)
	}
	return token, nil
}
