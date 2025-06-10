package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
)

// Config
const (
	MegapoolSettingsContractName           string = "rocketDAOProtocolSettingsMegapool"
	MegapoolTimeBeforeDissolveSettingsPath string = "megapool.time.before.dissolve"
	MegapoolMaximumMegapoolEthPenaltyPath  string = "maximum.megapool.eth.penalty"
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

func ProposeMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MegapoolTimeBeforeDissolveSettingsPath), MegapoolSettingsContractName, MegapoolTimeBeforeDissolveSettingsPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MegapoolTimeBeforeDissolveSettingsPath), MegapoolSettingsContractName, MegapoolTimeBeforeDissolveSettingsPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var megapoolSettingsContractLock sync.Mutex

func getMegapoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	megapoolSettingsContractLock.Lock()
	defer megapoolSettingsContractLock.Unlock()
	return rp.GetContract(MegapoolSettingsContractName, opts)
}
