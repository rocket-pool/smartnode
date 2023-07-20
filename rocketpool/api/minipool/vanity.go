package minipool

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
)

const (
	ozMinipoolBytecode     string = "0x3d602d80600a3d3981f3363d3d373d3d3d363d73%s5af43d82803e903d91602b57fd5bf3"
	rocketMinipoolBaseName string = "rocketMinipoolBase"
)

func getVanityArtifacts(c *cli.Context, depositAmount *big.Int, nodeAddressStr string) (*api.GetVanityArtifactsResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetVanityArtifactsResponse{}

	// Get node account
	var nodeAddress common.Address
	if nodeAddressStr == "0" {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		nodeAddress = nodeAccount.Address
	} else {
		if common.IsHexAddress(nodeAddressStr) {
			nodeAddress = common.HexToAddress(nodeAddressStr)
		} else {
			return nil, fmt.Errorf("%s is not a valid node address", nodeAddressStr)
		}
	}

	// Get some contract and ABI dependencies
	rocketMinipoolFactory, err := rp.GetContract("rocketMinipoolFactory", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting MinipoolFactory contract: %w", err)
	}
	minipoolAbi, err := rp.GetABI("rocketMinipool", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting RocketMinipool ABI: %w", err)
	}

	// Get the address of rocketMinipoolBase
	rocketMinipoolBaseAddress, err := rp.GetAddress(rocketMinipoolBaseName, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool base address: %w", err)
	}
	bytecodeString := fmt.Sprintf(ozMinipoolBytecode, hexutils.RemovePrefix(rocketMinipoolBaseAddress.Hex()))
	bytecodeString = hexutils.RemovePrefix(bytecodeString)
	minipoolBytecode, err := hex.DecodeString(bytecodeString)
	if err != nil {
		return nil, fmt.Errorf("error decoding minipool bytecode [%s]: %w", bytecodeString, err)
	}

	// Create the hash of the minipool constructor call
	packedConstructorArgs, err := minipoolAbi.Pack("")
	if err != nil {
		return nil, fmt.Errorf("Error creating minipool constructor args: %w", err)
	}

	// Get the initialization data hash
	initData := append(minipoolBytecode, packedConstructorArgs...)
	initHash := crypto.Keccak256Hash(initData)

	// Update & return response
	response.NodeAddress = nodeAddress
	response.MinipoolFactoryAddress = *rocketMinipoolFactory.Address
	response.InitHash = initHash
	return &response, nil

}
