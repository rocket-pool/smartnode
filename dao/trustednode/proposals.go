package trustednode

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/strings"
)

// Estimate the gas of ProposeInviteMember
func EstimateProposeInviteMemberGas(rp *rocketpool.RocketPool, message string, newMemberAddress common.Address, newMemberId, newMemberUrl string, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	newMemberUrl = strings.Sanitize(newMemberUrl)
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalInvite", newMemberId, newMemberUrl, newMemberAddress)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode invite member proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to invite a new member to the trusted node DAO
func ProposeInviteMember(rp *rocketpool.RocketPool, message string, newMemberAddress common.Address, newMemberId, newMemberUrl string, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	newMemberUrl = strings.Sanitize(newMemberUrl)
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalInvite", newMemberId, newMemberUrl, newMemberAddress)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode invite member proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeMemberLeave
func EstimateProposeMemberLeaveGas(rp *rocketpool.RocketPool, message string, memberAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalLeave", memberAddress)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode member leave proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal for a member to leave the trusted node DAO
func ProposeMemberLeave(rp *rocketpool.RocketPool, message string, memberAddress common.Address, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalLeave", memberAddress)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode member leave proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeReplaceMember
func EstimateProposeReplaceMemberGas(rp *rocketpool.RocketPool, message string, memberAddress, newMemberAddress common.Address, newMemberId, newMemberUrl string, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	newMemberUrl = strings.Sanitize(newMemberUrl)
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalReplace", memberAddress, newMemberId, newMemberUrl, newMemberAddress)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode replace member proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to replace a member in the trusted node DAO
func ProposeReplaceMember(rp *rocketpool.RocketPool, message string, memberAddress, newMemberAddress common.Address, newMemberId, newMemberUrl string, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	newMemberUrl = strings.Sanitize(newMemberUrl)
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalReplace", memberAddress, newMemberId, newMemberUrl, newMemberAddress)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode replace member proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeKickMember
func EstimateProposeKickMemberGas(rp *rocketpool.RocketPool, message string, memberAddress common.Address, rplFineAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalKick", memberAddress, rplFineAmount)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode kick member proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to kick a member from the trusted node DAO
func ProposeKickMember(rp *rocketpool.RocketPool, message string, memberAddress common.Address, rplFineAmount *big.Int, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalKick", memberAddress, rplFineAmount)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode kick member proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeSetBool
func EstimateProposeSetBoolGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode set bool setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to update a bool trusted node DAO setting
func ProposeSetBool(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode set bool setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeSetUint
func EstimateProposeSetUintGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode set uint setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to update a uint trusted node DAO setting
func ProposeSetUint(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode set uint setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of ProposeUpgradeContract
func EstimateProposeUpgradeContractGas(rp *rocketpool.RocketPool, message, upgradeType, contractName, contractAbi string, contractAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	compressedAbi, err := rocketpool.EncodeAbiStr(contractAbi)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalUpgrade", upgradeType, contractName, compressedAbi, contractAddress)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("Could not encode upgrade contract proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, opts)
}

// Submit a proposal to upgrade a contract
func ProposeUpgradeContract(rp *rocketpool.RocketPool, message, upgradeType, contractName, contractAbi string, contractAddress common.Address, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	compressedAbi, err := rocketpool.EncodeAbiStr(contractAbi)
	if err != nil {
		return 0, common.Hash{}, err
	}
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalUpgrade", upgradeType, contractName, compressedAbi, contractAddress)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not encode upgrade contract proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, opts)
}

// Estimate the gas of a proposal submission
func EstimateProposalGas(rp *rocketpool.RocketPool, message string, payload []byte, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedProposals.GetTransactionGasInfo(opts, "propose", message, payload)
}

// Submit a trusted node DAO proposal
// Returns the ID of the new proposal
func SubmitProposal(rp *rocketpool.RocketPool, message string, payload []byte, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	proposalCount, err := dao.GetProposalCount(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedProposals.Transact(opts, "propose", message, payload)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not submit trusted node DAO proposal: %w", err)
	}
	return proposalCount + 1, tx.Hash(), nil
}

// Estimate the gas of CancelProposal
func EstimateCancelProposalGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedProposals.GetTransactionGasInfo(opts, "cancel", big.NewInt(int64(proposalId)))
}

// Cancel a submitted proposal
func CancelProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedProposals.Transact(opts, "cancel", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not cancel trusted node DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of VoteOnProposal
func EstimateVoteOnProposalGas(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedProposals.GetTransactionGasInfo(opts, "vote", big.NewInt(int64(proposalId)), support)
}

// Vote on a submitted proposal
func VoteOnProposal(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedProposals.Transact(opts, "vote", big.NewInt(int64(proposalId)), support)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not vote on trusted node DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ExecuteProposal
func EstimateExecuteProposalGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedProposals.GetTransactionGasInfo(opts, "execute", big.NewInt(int64(proposalId)))
}

// Execute a submitted proposal
func ExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedProposals.Transact(opts, "execute", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not execute trusted node DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDAONodeTrustedProposalsLock sync.Mutex

func getRocketDAONodeTrustedProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAONodeTrustedProposalsLock.Lock()
	defer rocketDAONodeTrustedProposalsLock.Unlock()
	return rp.GetContract("rocketDAONodeTrustedProposals", opts)
}
