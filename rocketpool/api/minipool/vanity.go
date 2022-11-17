package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v110_node "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0/node"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
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
	cfg, err := services.GetConfig(c)
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
	var minipoolBytecode []byte
	minipoolBytecode, err = minipool.GetMinipoolBytecode(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting minipool contract bytecode: %w", err)
	}

	isAtlasDeployed, err := rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}

	var packedConstructorArgs []byte
	if !isAtlasDeployed {
		nodeDepositAddress := cfg.Smartnode.GetV110NodeDepositAddress()

		// Get the deposit type
		depositType, err := v110_node.GetDepositType(rp, depositAmount, nil, &nodeDepositAddress)
		if err != nil {
			return nil, err
		}
		if depositType == types.None {
			return nil, fmt.Errorf("Invalid deposit amount")
		}

		// Create the hash of the minipool constructor call
		depositTypeBytes := [32]byte{}
		depositTypeBytes[0] = byte(depositType)
		packedConstructorArgs, err = minipoolAbi.Pack("", rp.RocketStorageContract.Address, nodeAddress, depositType)
		if err != nil {
			return nil, fmt.Errorf("Error creating minipool constructor args: %w", err)
		}
	} else {
		// Create the hash of the minipool constructor call
		packedConstructorArgs, err = minipoolAbi.Pack("", rp.RocketStorageContract.Address, nodeAddress)
		if err != nil {
			return nil, fmt.Errorf("Error creating minipool constructor args: %w", err)
		}
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
