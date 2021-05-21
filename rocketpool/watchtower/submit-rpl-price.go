package watchtower

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
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
    effectiveRplState, err := node.CalculateTotalEffectiveRPLStake(t.rp, zero, zero, rplPrice, nil)
    if err != nil {
        return fmt.Errorf("Error getting total effective RPL stake: %w", err)
    }

    // Log
    t.log.Printlnf("RPL price: %.6f ETH", math.RoundDown(eth.WeiToEth(rplPrice), 6))

    // Submit RPL price
    if err := t.submitRplPrice(blockNumber, rplPrice, effectiveRplState); err != nil {
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
    rplPrice, err := t.oio.GetRate(opts, rplAddress, common.Address{})
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

    // Submit RPL price
    if _, err := network.SubmitPrices(t.rp, blockNumber, rplPrice, effectiveRplStake, opts); err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Successfully submitted RPL price for block %d.", blockNumber)

    // Return
    return nil

}

