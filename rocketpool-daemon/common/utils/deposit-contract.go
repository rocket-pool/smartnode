package utils

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

type DepositContractInfo struct {
	RPNetwork             uint64
	RPDepositContract     common.Address
	BeaconNetwork         uint64
	BeaconDepositContract common.Address
}

func GetDepositContractInfo(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient) (*DepositContractInfo, error) {
	info := &DepositContractInfo{}
	resources := cfg.GetRocketPoolResources()
	info.RPNetwork = uint64(resources.ChainID)

	// Get the deposit contract address Rocket Pool will deposit to
	rpDepositContract, err := rp.GetContract(rocketpool.ContractName_CasperDeposit)
	if err != nil {
		return nil, fmt.Errorf("error getting Casper deposit contract: %w", err)
	}
	if rpDepositContract == nil {
		return nil, fmt.Errorf("deposit contract was undefined")
	}
	info.RPDepositContract = rpDepositContract.Address

	// Get the Beacon Client info
	eth2DepositContract, err := bc.GetEth2DepositContract(context)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon client deposit contract: %w", err)
	}

	info.BeaconNetwork = eth2DepositContract.ChainID
	info.BeaconDepositContract = eth2DepositContract.Address

	return info, nil
}
