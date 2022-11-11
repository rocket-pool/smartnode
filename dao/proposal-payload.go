package dao

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	strutils "github.com/rocket-pool/rocketpool-go/utils/strings"
)

// Get the string representation of a proposal payload
var getProposalPayloadStringLock sync.Mutex

func GetProposalPayloadString(rp *rocketpool.RocketPool, daoName string, payload []byte, opts *bind.CallOpts) (string, error) {

	// Lock while getting proposal payload string
	getProposalPayloadStringLock.Lock()
	defer getProposalPayloadStringLock.Unlock()

	// Get proposal DAO contract ABI
	daoContractAbi, err := rp.GetABI(daoName, opts)
	if err != nil {
		return "", fmt.Errorf("Could not get '%s' DAO contract ABI: %w", daoName, err)
	}

	// Get proposal payload method
	method, err := daoContractAbi.MethodById(payload)
	if err != nil {
		return "", fmt.Errorf("Could not get proposal payload method: %w", err)
	}

	// Get proposal payload argument values
	args, err := method.Inputs.UnpackValues(payload[4:])
	if err != nil {
		return "", fmt.Errorf("Could not get proposal payload arguments: %w", err)
	}

	// Format argument values as strings
	argStrs := []string{}
	for ai, arg := range args {
		switch method.Inputs[ai].Type.T {
		case abi.AddressTy:
			argStrs = append(argStrs, arg.(common.Address).Hex())
		case abi.HashTy:
			argStrs = append(argStrs, arg.(common.Hash).Hex())
		case abi.FixedBytesTy:
			fallthrough
		case abi.BytesTy:
			argStrs = append(argStrs, hex.EncodeToString(arg.([]byte)))
		default:
			argStrs = append(argStrs, fmt.Sprintf("%v", arg))
		}
	}

	// Build & return payload string
	return strutils.Sanitize(fmt.Sprintf("%s(%s)", method.RawName, strings.Join(argStrs, ","))), nil

}
