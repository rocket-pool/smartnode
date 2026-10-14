package migration

import "strconv"

// Preserve the relay catalog and profile membership at the time ePBS settings
// were introduced, independently of future changes to the service defaults.
var epbsRelayMigrations = []struct {
	builderKey     string
	mevProfileKey  string
	commitBoostKey string
}{
	{"flashbotsEnabled", "enableRegulatedAllMev", "cbFlashbotsEnabled"},
	{"bloxRouteRegulatedEnabled", "enableRegulatedAllMev", "cbBloxRouteRegulatedEnabled"},
	{"ultrasoundEnabled", "enableUnregulatedAllMev", ""},
	{"ultrasoundFilteredEnabled", "enableRegulatedAllMev", "cbUltrasoundFilteredEnabled"},
	{"aestusEnabled", "enableUnregulatedAllMev", ""},
	{"titanGlobalEnabled", "enableUnregulatedAllMev", ""},
	{"titanRegionalEnabled", "enableRegulatedAllMev", "cbTitanRegionalEnabled"},
	{"btcsOfacEnabled", "enableRegulatedAllMev", "cbBtcsOfacEnabled"},
}

func upgradeFromV1245(serializedConfig map[string]map[string]string) error {
	// An existing section belongs to the user, including blank bids and disabled
	// builders. Never re-seed it from settings they may have since changed.
	if serializedConfig["epbs"] != nil {
		return nil
	}

	root := serializedConfig["root"]
	mev := serializedConfig["mevBoost"]
	cb := serializedConfig["commitBoostConfig"]

	// Missing settings use the same defaults as the legacy configuration. Match
	// the existing PBS URL precedence if both services were enabled.
	mevEnabled, exists := root["enableMevBoost"]
	useMev := !exists || isTrue(mevEnabled)
	useCb := !useMev && isTrue(root["enableCommitBoost"])
	useMev = useMev && (mev["mode"] == "" || mev["mode"] == "local")
	useCb = useCb && (cb["mode"] == "" || cb["mode"] == "local")

	// There were no shared bid limits to migrate. Default to counting only
	// trustless payments. External service URLs and custom relays remain in their
	// original sections; they are not entries in the built-in builder catalog.
	epbs := map[string]string{"minBid": "", "maxExecutionPayment": "0"}
	for _, relay := range epbsRelayMigrations {
		enabled := false
		if useMev {
			switch mev["selectionMode"] {
			case "", "profile":
				enabled = isTrue(mev[relay.mevProfileKey])
			case "relay":
				enabled = isTrue(mev[relay.builderKey])
			}
		} else if useCb && relay.commitBoostKey != "" {
			switch cb["relaySelectionMode"] {
			case "", "all":
				enabled = true
			case "manual":
				enabled = isTrue(cb[relay.commitBoostKey])
			}
		}
		// Keep selections across networks, just as manual relay selection does.
		// The ePBS config filters unavailable endpoints for the active network.
		epbs[relay.builderKey] = strconv.FormatBool(enabled)
	}
	serializedConfig["epbs"] = epbs
	return nil
}
