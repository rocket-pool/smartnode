package node

import "strings"

// submitsTransaction is every HTTP path that submits an execution-layer
// transaction or a consensus-layer voluntary exit. Checked first so a tx
// route cannot be treated as read-only by prefix rules.
var submitsTransaction = map[string]struct{}{
	// node
	"/api/node/register":                           {},
	"/api/node/set-timezone":                       {},
	"/api/node/set-primary-withdrawal-address":     {},
	"/api/node/confirm-primary-withdrawal-address": {},
	"/api/node/set-rpl-withdrawal-address":         {},
	"/api/node/confirm-rpl-withdrawal-address":     {},
	"/api/node/swap-rpl-approve-rpl":               {},
	"/api/node/wait-and-swap-rpl":                  {},
	"/api/node/swap-rpl":                           {},
	"/api/node/stake-rpl-approve-rpl":              {},
	"/api/node/wait-and-stake-rpl":                 {},
	"/api/node/stake-rpl":                          {},
	"/api/node/set-rpl-locking-allowed":            {},
	"/api/node/set-stake-rpl-for-allowed":          {},
	"/api/node/withdraw-rpl":                       {},
	"/api/node/unstake-legacy-rpl":                 {},
	"/api/node/withdraw-rpl-v131":                  {},
	"/api/node/unstake-rpl":                        {},
	"/api/node/withdraw-eth":                       {},
	"/api/node/withdraw-credit":                    {},
	"/api/node/deposit":                            {},
	"/api/node/send":                               {},
	"/api/node/send-all":                           {},
	"/api/node/burn":                               {},
	"/api/node/claim-rpl-rewards":                  {},
	"/api/node/initialize-fee-distributor":         {},
	"/api/node/distribute":                         {},
	"/api/node/claim-rewards":                      {},
	"/api/node/claim-and-stake-rewards":            {},
	"/api/node/set-smoothing-pool-status":          {},
	"/api/node/create-vacant-minipool":             {},
	"/api/node/send-message":                       {},
	"/api/node/provision-express-tickets":          {},
	"/api/node/claim-unclaimed-rewards":            {},
	"/api/wallet/set-ens-name":                     {},
	// minipool
	"/api/minipool/refund":                  {},
	"/api/minipool/stake":                   {},
	"/api/minipool/promote":                 {},
	"/api/minipool/dissolve":                {},
	"/api/minipool/exit":                    {},
	"/api/minipool/close":                   {},
	"/api/minipool/delegate-upgrade":        {},
	"/api/minipool/set-use-latest-delegate": {},
	"/api/minipool/distribute-balance":      {},
	"/api/minipool/change-withdrawal-creds": {},
	"/api/minipool/rescue-dissolved":        {},
	// megapool
	"/api/megapool/claim-refund":            {},
	"/api/megapool/repay-debt":              {},
	"/api/megapool/reduce-bond":             {},
	"/api/megapool/stake":                   {},
	"/api/megapool/dissolve-validator":      {},
	"/api/megapool/dissolve-with-proof":     {},
	"/api/megapool/exit-validator":          {},
	"/api/megapool/notify-validator-exit":   {},
	"/api/megapool/notify-final-balance":    {},
	"/api/megapool/exit-queue":              {},
	"/api/megapool/distribute":              {},
	"/api/megapool/delegate-upgrade":        {},
	"/api/megapool/set-use-latest-delegate": {},
	// auction
	"/api/auction/create-lot":  {},
	"/api/auction/bid-lot":     {},
	"/api/auction/claim-lot":   {},
	"/api/auction/recover-lot": {},
	// queue
	"/api/queue/process":         {},
	"/api/queue/assign-deposits": {},
	// pdao
	"/api/pdao/vote-proposal":                              {},
	"/api/pdao/override-vote":                              {},
	"/api/pdao/execute-proposal":                           {},
	"/api/pdao/propose-setting":                            {},
	"/api/pdao/propose-setting-multi":                      {},
	"/api/pdao/propose-rewards-percentages":                {},
	"/api/pdao/propose-one-time-spend":                     {},
	"/api/pdao/propose-recurring-spend":                    {},
	"/api/pdao/propose-recurring-spend-update":             {},
	"/api/pdao/propose-invite-to-security-council":         {},
	"/api/pdao/propose-kick-from-security-council":         {},
	"/api/pdao/propose-kick-multi-from-security-council":   {},
	"/api/pdao/propose-replace-member-of-security-council": {},
	"/api/pdao/claim-bonds":                                {},
	"/api/pdao/defeat-proposal":                            {},
	"/api/pdao/finalize-proposal":                          {},
	"/api/pdao/set-voting-delegate":                        {},
	"/api/pdao/set-signalling-address":                     {},
	"/api/pdao/clear-signalling-address":                   {},
	"/api/pdao/propose-allow-listed-controllers":           {},
	// odao
	"/api/odao/propose-invite":                       {},
	"/api/odao/propose-leave":                        {},
	"/api/odao/propose-kick":                         {},
	"/api/odao/cancel-proposal":                      {},
	"/api/odao/vote-proposal":                        {},
	"/api/odao/execute-proposal":                     {},
	"/api/odao/join-approve-rpl":                     {},
	"/api/odao/join":                                 {},
	"/api/odao/leave":                                {},
	"/api/odao/penalise-megapool":                    {},
	"/api/odao/propose-members-quorum":               {},
	"/api/odao/propose-members-rplbond":              {},
	"/api/odao/propose-proposal-cooldown":            {},
	"/api/odao/propose-proposal-vote-timespan":       {},
	"/api/odao/propose-proposal-vote-delay-timespan": {},
	"/api/odao/propose-proposal-execute-timespan":    {},
	"/api/odao/propose-proposal-action-timespan":     {},
	"/api/odao/propose-scrub-period":                 {},
	"/api/odao/propose-promotion-scrub-period":       {},
	"/api/odao/propose-scrub-penalty-enabled":        {},
	"/api/odao/propose-bond-reduction-window-start":  {},
	"/api/odao/propose-bond-reduction-window-length": {},
	// security council
	"/api/security/propose-leave":    {},
	"/api/security/propose-setting":  {},
	"/api/security/cancel-proposal":  {},
	"/api/security/vote-proposal":    {},
	"/api/security/execute-proposal": {},
	"/api/security/join":             {},
	"/api/security/leave":            {},
	"/api/upgrade/execute-upgrade":   {},
}

