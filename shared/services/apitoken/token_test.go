package apitoken

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseRoundTrip(t *testing.T) {
	tok, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	s := tok.String()
	if len(s) != len(Prefix)+tokenBytes {
		t.Fatalf("encoded length %d", len(s))
	}
	parsed, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	if !tok.Equal(parsed) {
		t.Fatal("round trip mismatch")
	}
	if !tok.Valid() || tok.IsZero() {
		t.Fatal("generated token should be valid")
	}
}

func TestParseRejects(t *testing.T) {
	for _, s := range []string{"", "foo", Prefix, Prefix + "zz", Prefix + "aa"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) succeeded", s)
		}
	}
}

func TestJSONFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "api-tokens.json")
	tok, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	in := File{Tokens: []Record{{
		Token:   tok,
		Comment: "external",
		Scope:   ScopeRead,
	}}}
	if err := WriteFile(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Tokens) != 1 {
		t.Fatalf("len %d", len(out.Tokens))
	}
	if !out.Tokens[0].Token.Equal(tok) || out.Tokens[0].Comment != "external" || out.Tokens[0].Scope != ScopeRead {
		t.Fatalf("got %+v", out.Tokens[0])
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureFileCreatesCLIWriteToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api-tokens.json")
	f, err := EnsureFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if idx := f.CLIIndex(); idx != 0 {
		t.Fatalf("cli index %d", idx)
	}
	if f.Tokens[0].Scope != ScopeWrite || f.Tokens[0].Comment != CLIComment {
		t.Fatalf("got %+v", f.Tokens[0])
	}
	again, err := EnsureFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !again.Tokens[0].Token.Equal(f.Tokens[0].Token) {
		t.Fatal("EnsureFile must not rotate an existing token")
	}
}

func TestLookup(t *testing.T) {
	writeTok, _ := Generate()
	readTok, _ := Generate()
	f := File{Tokens: []Record{
		{Token: writeTok, Comment: CLIComment, Scope: ScopeWrite},
		{Token: readTok, Comment: "app", Scope: ScopeRead},
	}}
	if rec, ok := f.Lookup(writeTok.String()); !ok || rec.Scope != ScopeWrite {
		t.Fatal("write lookup")
	}
	if rec, ok := f.Lookup(readTok.String()); !ok || rec.Scope != ScopeRead {
		t.Fatal("read lookup")
	}
	if _, ok := f.Lookup("rpsn_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"); ok {
		t.Fatal("unknown token")
	}
}
