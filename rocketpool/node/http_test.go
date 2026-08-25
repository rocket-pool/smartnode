package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func TestAuthMiddleware(t *testing.T) {
	const token = "rpsn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ok := func(ctx snroute.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
	}
	writeOK := func(ctx snroute.WriteContext) {
		ctx.Writer.WriteHeader(http.StatusOK)
	}

	register := func(sensitiveOnly bool) *snroute.Router {
		r := snroute.NewRouter(nil, token, sensitiveOnly)
		r.Handle(snroute.Open("/healthz", ok))
		r.Handle(snroute.Read("/api/version", func(ctx snroute.Context) {
			ctx.Writer.Header().Set("Content-Type", "application/json")
			_, _ = ctx.Writer.Write([]byte(`{"status":"success"}`))
		}))
		r.Handle(snroute.Read("/api/node/status", ok))
		r.Handle(snroute.Write("/api/node/send", writeOK))
		r.Handle(snroute.Write("/api/wallet/export", writeOK))
		r.Handle(snroute.Write("/api/megapool/exit-validator", writeOK))
		return r
	}

	handler := register(false)

	t.Run("healthz without token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("missing header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("wrong scheme", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		req.Header.Set("Authorization", "Basic "+token)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("wrong token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		req.Header.Set("Authorization", "Bearer not-the-token")
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("query token ignored", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version?token="+token, nil)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("matching token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("loopback still requires token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	sensitiveOnly := register(true)

	t.Run("sensitive-only status without token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/node/status", nil)
		sensitiveOnly.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("sensitive-only send without token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		sensitiveOnly.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("sensitive-only exit-validator without token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/megapool/exit-validator?validatorId=0", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		sensitiveOnly.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("sensitive-only wallet export without token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/wallet/export", nil)
		sensitiveOnly.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusUnauthorized, "unauthorized")
	})

	t.Run("sensitive-only send with token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/node/send", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		sensitiveOnly.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("unlimited when zero", func(t *testing.T) {
		handler := rateLimitMiddleware(newTokenBucket(0), okHandler)
		for i := 0; i < 20; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("request %d: status %d", i, rec.Code)
			}
		}
	})

	t.Run("rejects after burst", func(t *testing.T) {
		handler := rateLimitMiddleware(newTokenBucket(5), okHandler)
		var saw429 bool
		for i := 0; i < 10; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
			handler.ServeHTTP(rec, req)
			if rec.Code == http.StatusTooManyRequests {
				saw429 = true
				assertAPIError(t, rec, http.StatusTooManyRequests, "too many requests")
				break
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("request %d: status %d", i, rec.Code)
			}
		}
		if !saw429 {
			t.Fatal("expected a 429 after exceeding 5 req/s burst")
		}
	})
}

func TestAPIListenHost(t *testing.T) {
	cfgNative := &config.RocketPoolConfig{IsNativeMode: true}
	_, ok := apiListenHost(cfgNative, cfgtypes.RPC_Closed)
	if ok {
		t.Fatal("native closed should not listen")
	}
	host, ok := apiListenHost(cfgNative, cfgtypes.RPC_OpenLocalhost)
	if !ok || host != "127.0.0.1" {
		t.Fatalf("native localhost: %s %v", host, ok)
	}
	host, ok = apiListenHost(cfgNative, cfgtypes.RPC_OpenExternal)
	if !ok || host != "0.0.0.0" {
		t.Fatalf("native external: %s %v", host, ok)
	}

	cfgDocker := &config.RocketPoolConfig{IsNativeMode: false}
	host, ok = apiListenHost(cfgDocker, cfgtypes.RPC_Closed)
	if !ok || host != "0.0.0.0" {
		t.Fatalf("docker closed still binds in-container: %s %v", host, ok)
	}
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
