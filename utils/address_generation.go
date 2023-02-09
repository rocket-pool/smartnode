package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Combine a node's address and a salt to retreive a new salt compatible with depositing
func GetNodeSalt(nodeAddress common.Address, salt *big.Int) common.Hash {
	// Create a new salt by hashing the original and the node address
	saltBytes := [32]byte{}
	salt.FillBytes(saltBytes[:])
	saltHash := crypto.Keccak256Hash(nodeAddress.Bytes(), saltBytes[:])
	return saltHash
}
