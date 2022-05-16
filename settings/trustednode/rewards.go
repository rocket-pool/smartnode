package trustednode

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Config
const (
	RewardsSettingsContractName string = "rocketDAONodeTrustedSettingsRewards"
	NetworkEnabledPath          string = "rewards.network.enabled"
)

// Get whether or not the provided rewards network is enabled
func GetNetworkEnabled(rp *rocketpool.RocketPool, network *big.Int, opts *bind.CallOpts) (bool, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rewardsSettingsContract.Call(opts, value, "getNetworkEnabled", network); err != nil {
		return false, fmt.Errorf("Could not check if network %s is enabled: %w", network.String(), err)
	}
	return (*value), nil
}

// Get contracts
var rewardsSettingsContractLock sync.Mutex

func getRewardsSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rewardsSettingsContractLock.Lock()
	defer rewardsSettingsContractLock.Unlock()
	return rp.GetContract(RewardsSettingsContractName)
}
