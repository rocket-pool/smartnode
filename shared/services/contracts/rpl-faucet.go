package contracts

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	batch "github.com/rocket-pool/batch-query"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
)

const (
	faucetAbiString string = "[{\"constant\":true,\"inputs\":[],\"name\":\"maxWithdrawalPerPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_rplTokenAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"created\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdrawTo\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"getAllowanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getWithdrawalPeriodStart\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_withdrawalPeriod\",\"type\":\"uint256\"}],\"name\":\"setWithdrawalPeriod\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_maxWithdrawalPerPeriod\",\"type\":\"uint256\"}],\"name\":\"setMaxWithdrawalPerPeriod\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_withdrawalFee\",\"type\":\"uint256\"}],\"name\":\"setWithdrawalFee\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

// ABI cache
var faucetAbi abi.ABI
var mcOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for RPL Faucet
type RplFaucet struct {
	Details  RplFaucetDetails
	contract *core.Contract
}

// Details for RPL Faucet
type RplFaucetDetails struct {
	MaxWithdrawalPerPeriod *big.Int               `json:"maxWithdrawalPerPeriod"`
	WithdrawalFee          *big.Int               `json:"withdrawalFee"`
	Owner                  common.Address         `json:"owner"`
	WithdrawalPeriod       core.Parameter[uint64] `json:"withdrawalPeriod"`
	Balance                *big.Int               `json:"balance"`
	Allowance              *big.Int               `json:"allowance"`
	WithdrawalPeriodStart  core.Parameter[uint64] `json:"getWithdrawalPeriodStart"`

	AllottedRplBalance  *big.Int               `json:"allottedRplBalance"`
	RemainingRplBalance *big.Int               `json:"remainingRplBalance"`
	LotCount            core.Parameter[uint64] `json:"lotCount"`
}

// ====================
// === Constructors ===
// ====================

// Creates a new RPL Faucet contract binding
func NewRplFaucet(address common.Address, client core.ExecutionClient) (*RplFaucet, error) {
	// Parse the ABI
	var err error
	mcOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(faucetAbiString))
		if err == nil {
			faucetAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing faucet ABI: %w", err)
	}

	// Create the contract
	contract := &core.Contract{
		Contract: bind.NewBoundContract(address, faucetAbi, client, client, client),
		Address:  &address,
		ABI:      &faucetAbi,
		Client:   client,
	}

	return &RplFaucet{
		contract: contract,
	}, nil
}

// =============
// === Calls ===
// =============

// Get the max amount of RPL that can be withdrawn (in total) per withdrawal period
func (c *RplFaucet) GetMaxWithdrawalPerPeriod(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.MaxWithdrawalPerPeriod, "maxWithdrawalPerPeriod")
}

// Get the withdrawal fee, in ETH, for pulling RPL out of the faucet
func (c *RplFaucet) GetWithdrawalFee(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.WithdrawalFee, "withdrawalFee")
}

// Get the owner of the faucet that can perform administrative duties
func (c *RplFaucet) GetOwner(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.Owner, "owner")
}

// Get the length of a withdrawal period before it resets, in blocks
func (c *RplFaucet) GetWithdrawalPeriod(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.WithdrawalPeriod.RawValue, "withdrawalPeriod")
}

// Get the remaining RPL balance of the faucet
func (c *RplFaucet) GetBalance(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.Balance, "getBalance")
}

// Get the amount of RPL that can be withdrawn per address in each withdrawal period
func (c *RplFaucet) GetAllowance(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.Allowance, "getAllowance")
}

// Get the amount of RPL that can be withdrawn by the given address in the current withdrawal period
func (c *RplFaucet) GetAllowanceFor(mc *batch.MultiCaller, allowance **big.Int, address common.Address) {
	core.AddCall(mc, c.contract, allowance, "getAllowanceFor", address)
}

// Get the block number the current withdrawal period started
func (c *RplFaucet) GetWithdrawalPeriodStart(mc *batch.MultiCaller) {
	core.AddCall(mc, c.contract, &c.Details.WithdrawalPeriodStart.RawValue, "getWithdrawalPeriodStart")
}

// Get all basic details
func (c *RplFaucet) GetAllDetails(mc *batch.MultiCaller) {
	c.GetMaxWithdrawalPerPeriod(mc)
	c.GetWithdrawalFee(mc)
	c.GetOwner(mc)
	c.GetWithdrawalPeriod(mc)
	c.GetBalance(mc)
	c.GetAllowance(mc)
	c.GetWithdrawalPeriodStart(mc)
}

// ====================
// === Transactions ===
// ====================

// Get info for withdrawing RPL from the faucet
func (c *RplFaucet) Withdraw(opts *bind.TransactOpts, amount *big.Int) (*core.TransactionInfo, error) {
	return core.NewTransactionInfo(c.contract, "withdraw", opts, amount)
}

// Get info for withdrawing RPL from the faucet to a specific address
func (c *RplFaucet) WithdrawTo(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*core.TransactionInfo, error) {
	return core.NewTransactionInfo(c.contract, "withdrawTo", opts, to, amount)
}

// Set the withdrawal period, in blocks
func (c *RplFaucet) SetWithdrawalPeriod(opts *bind.TransactOpts, period core.Parameter[uint64]) (*core.TransactionInfo, error) {
	return core.NewTransactionInfo(c.contract, "setWithdrawalPeriod", opts, period.RawValue)
}

// Set the max total withdrawal amount per period
func (c *RplFaucet) SetMaxWithdrawalPerPeriod(opts *bind.TransactOpts, max *big.Int) (*core.TransactionInfo, error) {
	return core.NewTransactionInfo(c.contract, "setMaxWithdrawalPerPeriod", opts, max)
}

// Set the withdrawal fee
func (c *RplFaucet) SetWithdrawalFee(opts *bind.TransactOpts, fee *big.Int) (*core.TransactionInfo, error) {
	return core.NewTransactionInfo(c.contract, "setWithdrawalFee", opts, fee)
}
