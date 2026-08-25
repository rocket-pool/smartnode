package snroute

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestRequiresToken(t *testing.T) {
	ok := func(ctx Context) { ctx.Writer.WriteHeader(http.StatusOK) }
	writeOK := func(ctx WriteContext) { ctx.Writer.WriteHeader(http.StatusOK) }

	if Open("/healthz", ok).RequiresToken(false) || Open("/healthz", ok).RequiresToken(true) {
		t.Fatal("Open must never require a token")
	}
	if !Read("/api/version", ok).RequiresToken(false) {
		t.Fatal("Read must require a token when scope is all endpoints")
	}
	if Read("/api/version", ok).RequiresToken(true) {
		t.Fatal("Read must not require a token when scope is sensitive only")
	}
	if !Write("/api/node/send", writeOK).RequiresToken(false) || !Write("/api/node/send", writeOK).RequiresToken(true) {
		t.Fatal("Write must always require a token")
	}
}

func TestRouterAuth(t *testing.T) {
	const token = "rpsn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ok := func(ctx Context) { ctx.Writer.WriteHeader(http.StatusOK) }
	writeOK := func(ctx WriteContext) { ctx.Writer.WriteHeader(http.StatusOK) }

	t.Run("open without token", func(t *testing.T) {
		r := NewRouter(nil, token, false)
		r.Handle(Open("/healthz", ok))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("read without token when all endpoints", func(t *testing.T) {
		r := NewRouter(nil, token, false)
		r.Handle(Read("/api/version", ok))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/version", nil))
		assertUnauthorized(t, rec)
	})

	t.Run("read without token when sensitive only", func(t *testing.T) {
		r := NewRouter(nil, token, true)
		r.Handle(Read("/api/node/status", ok))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/node/status", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("write without token when sensitive only", func(t *testing.T) {
		r := NewRouter(nil, token, true)
		r.Handle(Write("/api/node/send", writeOK))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		r.ServeHTTP(rec, req)
		assertUnauthorized(t, rec)
	})

	t.Run("write with token", func(t *testing.T) {
		r := NewRouter(nil, token, true)
		r.Handle(Write("/api/node/send", writeOK))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})
}

func TestRegisterTo(t *testing.T) {
	ok := func(ctx Context) { ctx.Writer.WriteHeader(http.StatusOK) }
	writeOK := func(ctx WriteContext) { ctx.Writer.WriteHeader(http.StatusOK) }
	r := NewRouter(nil, "token", true)
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
	assertUnauthorized(t, rec)
}

func assertUnauthorized(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 body %s", rec.Code, rec.Body.String())
	}
	var body api.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "unauthorized" {
		t.Fatalf("error %q, want unauthorized", body.Error)
	}
}
