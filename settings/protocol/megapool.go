package protocol

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Config
const (
	MegapoolSettingsContractName string = "rocketDAOProtocolSettingsMegapool"
)

// Megapool time before dissolved
func GetMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(uint64)
	if err := minipoolSettingsContract.Call(opts, value, "getSubmitWithdrawableEnabled"); err != nil {
		return 0, fmt.Errorf("error getting minipool withdrawable submissions enabled status: %w", err)
	}
	return *value, nil
}

// Get contracts
var megapoolSettingsContractLock sync.Mutex

func getMegapoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	megapoolSettingsContractLock.Lock()
	defer megapoolSettingsContractLock.Unlock()
	return rp.GetContract(MegapoolSettingsContractName, opts)
}
