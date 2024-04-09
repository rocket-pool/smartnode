package contracts

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/v2/core"
)

const (
	faucetAbiString string = "[{\"constant\":true,\"inputs\":[],\"name\":\"maxWithdrawalPerPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_rplTokenAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"created\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdrawTo\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"getAllowanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getWithdrawalPeriodStart\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_withdrawalPeriod\",\"type\":\"uint256\"}],\"name\":\"setWithdrawalPeriod\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_maxWithdrawalPerPeriod\",\"type\":\"uint256\"}],\"name\":\"setMaxWithdrawalPerPeriod\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_withdrawalFee\",\"type\":\"uint256\"}],\"name\":\"setWithdrawalFee\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

// ABI cache
var faucetAbi abi.ABI
var faucetOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for RPL Faucet
type RplFaucet struct {
	// The max amount of RPL that can be withdrawn (in total) per withdrawal period
	MaxWithdrawalPerPeriod *core.SimpleField[*big.Int]

	// The withdrawal fee, in ETH, for pulling RPL out of the faucet
	WithdrawalFee *core.SimpleField[*big.Int]

	// The owner of the faucet that can perform administrative duties
	Owner *core.SimpleField[common.Address]

	// The length of a withdrawal period before it resets, in blocks
	WithdrawalPeriod *core.FormattedUint256Field[uint64]

	// The remaining RPL balance of the faucet
	Balance *core.SimpleField[*big.Int]

	// The amount of RPL that can be withdrawn per address in each withdrawal period
	Allowance *core.SimpleField[*big.Int]

	// The block number the current withdrawal period started
	WithdrawalPeriodStart *core.FormattedUint256Field[uint64]

	// === Internal fields ===
	contract *core.Contract
	txMgr    *eth.TransactionManager
}

// ====================
// === Constructors ===
// ====================

// Creates a new RPL Faucet contract binding
func NewRplFaucet(address common.Address, client eth.IExecutionClient, txMgr *eth.TransactionManager) (*RplFaucet, error) {
	// Parse the ABI
	var err error
	faucetOnce.Do(func() {
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
		Contract: &eth.Contract{
			ContractImpl: bind.NewBoundContract(address, faucetAbi, client, client, client),
			Address:      address,
			ABI:          &faucetAbi,
		},
	}

	return &RplFaucet{
		MaxWithdrawalPerPeriod: core.NewSimpleField[*big.Int](contract, "maxWithdrawalPerPeriod"),
		WithdrawalFee:          core.NewSimpleField[*big.Int](contract, "withdrawalFee"),
		Owner:                  core.NewSimpleField[common.Address](contract, "owner"),
		WithdrawalPeriod:       core.NewFormattedUint256Field[uint64](contract, "withdrawalPeriod"),
		Balance:                core.NewSimpleField[*big.Int](contract, "getBalance"),
		Allowance:              core.NewSimpleField[*big.Int](contract, "getAllowance"),
		WithdrawalPeriodStart:  core.NewFormattedUint256Field[uint64](contract, "getWithdrawalPeriodStart"),

		contract: contract,
		txMgr:    txMgr,
	}, nil
}

// =============
// === Calls ===
// =============

// Get the amount of RPL that can be withdrawn by the given address in the current withdrawal period
func (c *RplFaucet) GetAllowanceFor(mc *batch.MultiCaller, allowance **big.Int, address common.Address) {
	core.AddCall(mc, c.contract, allowance, "getAllowanceFor", address)
}

// ====================
// === Transactions ===
// ====================

// Get info for withdrawing RPL from the faucet
func (c *RplFaucet) Withdraw(amount *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract.Contract, "withdraw", opts, amount)
}

// Get info for withdrawing RPL from the faucet to a specific address
func (c *RplFaucet) WithdrawTo(to common.Address, amount *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract.Contract, "withdrawTo", opts, to, amount)
}

// Set the withdrawal period, in blocks
func (c *RplFaucet) SetWithdrawalPeriod(period *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract.Contract, "setWithdrawalPeriod", opts, period)
}

// Set the max total withdrawal amount per period
func (c *RplFaucet) SetMaxWithdrawalPerPeriod(max *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract.Contract, "setMaxWithdrawalPerPeriod", opts, max)
}

// Set the withdrawal fee
func (c *RplFaucet) SetWithdrawalFee(fee *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract.Contract, "setWithdrawalFee", opts, fee)
}
