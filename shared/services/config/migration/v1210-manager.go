package migration

import "strconv"

const (
	mevMaxProfitKey       = "bloxRouteMaxProfitEnabled"
	mevRegulatedKey       = "bloxRouteRegulatedEnabled"
	mevFlashbotsKey       = "flashbotsEnabled"
	mevSelectionModeKey   = "selectionMode"
	mevSelectionModeRelay = "relay"

	cbMaxProfitKey        = "cbBloxRouteMaxProfitEnabled"
	cbRegulatedKey        = "cbBloxRouteRegulatedEnabled"
	cbFlashbotsKey        = "cbFlashbotsEnabled"
	cbSelectionModeKey    = "relaySelectionMode"
	cbSelectionModeManual = "manual"
	cbCustomRelaysKey     = "customRelays"
)

var remainingMevRelayKeys = []string{
	mevFlashbotsKey,
	mevRegulatedKey,
	"ultrasoundEnabled",
	"ultrasoundFilteredEnabled",
	"aestusEnabled",
	"titanGlobalEnabled",
	"titanRegionalEnabled",
	"btcsOfacEnabled",
}

var remainingCbRelayKeys = []string{
	cbFlashbotsKey,
	cbRegulatedKey,
	"cbTitanRegionalEnabled",
	"cbUltrasoundFilteredEnabled",
	"cbBtcsOfacEnabled",
}

func upgradeFromV1210(serializedConfig map[string]map[string]string) error {
	network := ""
	if sn, ok := serializedConfig["smartnode"]; ok {
		network = sn["network"]
	}

	migrateMevBoostMaxProfit(serializedConfig["mevBoost"], network)
	migrateCommitBoostMaxProfit(serializedConfig["commitBoostConfig"], network)
	migrateApiPort(serializedConfig)
	return nil
}

func migrateApiPort(serializedConfig map[string]map[string]string) {
	smartnode, exists := serializedConfig["smartnode"]
	if !exists {
		return
	}

	port, exists := smartnode["apiPort"]
	if !exists || port == "" {
		return
	}

	api, exists := serializedConfig["api"]
	if !exists {
		api = map[string]string{}
		serializedConfig["api"] = api
	}
	if api["apiPort"] == "" {
		api["apiPort"] = port
	}
	delete(smartnode, "apiPort")
}

func migrateMevBoostMaxProfit(section map[string]string, network string) {
	if section == nil || !isTrue(section[mevMaxProfitKey]) {
		return
	}

	if section[mevSelectionModeKey] == mevSelectionModeRelay {
		if isMainnet(network) {
			section[mevRegulatedKey] = "true"
		} else if !anyTrue(section, remainingMevRelayKeys) {
			section[mevFlashbotsKey] = "true"
		}
	}

	delete(section, mevMaxProfitKey)
}

func migrateCommitBoostMaxProfit(section map[string]string, network string) {
	if section == nil || !isTrue(section[cbMaxProfitKey]) {
		return
	}

	if section[cbSelectionModeKey] == cbSelectionModeManual {
		if isMainnet(network) {
			section[cbRegulatedKey] = "true"
		} else if !anyTrue(section, remainingCbRelayKeys) && section[cbCustomRelaysKey] == "" {
			section[cbFlashbotsKey] = "true"
		}
	}

	delete(section, cbMaxProfitKey)
}

func isMainnet(network string) bool {
	return network == "" || network == "mainnet"
}

func isTrue(value string) bool {
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

func anyTrue(section map[string]string, keys []string) bool {
	for _, key := range keys {
		if isTrue(section[key]) {
			return true
		}
	}
	return false
}