// sensitiveLocal is high-impact local state that does not send an EL tx:
// wallet secrets, validator keys, signing, and destructive service actions.
var sensitiveLocal = map[string]struct{}{
	"/api/wallet/set-password":           {},
	"/api/wallet/init":                   {},
	"/api/wallet/recover":                {},
	"/api/wallet/search-and-recover":     {},
	"/api/wallet/rebuild":                {},
	"/api/wallet/export":                 {},
	"/api/wallet/masquerade":             {},
	"/api/wallet/end-masquerade":         {},
	"/api/minipool/import-key":           {},
	"/api/node/sign":                     {},
	"/api/node/sign-message":             {},
	"/api/service/restart-vc":            {},
	"/api/service/terminate-data-folder": {},
	"/api/network/generate-rewards-tree": {},
	"/api/network/download-rewards-file": {},
}

// readOnlyExact are last-path-segment names that only return information.
// Prefix rules in isSensitiveAPIPath cover can-*, get-*, is-*, estimate-*, and check-*.
var readOnlyExact = map[string]struct{}{
	"status":                           {},
	"alerts":                           {},
	"sync":                             {},
	"rewards":                          {},
	"lots":                             {},
	"members":                          {},
	"proposals":                        {},
	"proposal-details":                 {},
	"stats":                            {},
	"timezone-map":                     {},
	"dao-proposals":                    {},
	"node-fee":                         {},
	"rpl-price":                        {},
	"latest-delegate":                  {},
	"rewards-event":                    {},
	"recovery-status":                  {},
	"pending-rewards":                  {},
	"calculate-rewards":                {},
	"latest-block-withdrawals":         {},
	"beacon-withdrawal-queue-estimate": {},
	"validator-map-and-balances":       {},
	"deposit-contract-info":            {},
	"resolve-ens-name":                 {},
	"reverse-resolve-ens-name":         {},
	"test-recover":                     {},
	"test-search-and-recover":          {},
}

// isSensitiveAPIPath reports whether path must have a bearer token when
// Token Requirement is "sensitive endpoints only". Transaction-submitting
// routes and high-impact local operations are always sensitive. Status, gas
// estimates (can-*), and similar reads are not.
func isSensitiveAPIPath(path string) bool {
	path = strings.TrimSuffix(path, "/")
	if _, ok := submitsTransaction[path]; ok {
		return true
	}
	if _, ok := sensitiveLocal[path]; ok {
		return true
	}

	switch path {
	case healthzPath, "/api/version", "/api/wait":
		return false
	}

	i := strings.LastIndex(path, "/")
	name := path
	if i >= 0 {
		name = path[i+1:]
	}

	for _, prefix := range []string{"can-", "get-", "is-", "estimate-", "check-"} {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	if _, ok := readOnlyExact[name]; ok {
		return false
	}
	return true
}
