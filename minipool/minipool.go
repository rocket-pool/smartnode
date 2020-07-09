package minipool

import (
    "encoding/hex"
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    //"github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    //"github.com/rocket-pool/rocketpool-go/utils/contract"
)


// Contract access locks
var rocketMinipoolManagerLock sync.Mutex


// Minipool details
type MinipoolDetails struct {
    Address common.Address
    Exists bool
    Pubkey []byte
    WithdrawalTotalBalance *big.Int
    WithdrawalNodeBalance *big.Int
    Withdrawable bool
    WithdrawalProcessed bool
}


// Get a node's minipool details
func GetNodeMinipools(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]*MinipoolDetails, error) {

    // Get minipool addresses
    minipoolAddresses, err := GetNodeMinipoolAddresses(rp, nodeAddress)
    if err != nil {
        return []*MinipoolDetails{}, err
    }

    // Data
    var wg errgroup.Group
    details := make([]*MinipoolDetails, len(minipoolAddresses))

    // Load details
    for mi, minipoolAddress := range minipoolAddresses {
        mi, minipoolAddress := mi, minipoolAddress
        wg.Go(func() error {
            minipoolDetails, err := GetMinipoolDetails(rp, minipoolAddress)
            if err == nil { details[mi] = minipoolDetails }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []*MinipoolDetails{}, err
    }

    // Return
    return details, nil

}


// Get a node's minipool addresses
func GetNodeMinipoolAddresses(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]common.Address, error) {

    // Get minipool count
    minipoolCount, err := GetNodeMinipoolCount(rp, nodeAddress)
    if err != nil {
        return []common.Address{}, err
    }

    // Data
    var wg errgroup.Group
    addresses := make([]common.Address, minipoolCount)

    // Load addresses
    for mi := int64(0); mi < minipoolCount; mi++ {
        mi := mi
        wg.Go(func() error {
            address, err := GetNodeMinipoolAt(rp, nodeAddress, mi)
            if err == nil { addresses[mi] = address }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []common.Address{}, err
    }

    // Return
    return addresses, nil

}


// Get a minipool's details
func GetMinipoolDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*MinipoolDetails, error) {

    // Minipool data
    var wg errgroup.Group
    var minipoolExists bool
    var minipoolPubkey []byte
    var minipoolWithdrawalTotalBalance *big.Int
    var minipoolWithdrawalNodeBalance *big.Int
    var minipoolWithdrawable bool
    var minipoolWithdrawalProcessed bool

    // Get exists status
    wg.Go(func() error {
        exists, err := GetMinipoolExists(rp, minipoolAddress)
        if err == nil { minipoolExists = exists }
        return err
    })

    // Get pubkey
    wg.Go(func() error {
        pubkey, err := GetMinipoolPubkey(rp, minipoolAddress)
        if err == nil { minipoolPubkey = pubkey }
        return err
    })

    // Get withdrawal total balance
    wg.Go(func() error {
        withdrawalTotalBalance, err := GetMinipoolWithdrawalTotalBalance(rp, minipoolAddress)
        if err == nil { minipoolWithdrawalTotalBalance = withdrawalTotalBalance }
        return err
    })

    // Get withdrawal node balance
    wg.Go(func() error {
        withdrawalNodeBalance, err := GetMinipoolWithdrawalNodeBalance(rp, minipoolAddress)
        if err == nil { minipoolWithdrawalNodeBalance = withdrawalNodeBalance }
        return err
    })

    // Get withdrawable status
    wg.Go(func() error {
        withdrawable, err := GetMinipoolWithdrawable(rp, minipoolAddress)
        if err == nil { minipoolWithdrawable = withdrawable }
        return err
    })

    // Get withdrawal processed status
    wg.Go(func() error {
        withdrawalProcessed, err := GetMinipoolWithdrawalProcessed(rp, minipoolAddress)
        if err == nil { minipoolWithdrawalProcessed = withdrawalProcessed }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Return
    return &MinipoolDetails{
        Address: minipoolAddress,
        Exists: minipoolExists,
        Pubkey: minipoolPubkey,
        WithdrawalTotalBalance: minipoolWithdrawalTotalBalance,
        WithdrawalNodeBalance: minipoolWithdrawalNodeBalance,
        Withdrawable: minipoolWithdrawable,
        WithdrawalProcessed: minipoolWithdrawalProcessed,
    }, nil

}


// Get a node's minipool count
func GetNodeMinipoolCount(rp *rocketpool.RocketPool, nodeAddress common.Address) (int64, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return 0, err
    }
    minipoolCount := new(*big.Int)
    if err := rocketMinipoolManager.Call(nil, minipoolCount, "getNodeMinipoolCount", nodeAddress); err != nil {
        return 0, fmt.Errorf("Could not get node %v minipool count: %w", nodeAddress.Hex(), err)
    }
    return (*minipoolCount).Int64(), nil
}


// Get a node's minipool address by index
func GetNodeMinipoolAt(rp *rocketpool.RocketPool, nodeAddress common.Address, index int64) (common.Address, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return common.Address{}, err
    }
    minipoolAddress := new(common.Address)
    if err := rocketMinipoolManager.Call(nil, minipoolAddress, "getNodeMinipoolAt", nodeAddress, big.NewInt(index)); err != nil {
        return common.Address{}, fmt.Errorf("Could not get node %v minipool %v address: %w", nodeAddress.Hex(), index, err)
    }
    return *minipoolAddress, nil
}


// Get a minipool address by validator pubkey
func GetMinipoolByPubkey(rp *rocketpool.RocketPool, pubkey []byte) (common.Address, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return common.Address{}, err
    }
    minipoolAddress := new(common.Address)
    if err := rocketMinipoolManager.Call(nil, minipoolAddress, "getMinipoolByPubkey", pubkey); err != nil {
        return common.Address{}, fmt.Errorf("Could not get validator %v minipool address: %w", hex.EncodeToString(pubkey), err)
    }
    return *minipoolAddress, nil
}


// Check whether a minipool exists
func GetMinipoolExists(rp *rocketpool.RocketPool, minipoolAddress common.Address) (bool, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return false, err
    }
    exists := new(bool)
    if err := rocketMinipoolManager.Call(nil, exists, "getMinipoolExists", minipoolAddress); err != nil {
        return false, fmt.Errorf("Could not get minipool %v exists status: %w", minipoolAddress.Hex(), err)
    }
    return *exists, nil
}


