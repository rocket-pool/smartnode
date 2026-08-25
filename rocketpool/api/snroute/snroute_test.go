package snroute

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/apitoken"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	writeTokenStr = "rpsn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	readTokenStr  = "rpsn_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func testRecords(t *testing.T, write, read bool) []apitoken.Record {
	t.Helper()
	var recs []apitoken.Record
	if write {
		tok, err := apitoken.Parse(writeTokenStr)
		if err != nil {
			t.Fatal(err)
		}
		recs = append(recs, apitoken.Record{Token: tok, Comment: "cli", Scope: apitoken.ScopeWrite})
	}
	if read {
		tok, err := apitoken.Parse(readTokenStr)
		if err != nil {
			t.Fatal(err)
		}
		recs = append(recs, apitoken.Record{Token: tok, Comment: "app", Scope: apitoken.ScopeRead})
	}
	return recs
}

func TestRequiresToken(t *testing.T) {
	ok := func(ctx Context) { ctx.Writer.WriteHeader(http.StatusOK) }
	writeOK := func(ctx WriteContext) { ctx.Writer.WriteHeader(http.StatusOK) }

	if Open("/healthz", ok).RequiresToken(false) || Open("/healthz", ok).RequiresToken(true) {
		t.Fatal("Open must never require a token")
	}
	if !Read("/api/version", ok).RequiresToken(false) {
		t.Fatal("Read must require a token when unauthenticated reads are off")
	}
	if Read("/api/version", ok).RequiresToken(true) {
		t.Fatal("Read must not require a token when unauthenticated reads are on")
	}
	if !Write("/api/node/send", writeOK).RequiresToken(false) || !Write("/api/node/send", writeOK).RequiresToken(true) {
		t.Fatal("Write must always require a token")
	}
}

func TestRouterAuth(t *testing.T) {
	ok := func(ctx Context) { ctx.Writer.WriteHeader(http.StatusOK) }
	writeOK := func(ctx WriteContext) { ctx.Writer.WriteHeader(http.StatusOK) }
	tokens := testRecords(t, true, true)

	t.Run("open without token", func(t *testing.T) {
		r := NewRouter(nil, tokens, false)
		r.Handle(Open("/healthz", ok))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("read without token", func(t *testing.T) {
		r := NewRouter(nil, tokens, false)
		r.Handle(Read("/api/version", ok))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/version", nil))
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("unauthenticated reads without token", func(t *testing.T) {
		r := NewRouter(nil, tokens, true)
		r.Handle(Read("/api/node/status", ok))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/node/status", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("write without token even with unauthenticated reads", func(t *testing.T) {
		r := NewRouter(nil, tokens, true)
		r.Handle(Write("/api/node/send", writeOK))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		r.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("write with write token", func(t *testing.T) {
		r := NewRouter(nil, tokens, true)
		r.Handle(Write("/api/node/send", writeOK))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.Header.Set("Authorization", "Bearer "+writeTokenStr)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("read with read token", func(t *testing.T) {
		r := NewRouter(nil, tokens, false)
		r.Handle(Read("/api/node/status", ok))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/node/status", nil)
		req.Header.Set("Authorization", "Bearer "+readTokenStr)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("write with read token", func(t *testing.T) {
		r := NewRouter(nil, tokens, false)
		r.Handle(Write("/api/node/send", writeOK))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.Header.Set("Authorization", "Bearer "+readTokenStr)
		r.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusForbidden, "forbidden")
	})

	t.Run("wrong token", func(t *testing.T) {
		r := NewRouter(nil, tokens, false)
		r.Handle(Read("/api/version", ok))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		req.Header.Set("Authorization", "Bearer not-the-token")
		r.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})
}

func TestRegisterTo(t *testing.T) {
	ok := func(ctx Context) { ctx.Writer.WriteHeader(http.StatusOK) }
	writeOK := func(ctx WriteContext) { ctx.Writer.WriteHeader(http.StatusOK) }
	r := NewRouter(nil, testRecords(t, true, false), true)
	Open("/healthz", ok).RegisterTo(r)
	Read("/api/version", ok).RegisterTo(r)
	Write("/api/node/send", writeOK).RegisterTo(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("open status %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/version", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("read status %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/node/send", nil))
	assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, code int, msg string) {
	t.Helper()
	if rec.Code != code {
		t.Fatalf("status %d, want %d body %s", rec.Code, code, rec.Body.String())
	}
	var body api.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != msg {
		t.Fatalf("error %q, want %q", body.Error, msg)
	}
}
