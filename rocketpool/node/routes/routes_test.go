package routes

import (
	"path"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

func TestRouteClassification(t *testing.T) {
	r := snroute.NewRouter(nil, nil, true)
	RegisterRoutes(r)

	byPath := map[string]snroute.Route{}
	var writes int
	var opens int
	for _, rt := range r.Routes() {
		if _, dup := byPath[rt.Path()]; dup {
			t.Errorf("duplicate route %s", rt.Path())
		}
		byPath[rt.Path()] = rt
		switch rt.(type) {
		case snroute.WriteRoute:
			writes++
			name := path.Base(rt.Path())
			for _, prefix := range []string{"can-", "get-", "is-", "estimate-", "check-"} {
				if strings.HasPrefix(name, prefix) {
					t.Errorf("write route %s starts with %s", rt.Path(), prefix)
				}
			}
		case snroute.OpenRoute:
			opens++
		case snroute.ReadRoute:
		default:
			t.Errorf("untyped route %s %T", rt.Path(), rt)
		}
	}

	if _, ok := byPath["/healthz"].(snroute.OpenRoute); !ok {
		t.Fatal("/healthz must be Open")
	}
	if opens != 1 {
		t.Errorf("expected 1 Open route, got %d", opens)
	}

	safe := []string{
		"/api/version",
		"/api/wait",
		"/api/node/status",
		"/api/node/sync",
		"/api/node/alerts",
		"/api/node/rewards",
		"/api/node/can-send",
		"/api/node/get-eth-balance",
		"/api/node/check-collateral",
		"/api/wallet/status",
		"/api/wallet/recovery-status",
		"/api/minipool/status",
		"/api/minipool/can-exit",
		"/api/network/stats",
		"/api/odao/get-member-settings",
		"/api/service/get-client-status",
		"/api/service/get-gas-price-from-latest-block",
		"/api/wallet/estimate-gas-set-ens-name",
		"/api/pdao/estimate-set-voting-delegate-gas",
		"/api/megapool/can-exit-validator",
	}
	for _, p := range safe {
		rt, ok := byPath[p]
		if !ok {
			t.Errorf("missing route %s", p)
			continue
		}
		if _, isWrite := rt.(snroute.WriteRoute); isWrite {
			t.Errorf("%s should not be Write", p)
		}
	}

	sensitive := []string{
		"/api/node/send",
		"/api/node/send-all",
		"/api/node/deposit",
		"/api/node/withdraw-eth",
		"/api/node/withdraw-rpl",
		"/api/node/stake-rpl",
		"/api/node/sign",
		"/api/wallet/export",
		"/api/wallet/init",
		"/api/wallet/recover",
		"/api/wallet/set-password",
		"/api/minipool/exit",
		"/api/minipool/dissolve",
		"/api/minipool/import-key",
		"/api/megapool/exit-validator",
		"/api/megapool/dissolve-validator",
		"/api/pdao/vote-proposal",
		"/api/pdao/execute-proposal",
		"/api/odao/join",
		"/api/service/restart-vc",
		"/api/service/terminate-data-folder",
		"/api/network/generate-rewards-tree",
		"/api/auction/bid-lot",
		"/api/queue/process",
		"/api/node/register",
		"/api/node/set-primary-withdrawal-address",
		"/api/wallet/set-ens-name",
		"/api/minipool/close",
		"/api/minipool/rescue-dissolved",
		"/api/megapool/notify-validator-exit",
		"/api/security/execute-proposal",
		"/api/upgrade/execute-upgrade",
	}
	for _, p := range sensitive {
		rt, ok := byPath[p]
		if !ok {
			t.Errorf("missing route %s", p)
			continue
		}
		if _, isWrite := rt.(snroute.WriteRoute); !isWrite {
			t.Errorf("%s should be Write, got %T", p, rt)
		}
	}

	if writes == 0 {
		t.Fatal("expected write routes")
	}
}
