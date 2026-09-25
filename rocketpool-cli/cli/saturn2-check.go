package cli

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

const Saturn2NotDeployedMessage = "This command is not available until Saturn 2 is deployed."

// IsSaturn2OnlySetting reports whether a protocol setting exists only after Saturn 2.
func IsSaturn2OnlySetting(contract, setting string) bool {
	switch contract {
	case protocol.PerformanceSettingsContractName:
		return true
	case protocol.NetworkSettingsContractName:
		switch setting {
		case protocol.CooperativeExitPhaseSettingPath, protocol.DidNotExitPenaltyBaseSettingPath,
			protocol.DidNotExitBaseSettingPath, protocol.DidNotExitBackoffSettingPath:
			return true
		}
		return false
	case protocol.MegapoolSettingsContractName:
		return setting == protocol.MegapoolPrestakeChallengePeriodPath
	default:
		return false
	}
}

// RequireSaturn2 prints Saturn2NotDeployedMessage and returns false when Saturn 2 is not deployed
func RequireSaturn2(rp *rocketpool.Client) (ok bool, err error) {
	settings, err := rp.PDAOGetSettings()
	if err != nil {
		return false, fmt.Errorf("error checking if Saturn 2 is deployed: %w", err)
	}
	if !settings.Saturn2Deployed {
		fmt.Println(Saturn2NotDeployedMessage)
		return false, nil
	}
	return true, nil
}
