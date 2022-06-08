package minipool

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// The number of blocks to look for events in at once when scanning
const EventScanInterval = 10000

// Minipool detail types
type StatusDetails struct {
	Status      rptypes.MinipoolStatus `json:"status"`
	StatusBlock uint64                 `json:"statusBlock"`
	StatusTime  time.Time              `json:"statusTime"`
}
type NodeDetails struct {
	Address         common.Address `json:"address"`
	Fee             float64        `json:"fee"`
	DepositBalance  *big.Int       `json:"depositBalance"`
	RefundBalance   *big.Int       `json:"refundBalance"`
	DepositAssigned bool           `json:"depositAssigned"`
}
type UserDetails struct {
	DepositBalance      *big.Int  `json:"depositBalance"`
	DepositAssigned     bool      `json:"depositAssigned"`
	DepositAssignedTime time.Time `json:"depositAssignedTime"`
}

// The data from a minipool's MinipoolPrestaked event
type minipoolPrestakeEvent struct {
	Pubkey                []byte   `abi:"validatorPubkey"`
	Signature             []byte   `abi:"validatorSignature"`
	DepositDataRoot       [32]byte `abi:"depositDataRoot"`
	Amount                *big.Int `abi:"amount"`
	WithdrawalCredentials []byte   `abi:"withdrawalCredentials"`
	Time                  *big.Int `abi:"time"`
}

// Formatted MinipoolPrestaked event data
type PrestakeData struct {
	Pubkey                rptypes.ValidatorPubkey    `json:"pubkey"`
	WithdrawalCredentials common.Hash                `json:"withdrawalCredentials"`
	Amount                *big.Int                   `json:"amount"`
	Signature             rptypes.ValidatorSignature `json:"signature"`
	DepositDataRoot       common.Hash                `json:"depositDataRoot"`
	Time                  time.Time                  `json:"time"`
}

// Minipool contract
type Minipool struct {
	Address    common.Address
	Contract   *rocketpool.Contract
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
		Address:    address,
		Contract:   contract,
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
		Status:      status,
		StatusBlock: statusBlock,
		StatusTime:  statusTime,
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
func (mp *Minipool) GetFinalised(opts *bind.CallOpts) (bool, error) {
	finalised := new(bool)
	if err := mp.Contract.Call(opts, finalised, "getFinalised"); err != nil {
		return false, fmt.Errorf("Could not get minipool %s finalised: %w", mp.Address.Hex(), err)
	}
	return *finalised, nil
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

	// Wait for data
	if err := wg.Wait(); err != nil {
		return NodeDetails{}, err
	}

	// Return
	return NodeDetails{
		Address:         address,
		Fee:             fee,
		DepositBalance:  depositBalance,
		RefundBalance:   refundBalance,
		DepositAssigned: depositAssigned,
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
func (mp *Minipool) GetNodeFeeRaw(opts *bind.CallOpts) (*big.Int, error) {
	nodeFee := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeFee, "getNodeFee"); err != nil {
		return nil, fmt.Errorf("Could not get minipool %s node fee: %w", mp.Address.Hex(), err)
	}
	return *nodeFee, nil
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
		DepositBalance:      depositBalance,
		DepositAssigned:     depositAssigned,
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

// Estimate the gas of Refund
func (mp *Minipool) EstimateRefundGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "refund")
}

// Refund node ETH from the minipool
func (mp *Minipool) Refund(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "refund")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not refund from minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of DistributeBalance
func (mp *Minipool) EstimateDistributeBalanceGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "distributeBalance")
}

// Distribute the minipool's ETH balance to the node operator and rETH staking pool.
// !!! WARNING !!!
// DO NOT CALL THIS until the minipool's validator has exited from the Beacon Chain
// and the balance has been deposited into the minipool!
func (mp *Minipool) DistributeBalance(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "distributeBalance")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not process withdrawal for minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of DistributeBalanceAndFinalise
func (mp *Minipool) EstimateDistributeBalanceAndFinaliseGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "distributeBalanceAndFinalise")
}

// Distribute the minipool's ETH balance to the node operator and rETH staking pool,
// then finalises the minipool
// !!! WARNING !!!
// DO NOT CALL THIS until the minipool's validator has exited from the Beacon Chain
// and the balance has been deposited into the minipool!
func (mp *Minipool) DistributeBalanceAndFinalise(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "distributeBalanceAndFinalise")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not process withdrawal for and finalise minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of Stake
func (mp *Minipool) EstimateStakeGas(validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "stake", validatorSignature[:], depositDataRoot)
}

// Progress the prelaunch minipool to staking
func (mp *Minipool) Stake(validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "stake", validatorSignature[:], depositDataRoot)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of Dissolve
func (mp *Minipool) EstimateDissolveGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "dissolve")
}

// Dissolve the initialized or prelaunch minipool
func (mp *Minipool) Dissolve(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "dissolve")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not dissolve minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of Close
func (mp *Minipool) EstimateCloseGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "close")
}

// Withdraw node balances from the dissolved minipool and close it
func (mp *Minipool) Close(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "close")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not close minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of Finalise
func (mp *Minipool) EstimateFinaliseGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "finalise")
}

// Finalise a minipool to get the RPL stake back
func (mp *Minipool) Finalise(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "finalise")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not finalise minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of DelegateUpgrade
func (mp *Minipool) EstimateDelegateUpgradeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "delegateUpgrade")
}

// Upgrade this minipool to the latest network delegate contract
func (mp *Minipool) DelegateUpgrade(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "delegateUpgrade")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not upgrade delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of DelegateRollback
func (mp *Minipool) EstimateDelegateRollbackGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "delegateRollback")
}

