package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Config
const (
	ProposalsSettingsContractName  string = "rocketDAOProtocolSettingsProposals"
	VoteTimeSettingPath            string = "proposal.vote.time"
	VoteDelayTimeSettingPath       string = "proposal.vote.delay.time"
	ExecuteTimeSettingPath         string = "proposal.execute.time"
	ProposalBondSettingPath        string = "proposal.bond"
	ChallengeBondSettingPath       string = "proposal.challenge.bond"
	ChallengePeriodSettingPath     string = "proposal.challenge.period"
	ProposalQuorumSettingPath      string = "proposal.quorum"
	ProposalVetoQuorumSettingPath  string = "proposal.veto.quorum"
	ProposalMaxBlockAgeSettingPath string = "proposal.max.block.age"
)

// How long a proposal can be voted on before expiring
func GetVoteTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getVoteTime"); err != nil {
		return 0, fmt.Errorf("error getting vote time: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeVoteTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", VoteTimeSettingPath), ProposalsSettingsContractName, VoteTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeVoteTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", VoteTimeSettingPath), ProposalsSettingsContractName, VoteTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// How long before a proposal can be voted on after its created
func GetVoteDelayTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getVoteDelayTime"); err != nil {
		return 0, fmt.Errorf("error getting vote delay time: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeVoteDelayTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", VoteDelayTimeSettingPath), ProposalsSettingsContractName, VoteDelayTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeVoteDelayTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", VoteDelayTimeSettingPath), ProposalsSettingsContractName, VoteDelayTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// How long after a succesful proposal can it be executed before it expires
func GetExecuteTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getExecuteTime"); err != nil {
		return 0, fmt.Errorf("error getting execute time: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeExecuteTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ExecuteTimeSettingPath), ProposalsSettingsContractName, ExecuteTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeExecuteTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ExecuteTimeSettingPath), ProposalsSettingsContractName, ExecuteTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// How much RPL is locked when creating a proposal
func GetProposalBond(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getProposalBond"); err != nil {
		return nil, fmt.Errorf("error getting proposal bond: %w", err)
	}
	return *value, nil
}
func ProposeProposalBond(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ProposalBondSettingPath), ProposalsSettingsContractName, ProposalBondSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeProposalBondGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ProposalBondSettingPath), ProposalsSettingsContractName, ProposalBondSettingPath, value, blockNumber, treeNodes, opts)
}

// How much RPL is locked when challenging a proposal
func GetChallengeBond(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getChallengeBond"); err != nil {
		return nil, fmt.Errorf("error getting challenge bond: %w", err)
	}
	return *value, nil
}
func ProposeChallengeBond(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ChallengeBondSettingPath), ProposalsSettingsContractName, ChallengeBondSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeChallengeBondGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ChallengeBondSettingPath), ProposalsSettingsContractName, ChallengeBondSettingPath, value, blockNumber, treeNodes, opts)
}

// How long a proposer has to respond to a challenge before the proposal is defeated
func GetChallengePeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getChallengePeriod"); err != nil {
		return 0, fmt.Errorf("error getting challenge period: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeChallengePeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ChallengePeriodSettingPath), ProposalsSettingsContractName, ChallengePeriodSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeChallengePeriodGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ChallengePeriodSettingPath), ProposalsSettingsContractName, ChallengePeriodSettingPath, value, blockNumber, treeNodes, opts)
}

// The minimum amount of voting power a proposal needs to succeed
func GetProposalQuorum(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getProposalQuorum"); err != nil {
		return 0, fmt.Errorf("error getting proposal quorum: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeProposalQuorum(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ProposalQuorumSettingPath), ProposalsSettingsContractName, ProposalQuorumSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeProposalQuorumGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ProposalQuorumSettingPath), ProposalsSettingsContractName, ProposalQuorumSettingPath, value, blockNumber, treeNodes, opts)
}

// The amount of voting power vetoing a proposal require to veto it
func GetProposalVetoQuorum(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getProposalVetoQuorum"); err != nil {
		return 0, fmt.Errorf("error getting proposal veto quorum: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeProposalVetoQuorum(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ProposalVetoQuorumSettingPath), ProposalsSettingsContractName, ProposalVetoQuorumSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeProposalVetoQuorumGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ProposalVetoQuorumSettingPath), ProposalsSettingsContractName, ProposalVetoQuorumSettingPath, value, blockNumber, treeNodes, opts)
}

// The maximum number of blocks old a proposal can be submitted for
func GetProposalMaxBlockAge(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getProposalMaxBlockAge"); err != nil {
		return 0, fmt.Errorf("error getting proposal max block age: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeProposalMaxBlockAge(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ProposalMaxBlockAgeSettingPath), ProposalsSettingsContractName, ProposalMaxBlockAgeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeProposalMaxBlockAgeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ProposalMaxBlockAgeSettingPath), ProposalsSettingsContractName, ProposalMaxBlockAgeSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var proposalsSettingsContractLock sync.Mutex

func getProposalsSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	proposalsSettingsContractLock.Lock()
	defer proposalsSettingsContractLock.Unlock()
	return rp.GetContract(ProposalsSettingsContractName, opts)
}
