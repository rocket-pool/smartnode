package utils

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/legacy/v1.0.0/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
)

// Combine a node's address and a salt to retreive a new salt compatible with depositing
func GetNodeSalt(nodeAddress common.Address, salt *big.Int) common.Hash {
	// Create a new salt by hashing the original and the node address
	saltBytes := [32]byte{}
	salt.FillBytes(saltBytes[:])
	saltHash := crypto.Keccak256Hash(nodeAddress.Bytes(), saltBytes[:])
	return saltHash
}

// Precompute the address of a minipool based on the node wallet, deposit type, and unique salt
// If you set minipoolBytecode to nil, this will retrieve it from the contracts using minipool.GetMinipoolBytecode().
func GenerateAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, depositType rptypes.MinipoolDeposit, salt *big.Int, minipoolBytecode []byte, legacyRocketMinipoolManagerAddress *common.Address, opts *bind.CallOpts) (common.Address, error) {

	// Get dependencies
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, legacyRocketMinipoolManagerAddress, opts)
	if err != nil {
		return common.Address{}, err
	}
	minipoolAbi, err := rp.GetABI("rocketMinipool", nil)
	if err != nil {
		return common.Address{}, err
	}

	if len(minipoolBytecode) == 0 {
		minipoolBytecode, err = minipool.GetMinipoolBytecode(rp, nil, legacyRocketMinipoolManagerAddress)
		if err != nil {
			return common.Address{}, fmt.Errorf("Error getting minipool bytecode: %w", err)
		}
	}

	// Create the hash of the minipool constructor call
	depositTypeBytes := [32]byte{}
	depositTypeBytes[0] = byte(depositType)
	packedConstructorArgs, err := minipoolAbi.Pack("", rp.RocketStorageContract.Address, nodeAddress, depositType)
	if err != nil {
		return common.Address{}, fmt.Errorf("Error creating minipool constructor args: %w", err)
	}

	// Get the node salt and initialization data
	nodeSalt := GetNodeSalt(nodeAddress, salt)
	initData := append(minipoolBytecode, packedConstructorArgs...)
	initHash := crypto.Keccak256(initData)

	address := crypto.CreateAddress2(*rocketMinipoolManager.Address, nodeSalt, initHash)
	return address, nil

}

// Get contracts
var rocketMinipoolManagerLock sync.Mutex

func getRocketMinipoolManager(rp *rocketpool.RocketPool, address *common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolManagerLock.Lock()
	defer rocketMinipoolManagerLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketMinipoolManager", opts)
	} else {
		return rp.VersionManager.V1_0_0.GetContractWithAddress("rocketMinipoolManager", *address)
	}
}
