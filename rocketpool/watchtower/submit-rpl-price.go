package watchtower

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	mathutils "github.com/rocket-pool/smartnode/shared/utils/math"
)

// Submit RPL price task
type submitRplPrice struct {
    c *cli.Context
    log log.ColorLogger
    cfg config.RocketPoolConfig
    w *wallet.Wallet
    mnec *ethclient.Client
    rp *rocketpool.RocketPool
    oio *contracts.OneInchOracle
}


// Create submit RPL price task
func newSubmitRplPrice(c *cli.Context, logger log.ColorLogger) (*submitRplPrice, error) {

    // Get services
    cfg, err := services.GetConfig(c)
    if err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    mnec, err := services.GetMainnetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    oio, err := services.GetOneInchOracle(c)
    if err != nil { return nil, err }

    // Return task
    return &submitRplPrice{
        c: c,
        log: logger,
        cfg: cfg,
        w: w,
        mnec: mnec,
        rp: rp,
        oio: oio,
    }, nil

}


// Submit RPL price
func (t *submitRplPrice) run() error {

    // Wait for eth client to sync
    if err := services.WaitEthClientSynced(t.c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := t.w.GetNodeAccount()
    if err != nil {
        return err
    }

    // Data
    var wg errgroup.Group
    var nodeTrusted bool
    var submitPricesEnabled bool

    // Get data
    wg.Go(func() error {
        var err error
        nodeTrusted, err = trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        submitPricesEnabled, err = protocol.GetSubmitPricesEnabled(t.rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    // Check node trusted status & settings
    if !(nodeTrusted && submitPricesEnabled) {
        return nil
    }

    // Log
    t.log.Println("Checking for RPL price checkpoint...")

    // Get block to submit price for
    blockNumber, err := t.getLatestReportableBlock()
    if err != nil {
        return err
    }

    // Check if price for block can be submitted by node
    canSubmit, err := t.canSubmitBlockPrice(nodeAccount.Address, blockNumber)
    if err != nil {
        return err
    }
    if !canSubmit {
        return nil
    }

    // Log
    t.log.Printlnf("Getting RPL price for block %d...", blockNumber)

    // Get RPL price at block
    rplPrice, err := t.getRplPrice(blockNumber)
    if err != nil {
        return err
    }

    // Calculate the total effective RPL stake on the network
    zero := new(big.Int).SetUint64(0)
    effectiveRplStake, err := node.CalculateTotalEffectiveRPLStake(t.rp, zero, zero, rplPrice, nil)
    if err != nil {
        return fmt.Errorf("Error getting total effective RPL stake: %w", err)
    }

    // Log
    t.log.Printlnf("RPL price: %.6f ETH", mathutils.RoundDown(eth.WeiToEth(rplPrice), 6))

    // Submit RPL price
    if err := t.submitRplPrice(blockNumber, rplPrice, effectiveRplStake); err != nil {
        return fmt.Errorf("Could not submit RPL price: %w", err)
    }

    // Return
    return nil

}


// Get the latest block number to report RPL price for
func (t *submitRplPrice) getLatestReportableBlock() (uint64, error) {

    // Require mainnet eth client synced
    if err := services.RequireMainnetEthClientSynced(t.c); err != nil {
        return 0, err
    }

    /*
    // Data
    var wg errgroup.Group
    var currentMainnetBlock uint64
    var submitPricesFrequency uint64

    // Get current mainnet block
    wg.Go(func() error {
        header, err := t.mnec.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentMainnetBlock = header.Number.Uint64()
        }
        return err
    })

    // Get price submission frequency
    wg.Go(func() error {
        var err error
        submitPricesFrequency, err = protocol.GetSubmitPricesFrequency(t.rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return 0, err
    }
    
    // Calculate and return
    return (currentMainnetBlock / submitPricesFrequency) * submitPricesFrequency, nil
    */

    latestBlock, err := network.GetLatestReportablePricesBlock(t.rp, nil)
    if err != nil {
        return 0, fmt.Errorf("Error getting latest reportable block: %w", err)
    }
    //return latestBlock.Uint64(), nil
    
    block, err := t.rp.Client.BlockByNumber(context.Background(), latestBlock)
    if err != nil {
        return 0, err
    }
    
    closestMainnetBlock, err := t.findClosestMainnetBlock(block)
    if err != nil {
        return 0, err
    }

    return closestMainnetBlock, nil
}


// Performs a binary search to find the block on mainnet that has the closest timestamp
func (t *submitRplPrice) findClosestMainnetBlock(testnetBlock *types.Block) (uint64, error) {

    // Get the timestamp of the target block on the testnet, and the latest block on mainnet
    testnetTime := float64(testnetBlock.Time())
    latestMainnetBlock, err := t.mnec.BlockByNumber(context.Background(), nil)
    if err != nil {
        return 0, nil
    }

    // Start at the halfway point
    candidateBlockNumber := big.NewInt(0).Div(latestMainnetBlock.Number(), big.NewInt(2))
    candidateBlock, err := t.mnec.BlockByNumber(context.Background(), candidateBlockNumber)
    if err != nil {
        return 0, nil
    }
    bestBlock := candidateBlock
    pivotSize := candidateBlock.NumberU64()
    minimumDistance := +math.Inf(1)

    for {
        // Get the distance from the candidate block to the target time
        candidateTime := float64(candidateBlock.Time())
        delta := testnetTime - candidateTime
        distance := math.Abs(delta)

        // If it's better, replace the best candidate with it
        if distance < minimumDistance {
            minimumDistance = distance
            bestBlock = candidateBlock
        } else if pivotSize == 1 {
            // If the pivot is down to size 1 and we didn't find anything better
            // after another iteration, this is the best block!
            return bestBlock.NumberU64(), nil
        }

        // Iterate over the correct half, setting the pivot to the halfway point of that half (rounded up)
        pivotSize = uint64(math.Ceil(float64(pivotSize) / 2))
        if delta < 0 {
            // Go left
            candidateBlockNumber = big.NewInt(0).Sub(candidateBlockNumber, big.NewInt(int64(pivotSize)))
        } else {
            // Go right
            candidateBlockNumber = big.NewInt(0).Add(candidateBlockNumber, big.NewInt(int64(pivotSize)))
        }

        // Clamp the new candidate to the latest block
        if candidateBlockNumber.Uint64() > (latestMainnetBlock.NumberU64() - 1) {
            candidateBlockNumber.SetUint64(latestMainnetBlock.NumberU64() - 1)
        }
        
        candidateBlock, err = t.mnec.BlockByNumber(context.Background(), candidateBlockNumber)
        if err != nil {
            return 0, nil
        }
    }
}


// Check whether prices for a block can be submitted by the node
func (t *submitRplPrice) canSubmitBlockPrice(nodeAddress common.Address, blockNumber uint64) (bool, error) {

    // Data
    var wg errgroup.Group
    var currentPricesBlock uint64
    var nodeSubmittedBlock bool

    // Get data
    wg.Go(func() error {
        var err error
        currentPricesBlock, err = network.GetPricesBlock(t.rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        blockNumberBuf := make([]byte, 32)
        big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
        nodeSubmittedBlock, err = t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.prices.submitted.node"), nodeAddress.Bytes(), blockNumberBuf))
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (blockNumber > currentPricesBlock && !nodeSubmittedBlock), nil

}


// Get RPL price at block
func (t *submitRplPrice) getRplPrice(blockNumber uint64) (*big.Int, error) {

    // Require 1inch oracle contract
    if err := services.RequireOneInchOracle(t.c); err != nil {
        return nil, err
    }

    // Get RPL token address
    rplAddress := common.HexToAddress(t.cfg.Rocketpool.RplTokenAddress)

    // Initialize call options
    opts := &bind.CallOpts{
        BlockNumber: big.NewInt(int64(blockNumber)),
    }

    // Get RPL price
    rplPrice, err := t.oio.GetRateToEth(opts, rplAddress, true)
    if err != nil {
        return nil, fmt.Errorf("Could not get RPL price at block %d: %w", blockNumber, err)
    }

    // Return
    return rplPrice, nil

}


// Submit RPL price and total effective RPL stake
func (t *submitRplPrice) submitRplPrice(blockNumber uint64, rplPrice, effectiveRplStake *big.Int) error {

    // Log
    t.log.Printlnf("Submitting RPL price for block %d...", blockNumber)

    // Get transactor
    opts, err := t.w.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Get the gas estimates
    gasInfo, err := network.EstimateSubmitPricesGas(t.rp, blockNumber, rplPrice, effectiveRplStake, opts)
    if err != nil {
        return fmt.Errorf("Could not estimate the gas required to dissolve the minipool: %w", err)
    }
    if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log) {
        return nil
    }

    // Submit RPL price
    hash, err := network.SubmitPrices(t.rp, blockNumber, rplPrice, effectiveRplStake, opts)
    if err != nil {
        return err
    }

    // Print TX info and wait for it to be mined
    err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
    if err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Successfully submitted RPL price for block %d.", blockNumber)

    // Return
    return nil

}

