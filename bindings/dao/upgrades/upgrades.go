package upgrades

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/strings"
)

// Get the total number of upgrade proposals
func GetTotalUpgradeProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return 0, err
	}
	proposalCount := new(*big.Int)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, proposalCount, "getTotal"); err != nil {
		return 0, fmt.Errorf("error getting upgrade proposal count: %w", err)
	}
	return (*proposalCount).Uint64(), nil
}

func GetUpgradeProposalState(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.CallOpts) (rptypes.UpgradeProposalState, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return 0, err
	}
	state := new(uint8)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, state, "getState", big.NewInt(int64(upgradeProposalId))); err != nil {
		return 0, fmt.Errorf("error getting upgrade proposal %d state: %w", upgradeProposalId, err)
	}
	return rptypes.UpgradeProposalState(*state), nil
}

func GetUpgradeProposalEndTime(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return nil, err
	}
	endTime := new(*big.Int)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, endTime, "getEnd", big.NewInt(int64(upgradeProposalId))); err != nil {
		return nil, fmt.Errorf("error getting upgrade proposal %d end time: %w", upgradeProposalId, err)
	}
	return *endTime, nil
}

func GetUpgradeProposalUpgradeAddress(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.CallOpts) (common.Address, error) {

	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	upgradeAddress := new(common.Address)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, upgradeAddress, "getUpgradeAddress", big.NewInt(int64(upgradeProposalId))); err != nil {
		return common.Address{}, fmt.Errorf("error getting upgrade proposal %d upgrade address: %w", upgradeProposalId, err)
	}
	return *upgradeAddress, nil
}

func GetUpgradeProposalUpgradeAbi(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.CallOpts) (string, error) {

	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return "", err
	}
	upgradeAbi := new(string)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, upgradeAbi, "getUpgradeABI", big.NewInt(int64(upgradeProposalId))); err != nil {
		return "", fmt.Errorf("error getting upgrade proposal %d upgrade abi: %w", upgradeProposalId, err)
	}
	return *upgradeAbi, nil
}

func GetUpgradeProposalName(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.CallOpts) (string, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return "", err
	}
	name := new(string)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, name, "getName", big.NewInt(int64(upgradeProposalId))); err != nil {
		return "", fmt.Errorf("error getting upgrade proposal %d name: %w", upgradeProposalId, err)
	}
	return strings.Sanitize(*name), nil
}

// Estimate the gas of ExecuteUpgrade
func EstimateExecuteUpgradeGas(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedUpgrade.GetTransactionGasInfo(opts, "execute", big.NewInt(int64(upgradeProposalId)))
}

// Execute an upgrade
func ExecuteUpgrade(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedUpgrade.Transact(opts, "execute", big.NewInt(int64(upgradeProposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error executing trusted node DAO upgrade %d: %w", upgradeProposalId, err)
	}
	return tx.Hash(), nil
}

func GetUpgradeProposalType(rp *rocketpool.RocketPool, upgradeProposalId uint64, opts *bind.CallOpts) ([32]byte, error) {
	rocketDAONodeTrustedUpgrade, err := getRocketDAONodeTrustedUpgrade(rp, opts)
	if err != nil {
		return [32]byte{}, err

	}
	proposalType := new([32]byte)
	if err := rocketDAONodeTrustedUpgrade.Call(opts, proposalType, "getType", big.NewInt(int64(upgradeProposalId))); err != nil {
		return [32]byte{}, fmt.Errorf("error getting upgrade proposal %d type: %w", upgradeProposalId, err)
	}
	return *proposalType, nil
}

var rocketDAONodeTrustedUpgradeLock sync.Mutex

func getRocketDAONodeTrustedUpgrade(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAONodeTrustedUpgradeLock.Lock()
	defer rocketDAONodeTrustedUpgradeLock.Unlock()
	return rp.GetContract("rocketDAONodeTrustedUpgrade", opts)
}
