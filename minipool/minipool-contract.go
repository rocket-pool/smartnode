package minipool

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Minipool detail types
type StatusDetails struct {
    Status rptypes.MinipoolStatus   `json:"status"`
    StatusBlock uint64              `json:"statusBlock"`
    StatusTime time.Time            `json:"statusTime"`
}
type NodeDetails struct {
    Address common.Address          `json:"address"`
    Fee float64                     `json:"fee"`
    DepositBalance *big.Int         `json:"depositBalance"`
    RefundBalance *big.Int          `json:"refundBalance"`
    DepositAssigned bool            `json:"depositAssigned"`
    Withdrawn bool                  `json:"withdrawn"`
}
type UserDetails struct {
    DepositBalance *big.Int         `json:"depositBalance"`
    DepositAssigned bool            `json:"depositAssigned"`
    DepositAssignedTime time.Time   `json:"depositAssignedTime"`
}
type StakingDetails struct {
    StartBalance *big.Int           `json:"startBalance"`
    EndBalance *big.Int             `json:"endBalance"`
    BalanceWithdrawn bool           `json:"balanceWithdrawn"`
}


// Minipool contract
type Minipool struct {
    Address common.Address
    Contract *rocketpool.Contract
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
func (mp *Minipool) GetStatusDetails(opts *bind.CallOpts) (StatusDetails, error) {

    // Data
    var wg errgroup.Group
    var status rptypes.MinipoolStatus
    var statusBlock uint64
    var statusTime time.Time

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        statusBlock, err = mp.GetStatusBlock(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        statusTime, err = mp.GetStatusTime(opts)
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
func (mp *Minipool) GetStatus(opts *bind.CallOpts) (rptypes.MinipoolStatus, error) {
    status := new(uint8)
    if err := mp.Contract.Call(opts, status, "getStatus"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s status: %w", mp.Address.Hex(), err)
    }
    return rptypes.MinipoolStatus(*status), nil
}
func (mp *Minipool) GetStatusBlock(opts *bind.CallOpts) (uint64, error) {
    statusBlock := new(*big.Int)
    if err := mp.Contract.Call(opts, statusBlock, "getStatusBlock"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s status changed block: %w", mp.Address.Hex(), err)
    }
    return (*statusBlock).Uint64(), nil
}
func (mp *Minipool) GetStatusTime(opts *bind.CallOpts) (time.Time, error) {
    statusTime := new(*big.Int)
    if err := mp.Contract.Call(opts, statusTime, "getStatusTime"); err != nil {
        return time.Unix(0, 0), fmt.Errorf("Could not get minipool %s status changed time: %w", mp.Address.Hex(), err)
    }
    return time.Unix((*statusTime).Int64(), 0), nil
}


// Get deposit type
func (mp *Minipool) GetDepositType(opts *bind.CallOpts) (rptypes.MinipoolDeposit, error) {
    depositType := new(uint8)
    if err := mp.Contract.Call(opts, depositType, "getDepositType"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s deposit type: %w", mp.Address.Hex(), err)
    }
    return rptypes.MinipoolDeposit(*depositType), nil
}


// Get node details
func (mp *Minipool) GetNodeDetails(opts *bind.CallOpts) (NodeDetails, error) {

    // Data
    var wg errgroup.Group
    var address common.Address
    var fee float64
    var depositBalance *big.Int
    var refundBalance *big.Int
    var depositAssigned bool
    var withdrawn bool

    // Load data
    wg.Go(func() error {
        var err error
        address, err = mp.GetNodeAddress(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        fee, err = mp.GetNodeFee(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        depositBalance, err = mp.GetNodeDepositBalance(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        refundBalance, err = mp.GetNodeRefundBalance(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        depositAssigned, err = mp.GetNodeDepositAssigned(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        withdrawn, err = mp.GetNodeWithdrawn(opts)
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
        Withdrawn: withdrawn,
    }, nil

}
func (mp *Minipool) GetNodeAddress(opts *bind.CallOpts) (common.Address, error) {
    nodeAddress := new(common.Address)
    if err := mp.Contract.Call(opts, nodeAddress, "getNodeAddress"); err != nil {
        return common.Address{}, fmt.Errorf("Could not get minipool %s node address: %w", mp.Address.Hex(), err)
    }
    return *nodeAddress, nil
}
func (mp *Minipool) GetNodeFee(opts *bind.CallOpts) (float64, error) {
    nodeFee := new(*big.Int)
    if err := mp.Contract.Call(opts, nodeFee, "getNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get minipool %s node fee: %w", mp.Address.Hex(), err)
    }
    return eth.WeiToEth(*nodeFee), nil
}
func (mp *Minipool) GetNodeDepositBalance(opts *bind.CallOpts) (*big.Int, error) {
    nodeDepositBalance := new(*big.Int)
    if err := mp.Contract.Call(opts, nodeDepositBalance, "getNodeDepositBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s node deposit balance: %w", mp.Address.Hex(), err)
    }
    return *nodeDepositBalance, nil
}
func (mp *Minipool) GetNodeRefundBalance(opts *bind.CallOpts) (*big.Int, error) {
    nodeRefundBalance := new(*big.Int)
    if err := mp.Contract.Call(opts, nodeRefundBalance, "getNodeRefundBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s node refund balance: %w", mp.Address.Hex(), err)
    }
    return *nodeRefundBalance, nil
}
func (mp *Minipool) GetNodeDepositAssigned(opts *bind.CallOpts) (bool, error) {
    nodeDepositAssigned := new(bool)
    if err := mp.Contract.Call(opts, nodeDepositAssigned, "getNodeDepositAssigned"); err != nil {
        return false, fmt.Errorf("Could not get minipool %s node deposit assigned status: %w", mp.Address.Hex(), err)
    }
    return *nodeDepositAssigned, nil
}
func (mp *Minipool) GetNodeWithdrawn(opts *bind.CallOpts) (bool, error) {
    nodeWithdrawn := new(bool)
    if err := mp.Contract.Call(opts, nodeWithdrawn, "getNodeWithdrawn"); err != nil {
        return false, fmt.Errorf("Could not get minipool %s node withdrawn status: %w", mp.Address.Hex(), err)
    }
    return *nodeWithdrawn, nil
}


// Get user deposit details
func (mp *Minipool) GetUserDetails(opts *bind.CallOpts) (UserDetails, error) {

    // Data
    var wg errgroup.Group
    var depositBalance *big.Int
    var depositAssigned bool
    var depositAssignedTime time.Time

    // Load data
    wg.Go(func() error {
        var err error
        depositBalance, err = mp.GetUserDepositBalance(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        depositAssigned, err = mp.GetUserDepositAssigned(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        depositAssignedTime, err = mp.GetUserDepositAssignedTime(opts)
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
func (mp *Minipool) GetUserDepositBalance(opts *bind.CallOpts) (*big.Int, error) {
    userDepositBalance := new(*big.Int)
    if err := mp.Contract.Call(opts, userDepositBalance, "getUserDepositBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s user deposit balance: %w", mp.Address.Hex(), err)
    }
    return *userDepositBalance, nil
}
func (mp *Minipool) GetUserDepositAssigned(opts *bind.CallOpts) (bool, error) {
    userDepositAssigned := new(bool)
    if err := mp.Contract.Call(opts, userDepositAssigned, "getUserDepositAssigned"); err != nil {
        return false, fmt.Errorf("Could not get minipool %s user deposit assigned status: %w", mp.Address.Hex(), err)
    }
    return *userDepositAssigned, nil
}
func (mp *Minipool) GetUserDepositAssignedTime(opts *bind.CallOpts) (time.Time, error) {
    depositAssignedTime := new(*big.Int)
    if err := mp.Contract.Call(opts, depositAssignedTime, "getUserDepositAssignedTime"); err != nil {
        return time.Unix(0, 0), fmt.Errorf("Could not get minipool %s user deposit assigned time: %w", mp.Address.Hex(), err)
    }
    return time.Unix((*depositAssignedTime).Int64(), 0), nil
}


// Get staking details
func (mp *Minipool) GetStakingDetails(opts *bind.CallOpts) (StakingDetails, error) {

    // Data
    var wg errgroup.Group
    var startBalance *big.Int
    var endBalance *big.Int
    var balanceWithdrawn bool

    // Load data
    wg.Go(func() error {
        var err error
        startBalance, err = mp.GetStakingStartBalance(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        endBalance, err = mp.GetStakingEndBalance(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        balanceWithdrawn, err = mp.GetValidatorBalanceWithdrawn(opts)
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
        BalanceWithdrawn: balanceWithdrawn,
    }, nil

}
func (mp *Minipool) GetStakingStartBalance(opts *bind.CallOpts) (*big.Int, error) {
    stakingStartBalance := new(*big.Int)
    if err := mp.Contract.Call(opts, stakingStartBalance, "getStakingStartBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s staking start balance: %w", mp.Address.Hex(), err)
    }
    return *stakingStartBalance, nil
}
func (mp *Minipool) GetStakingEndBalance(opts *bind.CallOpts) (*big.Int, error) {
    stakingEndBalance := new(*big.Int)
    if err := mp.Contract.Call(opts, stakingEndBalance, "getStakingEndBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool %s staking end balance: %w", mp.Address.Hex(), err)
    }
    return *stakingEndBalance, nil
}
func (mp *Minipool) GetValidatorBalanceWithdrawn(opts *bind.CallOpts) (bool, error) {
    validatorBalanceWithdrawn := new(bool)
    if err := mp.Contract.Call(opts, validatorBalanceWithdrawn, "getValidatorBalanceWithdrawn"); err != nil {
        return false, fmt.Errorf("Could not get minipool %s validator balance withdrawn status: %w", mp.Address.Hex(), err)
    }
    return *validatorBalanceWithdrawn, nil
}


// Get withdrawal credentials
func (mp *Minipool) GetWithdrawalCredentials(opts *bind.CallOpts) (common.Hash, error) {
    withdrawalCredentials := new(common.Hash)
    if err := mp.Contract.Call(opts, withdrawalCredentials, "getWithdrawalCredentials"); err != nil {
        return common.Hash{}, fmt.Errorf("Could not get minipool %s withdrawal credentials: %w", mp.Address.Hex(), err)
    }
    return *withdrawalCredentials, nil
}


// Refund node ETH from the minipool
func (mp *Minipool) Refund(opts *bind.TransactOpts) (common.Hash, error) {
    hash, err := mp.Contract.Transact(opts, "refund")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not refund from minipool %s: %w", mp.Address.Hex(), err)
    }
    return hash, nil
}


// Progress the prelaunch minipool to staking
func (mp *Minipool) Stake(validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (common.Hash, error) {
    hash, err := mp.Contract.Transact(opts, "stake", validatorPubkey[:], validatorSignature[:], depositDataRoot)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err)
    }
    return hash, nil
}


// Withdraw node balances & rewards from the withdrawable minipool and close it
func (mp *Minipool) Withdraw(opts *bind.TransactOpts) (common.Hash, error) {
    hash, err := mp.Contract.Transact(opts, "withdraw")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not withdraw from minipool %s: %w", mp.Address.Hex(), err)
    }
    return hash, nil
}


// Dissolve the initialized or prelaunch minipool
func (mp *Minipool) Dissolve(opts *bind.TransactOpts) (common.Hash, error) {
    hash, err := mp.Contract.Transact(opts, "dissolve")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not dissolve minipool %s: %w", mp.Address.Hex(), err)
    }
    return hash, nil
}


// Withdraw node balances from the dissolved minipool and close it
func (mp *Minipool) Close(opts *bind.TransactOpts) (common.Hash, error) {
    hash, err := mp.Contract.Transact(opts, "close")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not close minipool %s: %w", mp.Address.Hex(), err)
    }
    return hash, nil
}


// Get a minipool contract
var rocketMinipoolLock sync.Mutex
func getMinipoolContract(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*rocketpool.Contract, error) {
    rocketMinipoolLock.Lock()
    defer rocketMinipoolLock.Unlock()
    return rp.MakeContract("rocketMinipool", minipoolAddress)
}