// Get a minipool's validator pubkey
func GetMinipoolPubkey(rp *rocketpool.RocketPool, minipoolAddress common.Address) ([]byte, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return []byte{}, err
    }
    pubkey := new([]byte)
    if err := rocketMinipoolManager.Call(nil, pubkey, "getMinipoolPubkey", minipoolAddress); err != nil {
        return []byte{}, fmt.Errorf("Could not get minipool %v pubkey: %w", minipoolAddress.Hex(), err)
    }
    return *pubkey, nil
}


// Get a minipool's total balance at withdrawal
func GetMinipoolWithdrawalTotalBalance(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*big.Int, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return nil, err
    }
    balance := new(*big.Int)
    if err := rocketMinipoolManager.Call(nil, balance, "getMinipoolWithdrawalTotalBalance", minipoolAddress); err != nil {
        return nil, fmt.Errorf("Could not get minipool %v withdrawal total balance: %w", minipoolAddress.Hex(), err)
    }
    return *balance, nil
}


// Get a minipool's node balance at withdrawal
func GetMinipoolWithdrawalNodeBalance(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*big.Int, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return nil, err
    }
    balance := new(*big.Int)
    if err := rocketMinipoolManager.Call(nil, balance, "getMinipoolWithdrawalNodeBalance", minipoolAddress); err != nil {
        return nil, fmt.Errorf("Could not get minipool %v withdrawal node balance: %w", minipoolAddress.Hex(), err)
    }
    return *balance, nil
}


// Check whether a minipool is withdrawable
func GetMinipoolWithdrawable(rp *rocketpool.RocketPool, minipoolAddress common.Address) (bool, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return false, err
    }
    withdrawable := new(bool)
    if err := rocketMinipoolManager.Call(nil, withdrawable, "getMinipoolWithdrawable", minipoolAddress); err != nil {
        return false, fmt.Errorf("Could not get minipool %v withdrawable status: %w", minipoolAddress.Hex(), err)
    }
    return *withdrawable, nil
}


// Check whether a minipool's validator withdrawal has been processed
func GetMinipoolWithdrawalProcessed(rp *rocketpool.RocketPool, minipoolAddress common.Address) (bool, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return false, err
    }
    processed := new(bool)
    if err := rocketMinipoolManager.Call(nil, processed, "getMinipoolWithdrawalProcessed", minipoolAddress); err != nil {
        return false, fmt.Errorf("Could not get minipool %v withdrawal processed status: %w", minipoolAddress.Hex(), err)
    }
    return *processed, nil
}


// Get contracts
func getRocketMinipoolManager(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketMinipoolManagerLock.Lock()
    defer rocketMinipoolManagerLock.Unlock()
    return rp.GetContract("rocketMinipoolManager")
}

