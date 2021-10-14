package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func getVanityArtifacts(c *cli.Context, depositAmount *big.Int) (*api.GetVanityArtifactsResponse, error) {

    // Get services
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.GetVanityArtifactsResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }
    nodeAddress := nodeAccount.Address

    // Get some contract and ABI dependencies
    rocketMinipoolManager, err := rp.GetContract("rocketMinipoolManager")
    if err != nil {
        return nil, fmt.Errorf("Error getting MinipoolManager contract: %w", err)
    }
    minipoolAbi, err := rp.GetABI("rocketMinipool")
    if err != nil {
        return nil, fmt.Errorf("Error getting RocketMinipool ABI: %w", err)
    }
    minipoolBytecode, err := minipool.GetMinipoolBytecode(rp, nil)
    if err != nil {
        return nil, fmt.Errorf("Error getting minipool contract bytecode: %w", err)
    }

    // Get the deposit type
    depositType, err := node.GetDepositType(rp, depositAmount, nil)
    if err != nil {
        return nil, err
    }
    if depositType == types.None {
        return nil, fmt.Errorf("Invalid deposit amount")
    }

    // Create the hash of the minipool constructor call
    depositTypeBytes := [32]byte{}
    depositTypeBytes[0] = byte(depositType)
    packedConstructorArgs, err := minipoolAbi.Pack("", rp.RocketStorageContract.Address, nodeAddress, depositType)
    if err != nil {
        return nil, fmt.Errorf("Error creating minipool constructor args: %w", err)
    }

    // Get the initialization data hash
    initData := append(minipoolBytecode, packedConstructorArgs...)
    initHash := crypto.Keccak256Hash(initData)

    // Update & return response
    response.NodeAddress = nodeAddress
    response.MinipoolManagerAddress = *rocketMinipoolManager.Address
    response.InitHash = initHash
    return &response, nil

}

