package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
)

// Config
const (
	SecuritySettingsContractName           string = "rocketDAOProtocolSettingsSecurity"
	SecurityMembersQuorumSettingPath       string = "members.quorum"
	SecurityMembersLeaveTimeSettingPath    string = "members.leave.time"
	SecurityProposalVoteTimeSettingPath    string = "proposal.vote.time"
	SecurityProposalExecuteTimeSettingPath string = "proposal.execute.time"
	SecurityProposalActionTimeSettingPath  string = "proposal.action.time"
)

// Security council member quorum threshold that must be met for proposals to pass
func GetSecurityMembersQuorum(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	securitySettingsContract, err := getSecuritySettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := securitySettingsContract.Call(opts, value, "getQuorum"); err != nil {
		return nil, fmt.Errorf("error getting security members quorum: %w", err)
	}
	return *value, nil
}
func ProposeSecurityMembersQuorum(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SecurityMembersQuorumSettingPath), SecuritySettingsContractName, SecurityMembersQuorumSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSecurityMembersQuorumGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SecurityMembersQuorumSettingPath), SecuritySettingsContractName, SecurityMembersQuorumSettingPath, value, blockNumber, treeNodes, opts)
}

// How long a member must give notice for before manually leaving the security council
func GetSecurityMembersLeaveTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	securitySettingsContract, err := getSecuritySettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := securitySettingsContract.Call(opts, value, "getLeaveTime"); err != nil {
		return 0, fmt.Errorf("error getting security members leave time: %w", err)
	}
	return time.Second * time.Duration((*value).Uint64()), nil
}
func ProposeSecurityMembersLeaveTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SecurityMembersLeaveTimeSettingPath), SecuritySettingsContractName, SecurityMembersLeaveTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSecurityMembersLeaveTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SecurityMembersLeaveTimeSettingPath), SecuritySettingsContractName, SecurityMembersLeaveTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// How long a security council proposal can be voted on (phase2)
func GetSecurityProposalVoteTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	securitySettingsContract, err := getSecuritySettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := securitySettingsContract.Call(opts, value, "getVoteTime"); err != nil {
		return 0, fmt.Errorf("error getting security proposal vote time: %w", err)
	}
	return time.Second * time.Duration((*value).Uint64()), nil
}
func ProposeSecurityProposalVoteTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SecurityProposalVoteTimeSettingPath), SecuritySettingsContractName, SecurityProposalVoteTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSecurityProposalVoteTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SecurityProposalVoteTimeSettingPath), SecuritySettingsContractName, SecurityProposalVoteTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// How long a security council proposal can be executed after its voting period is finished
func GetSecurityProposalExecuteTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	securitySettingsContract, err := getSecuritySettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := securitySettingsContract.Call(opts, value, "getExecuteTime"); err != nil {
		return 0, fmt.Errorf("error getting security proposal execute time: %w", err)
	}
	return time.Second * time.Duration((*value).Uint64()), nil
}
func ProposeSecurityProposalExecuteTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SecurityProposalExecuteTimeSettingPath), SecuritySettingsContractName, SecurityProposalExecuteTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSecurityProposalExecuteTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SecurityProposalExecuteTimeSettingPath), SecuritySettingsContractName, SecurityProposalExecuteTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// Certain security council proposals require a secondary action to be run after the proposal is successful (joining, leaving etc). This is how long until that action expires.
func GetSecurityProposalActionTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	securitySettingsContract, err := getSecuritySettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := securitySettingsContract.Call(opts, value, "getActionTime"); err != nil {
		return 0, fmt.Errorf("error getting security proposal action time: %w", err)
	}
	return time.Second * time.Duration((*value).Uint64()), nil
}
func ProposeSecurityProposalActionTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SecurityProposalActionTimeSettingPath), SecuritySettingsContractName, SecurityProposalActionTimeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSecurityProposalActionTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SecurityProposalActionTimeSettingPath), SecuritySettingsContractName, SecurityProposalActionTimeSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var securitySettingsContractLock sync.Mutex

func getSecuritySettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	securitySettingsContractLock.Lock()
	defer securitySettingsContractLock.Unlock()
	return rp.GetContract(SecuritySettingsContractName, opts)
}
