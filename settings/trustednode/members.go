package trustednode

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Config
const (
    MembersSettingsContractName = "rocketDAONodeTrustedSettingsMembers"
    QuorumSettingPath = "members.quorum"
    RPLBondSettingPath = "members.rplbond"
    MinipoolUnbondedMaxSettingPath = "members.minipool.unbonded.max"
)


// Member proposal quorum threshold
func GetQuorum(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    membersSettingsContract, err := getMembersSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := membersSettingsContract.Call(opts, value, "getQuorum"); err != nil {
        return 0, fmt.Errorf("Could not get member quorum threshold: %w", err)
    }
    return eth.WeiToEth(*value), nil
}
func BootstrapQuorum(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, MembersSettingsContractName, QuorumSettingPath, eth.EthToWei(value), opts)
}
func ProposeQuorum(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (uint64, *types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", QuorumSettingPath), MembersSettingsContractName, QuorumSettingPath, eth.EthToWei(value), opts)
}


// RPL bond required for a member
func GetRPLBond(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    membersSettingsContract, err := getMembersSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := membersSettingsContract.Call(opts, value, "getRPLBond"); err != nil {
        return nil, fmt.Errorf("Could not get member RPL bond amount: %w", err)
    }
    return *value, nil
}
func BootstrapRPLBond(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, MembersSettingsContractName, RPLBondSettingPath, value, opts)
}
func ProposeRPLBond(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (uint64, *types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", RPLBondSettingPath), MembersSettingsContractName, RPLBondSettingPath, value, opts)
}


// The maximum number of unbonded minipools a member can run
func GetMinipoolUnbondedMax(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    membersSettingsContract, err := getMembersSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := membersSettingsContract.Call(opts, value, "getMinipoolUnbondedMax"); err != nil {
        return 0, fmt.Errorf("Could not get member unbonded minipool limit: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapMinipoolUnbondedMax(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, MembersSettingsContractName, MinipoolUnbondedMaxSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeMinipoolUnbondedMax(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, *types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", MinipoolUnbondedMaxSettingPath), MembersSettingsContractName, MinipoolUnbondedMaxSettingPath, big.NewInt(int64(value)), opts)
}


// Get contracts
var membersSettingsContractLock sync.Mutex
func getMembersSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    membersSettingsContractLock.Lock()
    defer membersSettingsContractLock.Unlock()
    return rp.GetContract(MembersSettingsContractName)
}

