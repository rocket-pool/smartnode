package minipool

import (
    "fmt"
    "math/big"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    rptypes "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Minipool detail types
type StatusDetails struct {
    Status rptypes.MinipoolStatus   `json:"status"`
    StatusBlock int64               `json:"statusBlock"`
    StatusTime time.Time            `json:"statusTime"`
}
type NodeDetails struct {
    Address common.Address          `json:"address"`
    Fee float64                     `json:"fee"`
    DepositBalance *big.Int         `json:"depositBalance"`
    RefundBalance *big.Int          `json:"refundBalance"`
    DepositAssigned bool            `json:"depositAssigned"`
}
type UserDetails struct {
    DepositBalance *big.Int         `json:"depositBalance"`
    DepositAssigned bool            `json:"depositAssigned"`
    DepositAssignedTime time.Time   `json:"depositAssignedTime"`
}
type StakingDetails struct {
    StartBalance *big.Int           `json:"startBalance"`
    EndBalance *big.Int             `json:"endBalance"`
    StartEpoch int64                `json:"startEpoch"`
    EndEpoch int64                  `json:"endEpoch"`
    UserStartEpoch int64            `json:"userStartEpoch"`
}


// Minipool contract
type Minipool struct {
    Address common.Address
    Contract *bind.BoundContract
    RocketPool *rocketpool.RocketPool
}


// Create new minipool contract
func NewMinipool(rp *rocketpool.RocketPool, address common.Address) (*Minipool, error) {

    // Get contract
    contract, err := getMinipoolContract(rp, address)
    if err != nil {
        return nil, err
    }

    // Create and return
    return &Minipool{
        Address: address,
        Contract: contract,
        RocketPool: rp,
    }, nil
}


// Get status details
func (mp *Minipool) GetStatusDetails() (StatusDetails, error) {

    // Data
    var wg errgroup.Group
    var status rptypes.MinipoolStatus
    var statusBlock int64
    var statusTime time.Time

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus()
        return err
    })
    wg.Go(func() error {
        var err error
        statusBlock, err = mp.GetStatusBlock()
        return err
    })
    wg.Go(func() error {
        var err error
        statusTime, err = mp.GetStatusTime()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return StatusDetails{}, err
    }

    // Return
    return StatusDetails{
        Status: status,
        StatusBlock: statusBlock,
        StatusTime: statusTime,
    }, nil

}
func (mp *Minipool) GetStatus() (rptypes.MinipoolStatus, error) {
    status := new(rptypes.MinipoolStatus)
    if err := mp.Contract.Call(nil, status, "getStatus"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s status: %w", mp.Address.Hex(), err)
    }
    return *status, nil
}
func (mp *Minipool) GetStatusBlock() (int64, error) {
    statusBlock := new(*big.Int)
    if err := mp.Contract.Call(nil, statusBlock, "getStatusBlock"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s status changed block: %w", mp.Address.Hex(), err)
    }
    return (*statusBlock).Int64(), nil
}
func (mp *Minipool) GetStatusTime() (time.Time, error) {
    statusTime := new(*big.Int)
    if err := mp.Contract.Call(nil, statusTime, "getStatusTime"); err != nil {
        return time.Unix(0, 0), fmt.Errorf("Could not get minipool %s status changed time: %w", mp.Address.Hex(), err)
    }
    return time.Unix((*statusTime).Int64(), 0), nil
}


// Get deposit type
func (mp *Minipool) GetDepositType() (rptypes.MinipoolDeposit, error) {
    depositType := new(rptypes.MinipoolDeposit)
    if err := mp.Contract.Call(nil, depositType, "getDepositType"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s deposit type: %w", mp.Address.Hex(), err)
    }
    return *depositType, nil
}


