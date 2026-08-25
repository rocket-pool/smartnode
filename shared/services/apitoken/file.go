package apitoken

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	fileMode   = 0600
	dirMode    = 0700
	CLIComment = "rocketpool CLI"
)

// Record is one token in the sidecar JSON file.
type Record struct {
	Token   Token  `json:"token"`
	Comment string `json:"comment"`
	Scope   Scope  `json:"scope"`
}

// File is the on-disk token list.
type File struct {
	Tokens []Record `json:"tokens"`
}

// CLIIndex returns the index of the write-scoped CLI token, or -1.
func (f File) CLIIndex() int {
	for i, rec := range f.Tokens {
		if rec.Scope == ScopeWrite {
			return i
		}
	}
	return -1
}

// Lookup returns the record whose token matches bearer, if any.
func (f File) Lookup(bearer string) (Record, bool) {
	tok, err := Parse(bearer)
	if err != nil {
		return Record{}, false
	}
	for _, rec := range f.Tokens {
		if rec.Token.Equal(tok) {
			return rec, true
		}
	}
	return Record{}, false
}

// ReadFile loads tokens from path. A missing file is an empty list.
func ReadFile(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("could not read API token file: %w (if Docker created this file as root, run: sudo chown \"$USER\" %s)", err, path)
	}
	if len(data) == 0 {
		return File{}, nil
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, fmt.Errorf("could not parse API token file: %w", err)
	}
	return f, nil
}

// WriteFile writes f to path with 0600 permissions.
func WriteFile(path string, f File) error {
	if len(f.Tokens) == 0 {
		return fmt.Errorf("API token file cannot be empty")
	}
	for i, rec := range f.Tokens {
		if !rec.Token.Valid() {
			return fmt.Errorf("API token %d is empty", i)
		}
		if rec.Scope != ScopeRead && rec.Scope != ScopeWrite {
			return fmt.Errorf("API token %d has invalid scope", i)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return fmt.Errorf("could not create API token directory: %w", err)
	}

	payload, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode API token file: %w", err)
	}
	payload = append(payload, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-api-tokens-*")
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
	if _, err := tmp.Write(payload); err != nil {
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
	if err := alignOwnership(path); err != nil {
		return fmt.Errorf("could not set API token file ownership: %w", err)
	}
	return nil
}

// EnsureFile returns the token file at path, creating a write-scoped CLI token
// if the file is missing or has no tokens. Creation uses O_EXCL so concurrent
// callers converge on a single file.
func EnsureFile(path string) (File, error) {
	existing, err := ReadFile(path)
	if err != nil {
		return File{}, err
	}
	if len(existing.Tokens) > 0 {
		_ = alignOwnership(path)
		return existing, nil
	}

	tok, err := Generate()
	if err != nil {
		return File{}, err
	}
	created := File{Tokens: []Record{{
		Token:   tok,
		Comment: CLIComment,
		Scope:   ScopeWrite,
	}}}

	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return File{}, fmt.Errorf("could not create API token directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		if os.IsExist(err) {
			existing, readErr := ReadFile(path)
			if readErr != nil {
				return File{}, readErr
			}
			if len(existing.Tokens) > 0 {
				return existing, nil
			}
			if writeErr := WriteFile(path, created); writeErr != nil {
				return File{}, writeErr
			}
			return created, nil
		}
		return File{}, fmt.Errorf("could not create API token file: %w", err)
	}
	_ = f.Close()
	if err := WriteFile(path, created); err != nil {
		return File{}, err
	}
	return created, nil
}

func alignOwnership(path string) error {
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	if err := os.Chown(path, int(stat.Uid), int(stat.Gid)); err != nil {
		return err
	}
	return os.Chmod(path, fileMode)
}
