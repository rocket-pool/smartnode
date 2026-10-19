package migration

import "strconv"

// Preserve overlapping MEV-Boost / Commit-Boost operator selections at the time
// ePBS settings were introduced.
var epbsBuilderMigrations = []struct {
	builderKey      string
	mevRelayKeys    []string
	mevProfileKeys  []string
	commitBoostKeys []string
}{
	{"aestusEnabled", []string{"aestusEnabled"}, []string{"enableUnregulatedAllMev"}, nil},
	{"flashbotsEnabled", []string{"flashbotsEnabled"}, []string{"enableRegulatedAllMev"}, []string{"cbFlashbotsEnabled"}},
	{"nflaigEnabled", nil, nil, nil},
	{"titanEnabled", []string{"titanGlobalEnabled", "titanRegionalEnabled"}, []string{"enableRegulatedAllMev", "enableUnregulatedAllMev"}, []string{"cbTitanRegionalEnabled"}},
	{"ultrasoundEnabled", []string{"ultrasoundEnabled"}, []string{"enableUnregulatedAllMev"}, nil},
	{"pk910Enabled", nil, nil, nil},
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
	for _, builder := range epbsBuilderMigrations {
		enabled := false
		if useMev {
			switch mev["selectionMode"] {
			case "", "profile":
				for _, key := range builder.mevProfileKeys {
					enabled = enabled || isTrue(mev[key])
				}
			case "relay":
				for _, key := range builder.mevRelayKeys {
					enabled = enabled || isTrue(mev[key])
				}
			}
		} else if useCb && len(builder.commitBoostKeys) > 0 {
			switch cb["relaySelectionMode"] {
			case "", "all":
				enabled = true
			case "manual":
				for _, key := range builder.commitBoostKeys {
					enabled = enabled || isTrue(cb[key])
				}
			}
		}
		// Keep selections across networks. The ePBS config only exposes Hoodi
		// endpoints for the active network.
		epbs[builder.builderKey] = strconv.FormatBool(enabled)
	}
	serializedConfig["epbs"] = epbs
	return nil
}
