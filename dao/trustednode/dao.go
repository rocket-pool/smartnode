package trustednode

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get the member count
func GetMemberCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return 0, err
    }
    memberCount := new(*big.Int)
    if err := rocketDAONodeTrusted.Call(opts, memberCount, "getMemberCount"); err != nil {
        return 0, fmt.Errorf("Could not get trusted node DAO member count: %w", err)
    }
    return (*memberCount).Uint64(), nil
}


// Get a member address by index
func GetMemberAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return common.Address{}, err
    }
    memberAddress := new(common.Address)
    if err := rocketDAONodeTrusted.Call(opts, memberAddress, "getMemberAt", big.NewInt(int64(index))); err != nil {
        return common.Address{}, fmt.Errorf("Could not get trusted node DAO member %d address: %w", index, err)
    }
    return *memberAddress, nil
}


// Member details
func GetMemberExists(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (bool, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return false, err
    }
    exists := new(bool)
    if err := rocketDAONodeTrusted.Call(opts, exists, "getMemberIsValid", memberAddress); err != nil {
        return false, fmt.Errorf("Could not get trusted node DAO member %s exists status: %w", memberAddress.Hex(), err)
    }
    return *exists, nil
}
func GetMemberID(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (string, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return "", err
    }
    id := new(string)
    if err := rocketDAONodeTrusted.Call(opts, id, "getMemberID", memberAddress); err != nil {
        return "", fmt.Errorf("Could not get trusted node DAO member %s ID: %w", memberAddress.Hex(), err)
    }
    return *id, nil
}
func GetMemberEmail(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (string, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return "", err
    }
    email := new(string)
    if err := rocketDAONodeTrusted.Call(opts, email, "getMemberEmail", memberAddress); err != nil {
        return "", fmt.Errorf("Could not get trusted node DAO member %s email: %w", memberAddress.Hex(), err)
    }
    return *email, nil
}
func GetMemberJoinedBlock(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return 0, err
    }
    joinedBlock := new(*big.Int)
    if err := rocketDAONodeTrusted.Call(opts, joinedBlock, "getMemberJoinedBlock", memberAddress); err != nil {
        return 0, fmt.Errorf("Could not get trusted node DAO member %s joined block: %w", memberAddress.Hex(), err)
    }
    return (*joinedBlock).Uint64(), nil
}
func GetMemberLastProposalBlock(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return 0, err
    }
    lastProposalBlock := new(*big.Int)
    if err := rocketDAONodeTrusted.Call(opts, lastProposalBlock, "getMemberLastProposalBlock", memberAddress); err != nil {
        return 0, fmt.Errorf("Could not get trusted node DAO member %s last proposal block: %w", memberAddress.Hex(), err)
    }
    return (*lastProposalBlock).Uint64(), nil
}
func GetMemberRPLBondAmount(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (*big.Int, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    rplBondAmount := new(*big.Int)
    if err := rocketDAONodeTrusted.Call(opts, rplBondAmount, "getMemberRPLBondAmount", memberAddress); err != nil {
        return nil, fmt.Errorf("Could not get trusted node DAO member %s RPL bond amount: %w", memberAddress.Hex(), err)
    }
    return *rplBondAmount, nil
}
func GetMemberUnbondedValidatorCount(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, err) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return 0, err
    }
    unbondedValidatorCount := new(*big.Int)
    if err := rocketDAONodeTrusted.Call(opts, unbondedValidatorCount, "getMemberUnbondedValidatorCount", memberAddress); err != nil {
        return 0, fmt.Errorf("Could not get trusted node DAO member %s unbonded validator count: %w", memberAddress.Hex(), err)
    }
    return (*unbondedValidatorCount).Uint64(), nil
}


// Bootstrap a bool setting
func BootstrapBool(rp *rocketpool.RocketPool, contractName, settingPath string, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrusted.Transact(opts, "bootstrapSettingBool", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap trusted node setting %s.%s: %w", contractName, settingPath, err)
    }
    return txReceipt, nil
}


// Bootstrap a uint256 setting
func BootstrapUint(rp *rocketpool.RocketPool, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrusted.Transact(opts, "bootstrapSettingUint", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap trusted node setting %s.%s: %w", contractName, settingPath, err)
    }
    return txReceipt, nil
}


// Bootstrap a DAO member
func BootstrapMember(rp *rocketpool.RocketPool, id, email string, nodeAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrusted.Transact(opts, "bootstrapMember", id, email, nodeAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap trusted node member %s: %w", id, err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDAONodeTrustedLock sync.Mutex
func getRocketDAONodeTrusted(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAONodeTrustedLock.Lock()
    defer rocketDAONodeTrustedLock.Unlock()
    return rp.GetContract("rocketDAONodeTrusted")
}

