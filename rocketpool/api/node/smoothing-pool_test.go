package node

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services/apitoken"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

func TestSmoothingPoolMasqueradeLeavesFeeRecipientUnchanged(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.NewRocketPoolConfig(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Smartnode.DataPath.Value = dir
	if err := cfg.Save(dir, "settings.yml"); err != nil {
		t.Fatal(err)
	}
	path := cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	original := "FEE_RECIPIENT=0x1111111111111111111111111111111111111111"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	token, err := apitoken.Generate()
	if err != nil {
		t.Fatal(err)
	}
	command := &cli.Command{
		Flags: []cli.Flag{&cli.StringFlag{Name: "settings", Value: filepath.Join(dir, "settings.yml")}},
		Action: func(_ context.Context, c *cli.Command) error {
			router := snroute.NewRouter(c, []apitoken.Record{{Token: token, Scope: apitoken.ScopeWrite}}, false)
			RegisterRoutes(router)
			for _, observe := range []bool{false, true} {
				am := wallet.NewAddressManager(cfg.Smartnode.GetNodeAddressPath())
				if err := am.SetAndSaveAddress(common.HexToAddress("0x2222222222222222222222222222222222222222"), observe); err != nil {
					t.Fatal(err)
				}
				for _, status := range []string{"true", "false"} {
					req := httptest.NewRequest(http.MethodPost, "/api/node/set-smoothing-pool-status", strings.NewReader("status="+status))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					req.Header.Set("Authorization", "Bearer "+token.String())
					rec := httptest.NewRecorder()
					router.ServeHTTP(rec, req)
					var response struct{ Error string }
					if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
						t.Fatal(err)
					}
					if response.Error != wallet.ErrIsMasquerading.Error() {
						t.Fatalf("observe=%v status=%s: expected rejection before accessing clients, got %s", observe, status, rec.Body.String())
					}
					contents, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if string(contents) != original {
						t.Fatalf("observe=%v status=%s: fee recipient changed", observe, status)
					}
				}
			}
			return nil
		},
	}
	if err := command.Run(context.Background(), []string{"test"}); err != nil {
		t.Fatal(err)
	}
}
