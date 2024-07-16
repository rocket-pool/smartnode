package contracts

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
)

const (
	rocketSignerRegistryAbiString string = "[{\"type\":\"function\",\"name\":\"clearSigner\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"nodeToSigner\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setSigner\",\"inputs\":[{\"name\":\"_signer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"_r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"signerToNode\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"SignerSet\",\"inputs\":[{\"name\":\"nodeAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"signerAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"StringsInsufficientHexLength\",\"inputs\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]"
)

// ABI cache
var rocketSignerRegistryAbi abi.ABI
var rocketSignerRegistryOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for Rocket Signer Registry
type RocketSignerRegistry struct {
	contract *eth.Contract
	txMgr    *eth.TransactionManager
}

// ====================
// === Constructors ===
// ====================

// Creates a new Rocket Signer Registry contract binding
func NewRocketSignerRegistry(address common.Address, client eth.IExecutionClient, txMgr *eth.TransactionManager) (*RocketSignerRegistry, error) {
	// Parse the ABI
	var err error
	rocketSignerRegistryOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(rocketSignerRegistryAbiString))
		if err == nil {
			rocketSignerRegistryAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing rocket signer registry ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, rocketSignerRegistryAbi, client, client, client),
		Address:      address,
		ABI:          &rocketSignerRegistryAbi,
	}

	return &RocketSignerRegistry{
		contract: contract,
	}, nil
}

// =============
// === Calls ===
// =============

// Get the delegate for the provided address
func (c *RocketSignerRegistry) NodeToSigner(mc *batch.MultiCaller, out *common.Address, address common.Address) {
	eth.AddCallToMulticaller(mc, c.contract, out, "nodeToSigner", address)
}
func (c *RocketSignerRegistry) SignerToNode(mc *batch.MultiCaller, out *common.Address, address common.Address) {
	eth.AddCallToMulticaller(mc, c.contract, out, "signerToNode", address)
}

// ====================
// === Transactions ===
// ====================

// Get info for setting the signalling address
func (c *RocketSignerRegistry) SetSigner(id common.Hash, _signer common.Address, opts *bind.TransactOpts, _v uint8, _r [32]byte, _s [32]byte) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "setSigner", opts, _signer, _v, _r, _s)
}

// Get info for clearing the signalling address
func (c *RocketSignerRegistry) ClearSigner(opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "clearSigner", opts)
}