// Get node details
func (mp *Minipool) GetNodeDetails() (NodeDetails, error) {

    // Data
    var wg errgroup.Group
    var address common.Address
    var fee float64
    var depositBalance *big.Int
    var refundBalance *big.Int
    var depositAssigned bool

    // Load data
    wg.Go(func() error {
        var err error
        address, err = mp.GetNodeAddress()
        return err
    })
    wg.Go(func() error {
        var err error
        fee, err = mp.GetNodeFee()
        return err
    })
    wg.Go(func() error {
        var err error
        depositBalance, err = mp.GetNodeDepositBalance()
        return err
    })
    wg.Go(func() error {
        var err error
        refundBalance, err = mp.GetNodeRefundBalance()
        return err
    })
    wg.Go(func() error {
        var err error
        depositAssigned, err = mp.GetNodeDepositAssigned()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return NodeDetails{}, err
    }

    // Return
    return NodeDetails{
        Address: address,
        Fee: fee,
        DepositBalance: depositBalance,
        RefundBalance: refundBalance,
        DepositAssigned: depositAssigned,
    }, nil

}
func (mp *Minipool) GetNodeAddress() (common.Address, error) {
    nodeAddress := new(common.Address)
    if err := mp.Contract.Call(nil, nodeAddress, "getNodeAddress"); err != nil {
        return common.Address{}, fmt.Errorf("Could not get minipool %s node address: %w", mp.Address.Hex(), err)
    }
    return *nodeAddress, nil
}
func (mp *Minipool) GetNodeFee() (float64, error) {
    nodeFee := new(*big.Int)
    if err := mp.Contract.Call(nil, nodeFee, "getNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s node fee: %w", mp.Address.Hex(), err)
    }
    return eth.WeiToEth(*nodeFee), nil
}
func (mp *Minipool) GetNodeDepositBalance() (*big.Int, error) {
    nodeDepositBalance := new(*big.Int)
    if err := mp.Contract.Call(nil, nodeDepositBalance, "getNodeDepositBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s node deposit balance: %w", mp.Address.Hex(), err)
    }
    return *nodeDepositBalance, nil
}
func (mp *Minipool) GetNodeRefundBalance() (*big.Int, error) {
    nodeRefundBalance := new(*big.Int)
    if err := mp.Contract.Call(nil, nodeRefundBalance, "getNodeRefundBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s node refund balance: %w", mp.Address.Hex(), err)
    }
    return *nodeRefundBalance, nil
}
func (mp *Minipool) GetNodeDepositAssigned() (bool, error) {
    nodeDepositAssigned := new(bool)
    if err := mp.Contract.Call(nil, nodeDepositAssigned, "getNodeDepositAssigned"); err != nil {
        return false, fmt.Errorf("Could not get minipool %s node deposit assigned status: %w", mp.Address.Hex(), err)
    }
    return *nodeDepositAssigned, nil
}


// Get user deposit details
func (mp *Minipool) GetUserDetails() (UserDetails, error) {

    // Data
    var wg errgroup.Group
    var depositBalance *big.Int
    var depositAssigned bool
    var depositAssignedTime time.Time

    // Load data
    wg.Go(func() error {
        var err error
        depositBalance, err = mp.GetUserDepositBalance()
        return err
    })
    wg.Go(func() error {
        var err error
        depositAssigned, err = mp.GetUserDepositAssigned()
        return err
    })
    wg.Go(func() error {
        var err error
        depositAssignedTime, err = mp.GetUserDepositAssignedTime()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return UserDetails{}, err
    }

    // Return
    return UserDetails{
        DepositBalance: depositBalance,
        DepositAssigned: depositAssigned,
        DepositAssignedTime: depositAssignedTime,
    }, nil

}
func (mp *Minipool) GetUserDepositBalance() (*big.Int, error) {
    userDepositBalance := new(*big.Int)
    if err := mp.Contract.Call(nil, userDepositBalance, "getUserDepositBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s user deposit balance: %w", mp.Address.Hex(), err)
    }
    return *userDepositBalance, nil
}
func (mp *Minipool) GetUserDepositAssigned() (bool, error) {
    userDepositAssigned := new(bool)
    if err := mp.Contract.Call(nil, userDepositAssigned, "getUserDepositAssigned"); err != nil {
        return false, fmt.Errorf("Could not get minipool %s user deposit assigned status: %w", mp.Address.Hex(), err)
    }
    return *userDepositAssigned, nil
}
func (mp *Minipool) GetUserDepositAssignedTime() (time.Time, error) {
    depositAssignedTime := new(*big.Int)
    if err := mp.Contract.Call(nil, depositAssignedTime, "getUserDepositAssignedTime"); err != nil {
        return time.Unix(0, 0), fmt.Errorf("Could not get minipool %s user deposit assigned time: %w", mp.Address.Hex(), err)
    }
    return time.Unix((*depositAssignedTime).Int64(), 0), nil
}