// Rollback to previous delegate contract
func (mp *Minipool) DelegateRollback(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "delegateRollback")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not rollback delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Estimate the gas of SetUseLatestDelegate
func (mp *Minipool) EstimateSetUseLatestDelegateGas(setting bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "setUseLatestDelegate", setting)
}

// If set to true, will automatically use the latest delegate contract
func (mp *Minipool) SetUseLatestDelegate(setting bool, opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "setUseLatestDelegate", setting)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not set use latest delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Getter for useLatestDelegate setting
func (mp *Minipool) GetUseLatestDelegate(opts *bind.CallOpts) (bool, error) {
	setting := new(bool)
	if err := mp.Contract.Call(opts, setting, "getUseLatestDelegate"); err != nil {
		return false, fmt.Errorf("Could not get use latest delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *setting, nil
}

// Returns the address of the minipool's stored delegate
func (mp *Minipool) GetDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Returns the address of the minipool's previous delegate (or address(0) if not set)
func (mp *Minipool) GetPreviousDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getPreviousDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get previous delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Returns the delegate which will be used when calling this minipool taking into account useLatestDelegate setting
func (mp *Minipool) GetEffectiveDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getEffectiveDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get effective delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Given a validator balance, calculates how much belongs to the node taking into consideration rewards and penalties
func (mp *Minipool) CalculateNodeShare(balance *big.Int, opts *bind.CallOpts) (*big.Int, error) {
	nodeAmount := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeAmount, "calculateNodeShare", balance); err != nil {
		return nil, fmt.Errorf("Could not get minipool node portion: %w", err)
	}
	return *nodeAmount, nil
}

// Given a validator balance, calculates how much belongs to rETH users taking into consideration rewards and penalties
func (mp *Minipool) CalculateUserShare(balance *big.Int, opts *bind.CallOpts) (*big.Int, error) {
	userAmount := new(*big.Int)
	if err := mp.Contract.Call(opts, userAmount, "calculateUserShare", balance); err != nil {
		return nil, fmt.Errorf("Could not get minipool user portion: %w", err)
	}
	return *userAmount, nil
}

// Estimate the gas requiired to vote to scrub a minipool
func (mp *Minipool) EstimateVoteScrubGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "voteScrub")
}

// Vote to scrub a minipool
func (mp *Minipool) VoteScrub(opts *bind.TransactOpts) (common.Hash, error) {
	hash, err := mp.Contract.Transact(opts, "voteScrub")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not vote to scrub minipool %s: %w", mp.Address.Hex(), err)
	}
	return hash, nil
}

// Get the data from this minipool's MinipoolPrestaked event
func (mp *Minipool) GetPrestakeEvent(intervalSize *big.Int, opts *bind.CallOpts) (PrestakeData, error) {

	addressFilter := []common.Address{mp.Address}
	topicFilter := [][]common.Hash{{mp.Contract.ABI.Events["MinipoolPrestaked"].ID}}

	// Grab the latest block number
	currentBlock, err := mp.RocketPool.Client.BlockNumber(context.Background())
	if err != nil {
		return PrestakeData{}, fmt.Errorf("Error getting current block %s: %w", mp.Address.Hex(), err)
	}

	// Grab the lowest block number worth querying from (should never have to go back this far in practice)
	deployBlockHash := crypto.Keccak256Hash([]byte("deploy.block"))
	fromBlockBig, err := mp.RocketPool.RocketStorage.GetUint(nil, deployBlockHash)
	if err != nil {
		return PrestakeData{}, fmt.Errorf("Error getting deploy block %s: %w", mp.Address.Hex(), err)
	}

	fromBlock := fromBlockBig.Uint64()
	var log types.Log
	found := false

	// Backwards scan through blocks to find the event
	for i := currentBlock; i >= fromBlock; i -= EventScanInterval {
		from := i - EventScanInterval + 1
		if from < fromBlock {
			from = fromBlock
		}

		fromBig := big.NewInt(0).SetUint64(from)
		toBig := big.NewInt(0).SetUint64(i)

		logs, err := eth.GetLogs(mp.RocketPool, addressFilter, topicFilter, intervalSize, fromBig, toBig, nil)
		if err != nil {
			return PrestakeData{}, fmt.Errorf("Error getting prestake logs for minipool %s: %w", mp.Address.Hex(), err)
		}

		if len(logs) > 0 {
			log = logs[0]
			found = true
			break
		}
	}

	if !found {
		// This should never happen
		return PrestakeData{}, fmt.Errorf("Error finding prestake log for minipool %s", mp.Address.Hex())
	}

	// Decode the event
	prestakeEvent := new(minipoolPrestakeEvent)
	mp.Contract.Contract.UnpackLog(prestakeEvent, "MinipoolPrestaked", log)
	if err != nil {
		return PrestakeData{}, fmt.Errorf("Error unpacking prestake data: %w", err)
	}

	// Convert the event to a more useable struct
	prestakeData := PrestakeData{
		Pubkey:                rptypes.BytesToValidatorPubkey(prestakeEvent.Pubkey),
		WithdrawalCredentials: common.BytesToHash(prestakeEvent.WithdrawalCredentials),
		Amount:                prestakeEvent.Amount,
		Signature:             rptypes.BytesToValidatorSignature(prestakeEvent.Signature),
		DepositDataRoot:       prestakeEvent.DepositDataRoot,
		Time:                  time.Unix(prestakeEvent.Time.Int64(), 0),
	}
	return prestakeData, nil
}

// Get a minipool contract
var rocketMinipoolLock sync.Mutex

func getMinipoolContract(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*rocketpool.Contract, error) {
	rocketMinipoolLock.Lock()
	defer rocketMinipoolLock.Unlock()
	return rp.MakeContract("rocketMinipool", minipoolAddress)
}
