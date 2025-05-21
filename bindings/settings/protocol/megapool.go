package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Config
const (
	MegapoolSettingsContractName string = "rocketDAOProtocolSettingsMegapool"
)

// Megapool time before dissolve
func GetMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	megapoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := megapoolSettingsContract.Call(opts, value, "getTimeBeforeDissolve"); err != nil {
		return 0, fmt.Errorf("error getting megapool time before dissolve value: %w", err)
	}
	return (*value).Uint64(), nil
}

// Get contracts
var megapoolSettingsContractLock sync.Mutex

func getMegapoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	megapoolSettingsContractLock.Lock()
	defer megapoolSettingsContractLock.Unlock()
	return rp.GetContract(MegapoolSettingsContractName, opts)
}
