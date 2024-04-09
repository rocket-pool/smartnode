package eth1

import (
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Determines if the primary EC can be used for historical queries, or if the Archive EC is required
func GetBestApiClient(primary *rocketpool.RocketPool, cfg *config.SmartNodeConfig, logger *slog.Logger, blockNumber *big.Int) (*rocketpool.RocketPool, error) {
	client := primary
	resources := cfg.GetRocketPoolResources()

	// Try getting the rETH address as a canary to see if the block is available
	opts := &bind.CallOpts{
		BlockNumber: blockNumber,
	}
	var address common.Address
	err := client.Query(func(mc *batch.MultiCaller) error {
		client.Storage.GetAddress(mc, &address, string(rocketpool.ContractName_RocketTokenRETH))
		return nil
	}, opts)
	if err != nil {
		errMessage := err.Error()
		logger.Error("Error getting rETH address", slog.Uint64(keys.BlockKey, blockNumber.Uint64()), slog.String(log.ErrorKey, errMessage))
		if strings.Contains(errMessage, "missing trie node") || // Geth
			strings.Contains(errMessage, "No state available for block") || // Nethermind
			strings.Contains(errMessage, "Internal error") { // Besu

			// The state was missing so fall back to the archive node
			archiveEcUrl := cfg.ArchiveEcUrl.Value
			if archiveEcUrl != "" {
				logger.Info("Primary EC cannot retrieve state for historical block, using archive EC", slog.Uint64(keys.BlockKey, blockNumber.Uint64()), slog.String(keys.ArchiveEcKey, archiveEcUrl))
				ec, err := ethclient.Dial(archiveEcUrl)
				if err != nil {
					return nil, fmt.Errorf("error connecting to archive EC: %w", err)
				}
				client, err = rocketpool.NewRocketPool(
					ec,
					resources.StorageAddress,
					resources.MulticallAddress,
					resources.BalanceBatcherAddress,
				)
				if err != nil {
					return nil, fmt.Errorf("error creating Rocket Pool client connected to archive EC: %w", err)
				}

				// Get the rETH address from the archive EC
				err = client.Query(func(mc *batch.MultiCaller) error {
					client.Storage.GetAddress(mc, &address, string(rocketpool.ContractName_RocketTokenRETH))
					return nil
				}, opts)
				if err != nil {
					return nil, fmt.Errorf("error verifying rETH address with Archive EC: %w", err)
				}
			} else {
				// No archive node specified
				return nil, fmt.Errorf("***ERROR*** Primary EC cannot retrieve state for historical block %d and the Archive EC is not specified", blockNumber.Uint64())
			}

		}
	}

	// Sanity check the rETH address to make sure the client is working right
	if address != resources.RethAddress {
		return nil, fmt.Errorf("***ERROR*** Your Primary EC provided %s as the rETH address, but it should have been %s", address.Hex(), resources.RethAddress.Hex())
	}

	return client, nil
}
