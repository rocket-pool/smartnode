package security

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas of ProposeSetUint
func EstimateProposeSetUintGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOSecurityProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set uint setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to update a uint trusted node DAO setting
func ProposeSetUint(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOSecurityProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set uint setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeSetBool
func EstimateProposeSetBoolGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOSecurityProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set bool setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to update a bool trusted node DAO setting
func ProposeSetBool(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOSecurityProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set bool setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of a proposal submission
func EstimateProposalGas(rp *rocketpool.RocketPool, message string, payload []byte, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityProposals.GetTransactionGasInfo(opts, "propose", message, payload)
}

// Submit a security DAO proposal
// Returns the ID of the new proposal
func SubmitProposal(rp *rocketpool.RocketPool, message string, payload []byte, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	proposalCount, err := dao.GetProposalCount(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	tx, err := rocketDAOSecurityProposals.Transact(opts, "propose", message, payload)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error submitting security DAO proposal: %w", err)
	}
	return proposalCount + 1, tx.Hash(), nil
}

// Estimate the gas of VoteOnProposal
func EstimateVoteOnProposalGas(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityProposals.GetTransactionGasInfo(opts, "vote", big.NewInt(int64(proposalId)), support)
}

// Vote on a submitted proposal
func VoteOnProposal(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityProposals.Transact(opts, "vote", big.NewInt(int64(proposalId)), support)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error voting on security DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of CancelProposal
func EstimateCancelProposalGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityProposals.GetTransactionGasInfo(opts, "cancel", big.NewInt(int64(proposalId)))
}

// Cancel a submitted proposal
func CancelProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityProposals.Transact(opts, "cancel", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error cancelling security DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ExecuteProposal
func EstimateExecuteProposalGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityProposals.GetTransactionGasInfo(opts, "execute", big.NewInt(int64(proposalId)))
}

// Execute a submitted proposal
func ExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityProposals, err := getRocketDAOSecurityProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityProposals.Transact(opts, "execute", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error executing security DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDAOSecurityProposalsLock sync.Mutex

func getRocketDAOSecurityProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOSecurityProposalsLock.Lock()
	defer rocketDAOSecurityProposalsLock.Unlock()
	return rp.GetContract("rocketDAOSecurityProposals", opts)
}
