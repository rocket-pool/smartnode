package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type FilterQuery struct {
	BlockHash *common.Hash
	FromBlock *big.Int
	ToBlock   *big.Int
	Topics    [][]common.Hash
}

func FilterContractLogs(rp *rocketpool.RocketPool, contractName string, q FilterQuery) ([]types.Log, error) {
	rocketDaoNodeTrustedUpgrade, err := rp.GetContract("rocketDAONodeTrustedUpgrade")
	if err != nil {
		return nil, err
	}
	// Get all the addresses this contract has ever been deployed at
	addresses := make([]common.Address, 0)
	// Construct a filter to query ContractUpgraded event
	addressFilter := []common.Address{*rocketDaoNodeTrustedUpgrade.Address}
	topicFilter := [][]common.Hash{{rocketDaoNodeTrustedUpgrade.ABI.Events["ContractUpgraded"].ID}, {crypto.Keccak256Hash([]byte(contractName))}}
	logs, err := rp.Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: addressFilter,
		Topics:    topicFilter,
	})
	if err != nil {
		return nil, err
	}
	// Interate the logs and store every past contract address
	for _, log := range logs {
		addresses = append(addresses, common.HexToAddress(log.Topics[2].Hex()))
	}
	// Append current address
	currentAddress, err := rp.GetAddress(contractName)
	if err != nil {
		return nil, err
	}
	addresses = append(addresses, *currentAddress)
	// Perform the desired getLogs call and return results
	return rp.Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		BlockHash: q.BlockHash,
		Addresses: addresses,
		FromBlock: q.FromBlock,
		ToBlock: q.ToBlock,
		Topics: q.Topics,
	})
}
