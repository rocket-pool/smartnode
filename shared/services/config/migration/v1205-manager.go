package migration

import (
	"math/rand/v2"
	"strings"
)

var checkpointSyncUrlReplacements = []string{
	"https://beaconstate.ethstaker.cc",
	"https://sync-mainnet.beaconcha.in",
	"https://beaconstate-mainnet.chainsafe.io",
	"https://mainnet.checkpoint.sigp.io",
	"https://mainnet-checkpoint-sync.attestant.io",
}

var testnetCheckpointSyncUrlReplacements = []string{
	"https://beaconstate-hoodi.chainsafe.io",
	"https://hoodi.beaconstate.ethstaker.cc/",
	"https://hoodi-checkpoint-sync.stakely.io",
	"https://checkpoint-sync.hoodi.ethpandaops.io",
	"https://hoodi-checkpoint-sync.attestant.io",
}

func upgradeFromV1205(serializedConfig map[string]map[string]string) error {

	// beaconstate.info has been sunset, so migrate users to a randomly selected alternative
	checkpointSyncUrl, exists := serializedConfig["consensusCommon"]["checkpointSyncUrl"]
	if exists && strings.Contains(checkpointSyncUrl, "https://beaconstate.info") {
		serializedConfig["consensusCommon"]["checkpointSyncUrl"] = checkpointSyncUrlReplacements[rand.IntN(len(checkpointSyncUrlReplacements))]
	}
	if exists && strings.Contains(checkpointSyncUrl, "https://hoodi.beaconstate.info") {
		serializedConfig["consensusCommon"]["checkpointSyncUrl"] = testnetCheckpointSyncUrlReplacements[rand.IntN(len(testnetCheckpointSyncUrlReplacements))]
	}
	return nil
}
