package eth1

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// Determines if the primary EC can be used for historical queries, or if the Archive EC is required
func GetBestApiClient(primary *rocketpool.RocketPool, cfg *config.RocketPoolConfig, printMessage func(string), blockNumber *big.Int) (*rocketpool.RocketPool, error) {

	client := primary

	// Try getting the rETH address as a canary to see if the block is available
	opts := &bind.CallOpts{
		BlockNumber: blockNumber,
	}
	address, err := client.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.addressrocketTokenRETH")))
	if err != nil {
		errMessage := err.Error()
		printMessage(fmt.Sprintf("Error getting state for block %d: %s", blockNumber.Uint64(), errMessage))
		// The state was missing so fall back to the archive node
		archiveEcUrl := cfg.Smartnode.ArchiveECUrl.Value.(string)
		if archiveEcUrl != "" {
			printMessage(fmt.Sprintf("Primary EC cannot retrieve state for historical block %d, using archive EC [%s]", blockNumber.Uint64(), archiveEcUrl))
			ec, err := services.NewEthClient(archiveEcUrl)
			if err != nil {
				return nil, fmt.Errorf("Error connecting to archive EC: %w", err)
			}
			client, err = rocketpool.NewRocketPool(ec, common.HexToAddress(cfg.Smartnode.GetStorageAddress()))
			if err != nil {
				return nil, fmt.Errorf("Error creating Rocket Pool client connected to archive EC: %w", err)
			}

			// Get the rETH address from the archive EC
			address, err = client.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.addressrocketTokenRETH")))
			if err != nil {
				return nil, fmt.Errorf("Error verifying rETH address with Archive EC: %w", err)
			}
		} else {
			// No archive node specified
			return nil, fmt.Errorf("***ERROR*** Primary EC cannot retrieve state for historical block %d and the Archive EC is not specified.", blockNumber.Uint64())
		}
	}

	// Sanity check the rETH address to make sure the client is working right
	if address != cfg.Smartnode.GetRethAddress() {
		return nil, fmt.Errorf("***ERROR*** Your Primary EC provided %s as the rETH address, but it should have been %s!", address.Hex(), cfg.Smartnode.GetRethAddress().Hex())
	}

	return client, nil

}
