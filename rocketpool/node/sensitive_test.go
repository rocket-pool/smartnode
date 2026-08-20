package node

import "testing"

func TestIsSensitiveAPIPath(t *testing.T) {
	safe := []string{
		"/healthz",
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
		"/api/minipool/can-exit",
		"/api/megapool/can-exit-validator",
	}
	for _, path := range safe {
		if isSensitiveAPIPath(path) {
			t.Errorf("%s should not be sensitive", path)
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
	for _, path := range sensitive {
		if !isSensitiveAPIPath(path) {
			t.Errorf("%s should be sensitive", path)
		}
	}
}