// Get staking details
func (mp *Minipool) GetStakingDetails() (StakingDetails, error) {

    // Data
    var wg errgroup.Group
    var startBalance *big.Int
    var endBalance *big.Int
    var startEpoch int64
    var endEpoch int64
    var userStartEpoch int64

    // Load data
    wg.Go(func() error {
        var err error
        startBalance, err = mp.GetStakingStartBalance()
        return err
    })
    wg.Go(func() error {
        var err error
        endBalance, err = mp.GetStakingEndBalance()
        return err
    })
    wg.Go(func() error {
        var err error
        startEpoch, err = mp.GetStakingStartEpoch()
        return err
    })
    wg.Go(func() error {
        var err error
        endEpoch, err = mp.GetStakingEndEpoch()
        return err
    })
    wg.Go(func() error {
        var err error
        userStartEpoch, err = mp.GetStakingUserStartEpoch()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return StakingDetails{}, err
    }

    // Return
    return StakingDetails{
        StartBalance: startBalance,
        EndBalance: endBalance,
        StartEpoch: startEpoch,
        EndEpoch: endEpoch,
        UserStartEpoch: userStartEpoch,
    }, nil

}
func (mp *Minipool) GetStakingStartBalance() (*big.Int, error) {
    stakingStartBalance := new(*big.Int)
    if err := mp.Contract.Call(nil, stakingStartBalance, "getStakingStartBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s staking start balance: %w", mp.Address.Hex(), err)
    }
    return *stakingStartBalance, nil
}
func (mp *Minipool) GetStakingEndBalance() (*big.Int, error) {
    stakingEndBalance := new(*big.Int)
    if err := mp.Contract.Call(nil, stakingEndBalance, "getStakingEndBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s staking end balance: %w", mp.Address.Hex(), err)
    }
    return *stakingEndBalance, nil
}
func (mp *Minipool) GetStakingStartEpoch() (int64, error) {
    stakingStartEpoch := new(*big.Int)
    if err := mp.Contract.Call(nil, stakingStartEpoch, "getStakingStartEpoch"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s staking start epoch: %w", mp.Address.Hex(), err)
    }
    return (*stakingStartEpoch).Int64(), nil
}
func (mp *Minipool) GetStakingEndEpoch() (int64, error) {
    stakingEndEpoch := new(*big.Int)
    if err := mp.Contract.Call(nil, stakingEndEpoch, "getStakingEndEpoch"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s staking end epoch: %w", mp.Address.Hex(), err)
    }
    return (*stakingEndEpoch).Int64(), nil
}
func (mp *Minipool) GetStakingUserStartEpoch() (int64, error) {
    stakingUserStartEpoch := new(*big.Int)
    if err := mp.Contract.Call(nil, stakingUserStartEpoch, "getStakingUserStartEpoch"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s staking user start epoch: %w", mp.Address.Hex(), err)
    }
    return (*stakingUserStartEpoch).Int64(), nil
}


// Refund node ETH from the minipool
func (mp *Minipool) Refund(opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := contract.Transact(mp.RocketPool.Client, mp.Contract, opts, "refund")
    if err != nil {
        return nil, fmt.Errorf("Could not refund from minipool %s: %w", mp.Address.Hex(), err)
    }
    return txReceipt, nil
}


// Progress the prelaunch minipool to staking
func (mp *Minipool) Stake(validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := contract.Transact(mp.RocketPool.Client, mp.Contract, opts, "stake", validatorPubkey[:], validatorSignature[:], depositDataRoot)
    if err != nil {
        return nil, fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err)
    }
    return txReceipt, nil
}


// Withdraw node balances & rewards from the withdrawable minipool and close it
func (mp *Minipool) Withdraw(opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := contract.Transact(mp.RocketPool.Client, mp.Contract, opts, "withdraw")
    if err != nil {
        return nil, fmt.Errorf("Could not withdraw from minipool %s: %w", mp.Address.Hex(), err)
    }
    return txReceipt, nil
}


// Dissolve the initialized or prelaunch minipool
func (mp *Minipool) Dissolve(opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := contract.Transact(mp.RocketPool.Client, mp.Contract, opts, "dissolve")
    if err != nil {
        return nil, fmt.Errorf("Could not dissolve minipool %s: %w", mp.Address.Hex(), err)
    }
    return txReceipt, nil
}


// Withdraw node balances from the dissolved minipool and close it
func (mp *Minipool) Close(opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := contract.Transact(mp.RocketPool.Client, mp.Contract, opts, "close")
    if err != nil {
        return nil, fmt.Errorf("Could not close minipool %s: %w", mp.Address.Hex(), err)
    }
    return txReceipt, nil
}


// Get a minipool contract
var rocketMinipoolLock sync.Mutex
func getMinipoolContract(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*bind.BoundContract, error) {
    rocketMinipoolLock.Lock()
    defer rocketMinipoolLock.Unlock()
    return rp.MakeContract("rocketMinipool", minipoolAddress)
}

