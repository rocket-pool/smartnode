package rp

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/shared/config"
)

type DepositContractInfo struct {
	RPNetwork             uint64
	RPDepositContract     common.Address
	BeaconNetwork         uint64
	BeaconDepositContract common.Address
}

func GetDepositContractInfo(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*DepositContractInfo, error) {
	info := &DepositContractInfo{}
	info.RPNetwork = uint64(cfg.Smartnode.GetChainID())

	// Get the deposit contract address Rocket Pool will deposit to
	rpDepositContract, err := rp.GetContract(rocketpool.ContractName_CasperDeposit)
	if err != nil {
		return nil, fmt.Errorf("Error getting Casper deposit contract: %w", err)
	}
	if rpDepositContract == nil {
		return nil, fmt.Errorf("Deposit contract was undefined.")
	}
	info.RPDepositContract = *rpDepositContract.Address

	// Get the Beacon Client info
	eth2DepositContract, err := bc.GetEth2DepositContract()
	if err != nil {
		return nil, fmt.Errorf("Error getting beacon client deposit contract: %w", err)
	}

	info.BeaconNetwork = eth2DepositContract.ChainID
	info.BeaconDepositContract = eth2DepositContract.Address

	return info, nil
}
