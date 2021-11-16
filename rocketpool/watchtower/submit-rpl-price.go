package watchtower

import (
    "context"
    "fmt"
    "github.com/ethereum/go-ethereum/ethclient"
    "math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

// Settings
const SubmitFollowDistancePrices = 2
const ConfirmDistancePrices = 30

// Submit RPL price task
type submitRplPrice struct {
    c *cli.Context
    log log.ColorLogger
    cfg config.RocketPoolConfig
    ec *ethclient.Client
    w *wallet.Wallet
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
    ec, err := services.GetEthClient(c)
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
        ec: ec,
        w: w,
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

    // Allow some blocks to pass in case of a short reorg
    currentBlockNumber, err := t.ec.BlockNumber(context.Background())
    if err != nil {
        return err
    }
    if blockNumber + SubmitFollowDistancePrices > currentBlockNumber {
        return nil
    }

    // Check if a submission needs to be made
    pricesBlock, err := network.GetPricesBlock(t.rp, nil)
    if err != nil {
        return err
    }
    if blockNumber > pricesBlock {
        return nil
    }

    // If confirm distance has passed, we just want to ensure we have submitted and then early exit
    if blockNumber + ConfirmDistancePrices <= currentBlockNumber {
        hasSubmitted, err := t.hasSubmittedBlockPrices(nodeAccount.Address, blockNumber)
        if err != nil {
            return err
        }
        if hasSubmitted {
            return nil
        }
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

    // Check if we have reported these specific values before
    hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockPrices(nodeAccount.Address, blockNumber, rplPrice, effectiveRplStake)
    if err != nil {
        return err
    }
    if hasSubmittedSpecific {
        return nil
    }

    // We haven't submitted these values, check if we've submitted any for this block so we can log it
    hasSubmitted, err := t.hasSubmittedBlockPrices(nodeAccount.Address, blockNumber)
    if err != nil {
        return err
    }
    if hasSubmitted {
        t.log.Printlnf("Have previously submitted out-of-date prices for block $d, trying again...", blockNumber)
    }

    // Log
    t.log.Println("Submitting RPL price...")

    // Submit RPL price
    if err := t.submitRplPrice(blockNumber, rplPrice, effectiveRplStake); err != nil {
        return fmt.Errorf("Could not submit RPL price: %w", err)
    }

    // Return
    return nil

}


// Get the latest block number to report RPL price for
func (t *submitRplPrice) getLatestReportableBlock() (uint64, error) {

    // Require eth client synced
    if err := services.RequireEthClientSynced(t.c); err != nil {
        return 0, err
    }

    latestBlock, err := network.GetLatestReportablePricesBlock(t.rp, nil)
    if err != nil {
        return 0, fmt.Errorf("Error getting latest reportable block: %w", err)
    }
    return latestBlock.Uint64(), nil

}


// Check whether prices for a block has already been submitted by the node
func (t *submitRplPrice) hasSubmittedBlockPrices(nodeAddress common.Address, blockNumber uint64) (bool, error) {

    blockNumberBuf := make([]byte, 32)
    big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
    return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.prices.submitted.node"), nodeAddress.Bytes(), blockNumberBuf))

}


// Check whether specific prices for a block has already been submitted by the node
func (t *submitRplPrice) hasSubmittedSpecificBlockPrices(nodeAddress common.Address, blockNumber uint64, rplPrice, effectiveRplStake *big.Int) (bool, error) {

    blockNumberBuf := make([]byte, 32)
    big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)

    rplPriceBuf := make([]byte, 32)
    rplPrice.FillBytes(rplPriceBuf)

    effectiveRplStakeBuf := make([]byte, 32)
    effectiveRplStake.FillBytes(effectiveRplStakeBuf)

    return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.prices.submitted.node"), nodeAddress.Bytes(), blockNumberBuf, rplPriceBuf, effectiveRplStakeBuf))

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
        return fmt.Errorf("Could not estimate the gas required to submit RPL price: %w", err)
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

