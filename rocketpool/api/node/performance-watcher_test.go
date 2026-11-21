package node

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services/apitoken"
	"github.com/urfave/cli/v3"
)

func TestWatcherReadRoutesAndValidation(t *testing.T) {
	token, err := apitoken.Generate()
	if err != nil {
		t.Fatal(err)
	}
	command := &cli.Command{Action: func(_ context.Context, c *cli.Command) error {
		router := snroute.NewRouter(c, []apitoken.Record{{Token: token, Scope: apitoken.ScopeRead}}, false)
		RegisterRoutes(router)
		for _, tc := range []struct {
			path string
			code int
			text string
		}{
			{"/api/node/performance-watcher/candidates", 200, "disabled or unavailable"},
			{"/api/node/performance-watcher/candidates?limit=0", 400, "limit must be positive"},
			{"/api/node/performance-watcher/candidates?offset=-1", 400, "invalid pagination"},
			{"/api/node/performance-watcher/candidate?validatorIndex=foo", 400, "invalid validator index"},
			{"/api/node/performance-watcher/candidate?validatorIndex=1", 500, "disabled or unavailable"},
			{"/api/node/performance-watcher/validator", 400, "invalid validator index"},
			{"/api/node/performance-watcher/validator?validatorIndex=foo", 400, "invalid validator index"},
			{"/api/node/performance-watcher/validator?validatorIndex=1", 500, "disabled or unavailable"},
			{"/api/node/performance-watcher/validators", 500, "disabled or unavailable"},
		} {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token.String())
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.code || !strings.Contains(rec.Body.String(), tc.text) {
				t.Fatalf("%s: %d %s", tc.path, rec.Code, rec.Body.String())
			}
		}
		for _, path := range []string{"/api/node/performance-watcher/candidates", "/api/node/performance-watcher/validator?validatorIndex=1", "/api/node/performance-watcher/validators"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatal("unauthenticated watcher access")
			}
		}
		return nil
	}}
	if err = command.Run(context.Background(), []string{"test"}); err != nil {
		t.Fatal(err)
	}
}
