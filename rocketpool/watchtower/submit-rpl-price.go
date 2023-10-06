package watchtower

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	mathutils "github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	OptimismMessengerAbi string = `[
		{
		"inputs": [],
		"name": "rateStale",
		"outputs": [
			{
			"internalType": "bool",
			"name": "",
			"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
		},
		{
		"inputs": [],
		"name": "submitRate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
		}
	]`

	PolygonMessengerAbi string = `[
		{
		"inputs": [],
		"name": "rateStale",
		"outputs": [
			{
			"internalType": "bool",
			"name": "",
			"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
		},
		{
		"inputs": [],
		"name": "submitRate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
		}
	]`

	ArbitrumMessengerAbi string = `[
		{
		"inputs": [],
		"name": "rateStale",
		"outputs": [
			{
			"internalType": "bool",
			"name": "",
			"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
		},
		{
		"inputs": [
			{
			"internalType": "uint256",
			"name": "_maxSubmissionCost",
			"type": "uint256"
			},
			{
			"internalType": "uint256",
			"name": "_gasLimit",
			"type": "uint256"
			},
			{
			"internalType": "uint256",
			"name": "_gasPriceBid",
			"type": "uint256"
			}
		],
		"name": "submitRate",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
		}
	]`

	zkSyncEraMessengerAbi string = `[
		{
			"inputs": [],
			"name": "rateStale",
			"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
			],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [
			{
				"internalType": "uint256",
				"name": "_l2GasLimit",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "_l2GasPerPubdataByteLimit",
				"type": "uint256"
			}
			],
			"name": "submitRate",
			"outputs": [],
			"stateMutability": "payable",
			"type": "function"
		}
	]`

	RplTwapPoolAbi string = `[
		{
		"inputs": [{
			"internalType": "uint32[]",
			"name": "secondsAgos",
			"type": "uint32[]"
		}],
		"name": "observe",
		"outputs": [{
			"internalType": "int56[]",
			"name": "tickCumulatives",
			"type": "int56[]"
		}, {
			"internalType": "uint160[]",
			"name": "secondsPerLiquidityCumulativeX128s",
			"type": "uint160[]"
		}],
		"stateMutability": "view",
		"type": "function"
		}
	]`
)

// Settings
const (
	SubmissionKey string = "network.prices.submitted.node.key"
	BlocksPerTurn uint64 = 75 // Approx. 15 minutes

	twapNumberOfSeconds uint32 = 60 * 60 * 12 // 12 hours
)

type poolObserveResponse struct {
	TickCumulatives                    []*big.Int `abi:"tickCumulatives"`
	SecondsPerLiquidityCumulativeX128s []*big.Int `abi:"secondsPerLiquidityCumulativeX128s"`
}

// Submit RPL price task
type submitRplPrice struct {
	c         *cli.Context
	log       log.ColorLogger
	errLog    log.ColorLogger
	cfg       *config.RocketPoolConfig
	ec        rocketpool.ExecutionClient
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	lock      *sync.Mutex
	isRunning bool
}

// Create submit RPL price task
func newSubmitRplPrice(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger) (*submitRplPrice, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	lock := &sync.Mutex{}
	return &submitRplPrice{
		c:      c,
		log:    logger,
		errLog: errorLogger,
		cfg:    cfg,
		ec:     ec,
		w:      w,
		rp:     rp,
		bc:     bc,
		lock:   lock,
	}, nil

}

// Submit RPL price
func (t *submitRplPrice) run(state *state.NetworkState) error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Check if submission is enabled
	if !state.NetworkDetails.SubmitPricesEnabled {
		return nil
	}

	// Check if Optimism rate is stale and submit
	err = t.submitOptimismPrice()
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printlnf("Error submitting Optimism price: %s", err.Error())
	}

	// Check if Polygon rate is stale and submit
	err = t.submitPolygonPrice()
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printlnf("Error submitting Polygon price: %s", err.Error())
	}

	// Check if Arbitrum rate is stale and submit
	err = t.submitArbitrumPrice()
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printlnf("Error submitting Arbitrum price: %s", err.Error())
	}

	// Check if zkSync rate is stale and submit
	err = t.submitZkSyncEraPrice()
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printlnf("Error submitting zkSync Era price: %s", err.Error())
	}

	// Check if Base rate is stale and submit
	err = t.submitBasePrice()
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printlnf("Error submitting Base price: %s", err.Error())
	}

	// Log
	t.log.Println("Checking for RPL price checkpoint...")

	// Get block to submit price for
	blockNumber := state.NetworkDetails.LatestReportablePricesBlock

	// Check if a submission needs to be made
	pricesBlock := state.NetworkDetails.PricesBlock
	if blockNumber <= pricesBlock {
		return nil
	}

	// Get the time of the block
	header, err := t.ec.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return err
	}
	blockTime := time.Unix(int64(header.Time), 0)

	// Get the Beacon block corresponding to this time
	eth2Config := state.BeaconConfig
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := blockTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot

	// Check if the targetEpoch is finalized yet
	targetEpoch := slotNumber / eth2Config.SlotsPerEpoch
	beaconHead, err := t.bc.GetBeaconHead()
	if err != nil {
		return err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch
	if targetEpoch > finalizedEpoch {
		t.log.Printlnf("Prices must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", blockNumber, targetEpoch, finalizedEpoch)
		return nil
	}

	// Check if the process is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Prices report is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		logPrefix := "[Price Report]"
		t.log.Printlnf("%s Starting price report in a separate thread.", logPrefix)

		// Log
		t.log.Printlnf("Getting RPL price for block %d...", blockNumber)

		// Get RPL price at block
		rplPrice, err := t.getRplTwap(blockNumber)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}

		// Log
		t.log.Printlnf("RPL price: %.6f ETH", mathutils.RoundDown(eth.WeiToEth(rplPrice), 6))

		// Check if we have reported these specific values before
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockPrices(nodeAccount.Address, blockNumber, rplPrice)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmittedSpecific {
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		// We haven't submitted these values, check if we've submitted any for this block so we can log it
		hasSubmitted, err := t.hasSubmittedBlockPrices(nodeAccount.Address, blockNumber)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmitted {
			t.log.Printlnf("Have previously submitted out-of-date prices for block %d, trying again...", blockNumber)
		}

		// Log
		t.log.Println("Submitting RPL price...")

		// Submit RPL price
		if err := t.submitRplPrice(blockNumber, rplPrice); err != nil {
			t.handleError(fmt.Errorf("%s could not submit RPL price: %w", logPrefix, err))
			return
		}

		// Log and return
		t.log.Printlnf("%s Price report complete.", logPrefix)
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

}

func (t *submitRplPrice) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Price report failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Check whether prices for a block has already been submitted by the node
func (t *submitRplPrice) hasSubmittedBlockPrices(nodeAddress common.Address, blockNumber uint64) (bool, error) {

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte(SubmissionKey), nodeAddress.Bytes(), blockNumberBuf))

}

// Check whether specific prices for a block has already been submitted by the node
func (t *submitRplPrice) hasSubmittedSpecificBlockPrices(nodeAddress common.Address, blockNumber uint64, rplPrice *big.Int) (bool, error) {
	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)

	rplPriceBuf := make([]byte, 32)
	rplPrice.FillBytes(rplPriceBuf)

	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte(SubmissionKey), nodeAddress.Bytes(), blockNumberBuf, rplPriceBuf))
}

// Get RPL price via TWAP at block
func (t *submitRplPrice) getRplTwap(blockNumber uint64) (*big.Int, error) {

	// Initialize call options
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(blockNumber)),
	}

	poolAddress := t.cfg.Smartnode.GetRplTwapPoolAddress()
	if poolAddress == "" {
		return nil, fmt.Errorf("RPL TWAP pool contract not deployed on this network")
	}

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, opts.BlockNumber)
	if err != nil {
		return nil, err
	}

	// Construct the pool contract instance
	parsed, err := abi.JSON(strings.NewReader(RplTwapPoolAbi))
	if err != nil {
		return nil, fmt.Errorf("error decoding RPL TWAP pool ABI: %w", err)
	}
	addr := common.HexToAddress(poolAddress)
	poolContract := bind.NewBoundContract(addr, parsed, client.Client, client.Client, client.Client)
	pool := rocketpool.Contract{
		Contract: poolContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   client.Client,
	}

	// Get RPL price
	response := poolObserveResponse{}
	interval := twapNumberOfSeconds
	args := []uint32{interval, 0}

	err = pool.Call(opts, &response, "observe", args)
	if err != nil {
		return nil, fmt.Errorf("could not get RPL price at block %d: %w", blockNumber, err)
	}
	if len(response.TickCumulatives) < 2 {
		return nil, fmt.Errorf("TWAP contract didn't have enough tick cumulatives for block %d (raw: %v)", blockNumber, response.TickCumulatives)
	}

	tick := big.NewInt(0).Sub(response.TickCumulatives[1], response.TickCumulatives[0])
	tick.Div(tick, big.NewInt(int64(interval))) // tick = (cumulative[1] - cumulative[0]) / interval

	base := eth.EthToWei(1.0001) // 1.0001e18
	one := eth.EthToWei(1)       // 1e18

	numerator := big.NewInt(0).Exp(base, tick, nil) // 1.0001e18 ^ tick
	numerator.Mul(numerator, one)

	denominator := big.NewInt(0).Exp(one, tick, nil) // 1e18 ^ tick
	denominator.Div(numerator, denominator)          // denominator = (1.0001e18^tick / 1e18^tick)

	numerator.Mul(one, one)                               // 1e18 ^ 2
	rplPrice := big.NewInt(0).Div(numerator, denominator) // 1e18 ^ 2 / (1.0001e18^tick * 1e18 / 1e18^tick)

	// Return
	return rplPrice, nil

}

func (t *submitRplPrice) printMessage(message string) {
	t.log.Println(message)
}

// Submit RPL price and total effective RPL stake
func (t *submitRplPrice) submitRplPrice(blockNumber uint64, rplPrice *big.Int) error {

	// Log
	t.log.Printlnf("Submitting RPL price for block %d...", blockNumber)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimateSubmitPricesGas(t.rp, blockNumber, rplPrice, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit RPL price: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = gasInfo.SafeGasLimit

	// Submit RPL price
	hash, err := network.SubmitPrices(t.rp, blockNumber, rplPrice, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted RPL price for block %d.", blockNumber)

	// Return
	return nil

}

// Checks if Optimism rate is stale and if it's our turn to submit, calls submitRate on the messenger
func (t *submitRplPrice) submitOptimismPrice() error {
	priceMessengerAddress := t.cfg.Smartnode.GetOptimismMessengerAddress()

	if priceMessengerAddress == "" {
		// No price messenger deployed on the current network
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Construct the price messenger contract instance
	parsed, err := abi.JSON(strings.NewReader(OptimismMessengerAbi))
	if err != nil {
		return fmt.Errorf("Failed decoding ABI: %q", err)
	}

	addr := common.HexToAddress(priceMessengerAddress)
	priceMessengerContract := bind.NewBoundContract(addr, parsed, t.ec, t.ec, t.ec)
	priceMessenger := rocketpool.Contract{
		Contract: priceMessengerContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   t.ec,
	}

	// Check if the rate is stale
	var out []interface{}
	err = priceMessengerContract.Call(nil, &out, "rateStale")

	if err != nil {
		return fmt.Errorf("Failed to query rate staleness for Optimism: %q", err)
	}

	rateStale := *abi.ConvertType(out[0], new(bool)).(*bool)

	if !rateStale {
		// Nothing to do
		return nil
	}

	// Get total number of ODAO members
	count, err := trustednode.GetMemberCount(t.rp, nil)
	if err != nil {
		return fmt.Errorf("Failed to get member count: %q", err)
	}

	// Find out which index we are
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		addr, err := trustednode.GetMemberAt(t.rp, i, nil)
		if err != nil {
			return fmt.Errorf("Failed to get member at %d: %q", i, err)
		}

		if bytes.Compare(addr.Bytes(), opts.From.Bytes()) == 0 {
			index = i
			break
		}
	}

	// Get current block number
	blockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get block number: %q", err)
	}

	// Calculate whose turn it is to submit
	indexToSubmit := (blockNumber / BlocksPerTurn) % count

	if index == indexToSubmit {

		// Temporary gas calculations until this gets put into a binding
		input, err := priceMessenger.ABI.Pack("submitRate")
		if err != nil {
			return fmt.Errorf("Could not encode input data: %w", err)
		}

		// Estimate gas limit
		gasLimit, err := t.rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       priceMessenger.Address,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Value:    opts.Value,
			Data:     input,
		})
		if err != nil {
			return fmt.Errorf("Error estimating gas limit of submitOptimismPrice: %w", err)
		}

		// Get the safe gas limit
		safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		if gasLimit > rocketpool.MaxGasLimit {
			gasLimit = rocketpool.MaxGasLimit
		}
		if safeGasLimit > rocketpool.MaxGasLimit {
			safeGasLimit = rocketpool.MaxGasLimit
		}
		gasInfo := rocketpool.GasInfo{
			EstGasLimit:  gasLimit,
			SafeGasLimit: safeGasLimit,
		}

		// Print the gas info
		maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasInfo.SafeGasLimit

		t.log.Println("Submitting rate to Optimism...")

		// Submit rates
		tx, err := priceMessenger.Transact(opts, "submitRate")
		if err != nil {
			return fmt.Errorf("Failed to submit rate: %q", err)
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
		if err != nil {
			return err
		}

		// Log
		t.log.Printlnf("Successfully submitted Optimism price for block %d.", blockNumber)

	}

	return nil
}

// Checks if Polygon rate is stale and if it's our turn to submit, calls submitRate on the messenger
func (t *submitRplPrice) submitPolygonPrice() error {
	priceMessengerAddress := t.cfg.Smartnode.GetPolygonMessengerAddress()

	if priceMessengerAddress == "" {
		// No price messenger deployed on the current network
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Construct the price messenger contract instance
	parsed, err := abi.JSON(strings.NewReader(PolygonMessengerAbi))
	if err != nil {
		return fmt.Errorf("Failed decoding ABI: %q", err)
	}

	addr := common.HexToAddress(priceMessengerAddress)
	priceMessengerContract := bind.NewBoundContract(addr, parsed, t.ec, t.ec, t.ec)
	priceMessenger := rocketpool.Contract{
		Contract: priceMessengerContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   t.ec,
	}

	// Check if the rate is stale
	var out []interface{}
	err = priceMessengerContract.Call(nil, &out, "rateStale")

	if err != nil {
		return fmt.Errorf("Failed to query rate staleness for Polygon: %q", err)
	}

	rateStale := *abi.ConvertType(out[0], new(bool)).(*bool)

	if !rateStale {
		// Nothing to do
		return nil
	}

	// Get total number of ODAO members
	count, err := trustednode.GetMemberCount(t.rp, nil)
	if err != nil {
		return fmt.Errorf("Failed to get member count: %q", err)
	}

	// Find out which index we are
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		addr, err := trustednode.GetMemberAt(t.rp, i, nil)
		if err != nil {
			return fmt.Errorf("Failed to get member at %d: %q", i, err)
		}

		if bytes.Compare(addr.Bytes(), opts.From.Bytes()) == 0 {
			index = i
			break
		}
	}

	// Get current block number
	blockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get block number: %q", err)
	}

	// Calculate whose turn it is to submit
	indexToSubmit := (blockNumber / BlocksPerTurn) % count

	if index == indexToSubmit {

		// Temporary gas calculations until this gets put into a binding
		input, err := priceMessenger.ABI.Pack("submitRate")
		if err != nil {
			return fmt.Errorf("Could not encode input data for Polygon price submission: %w", err)
		}

		// Estimate gas limit
		gasLimit, err := t.rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       priceMessenger.Address,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Value:    opts.Value,
			Data:     input,
		})
		if err != nil {
			return fmt.Errorf("Error estimating gas limit of submitPolygonPrice: %w", err)
		}

		// Get the safe gas limit
		safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		if gasLimit > rocketpool.MaxGasLimit {
			gasLimit = rocketpool.MaxGasLimit
		}
		if safeGasLimit > rocketpool.MaxGasLimit {
			safeGasLimit = rocketpool.MaxGasLimit
		}
		gasInfo := rocketpool.GasInfo{
			EstGasLimit:  gasLimit,
			SafeGasLimit: safeGasLimit,
		}

		// Print the gas info
		maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasInfo.SafeGasLimit

		t.log.Println("Submitting rate to Polygon...")

		// Submit rates
		tx, err := priceMessenger.Transact(opts, "submitRate")
		if err != nil {
			return fmt.Errorf("Failed to submit rate to Polygon: %q", err)
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
		if err != nil {
			return err
		}

		// Log
		t.log.Printlnf("Successfully submitted Polygon price for block %d.", blockNumber)

	}

	return nil
}

// Checks if Arbitrum rate is stale and if it's our turn to submit, calls submitRate on the messenger
func (t *submitRplPrice) submitArbitrumPrice() error {
	priceMessengerAddress := t.cfg.Smartnode.GetArbitrumMessengerAddress()

	if priceMessengerAddress == "" {
		// No price messenger deployed on the current network
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Construct the price messenger contract instance
	parsed, err := abi.JSON(strings.NewReader(ArbitrumMessengerAbi))
	if err != nil {
		return fmt.Errorf("Failed decoding ABI: %q", err)
	}

	addr := common.HexToAddress(priceMessengerAddress)
	priceMessengerContract := bind.NewBoundContract(addr, parsed, t.ec, t.ec, t.ec)
	priceMessenger := rocketpool.Contract{
		Contract: priceMessengerContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   t.ec,
	}

	// Check if the rate is stale
	var out []interface{}
	err = priceMessengerContract.Call(nil, &out, "rateStale")

	if err != nil {
		return fmt.Errorf("Failed to query rate staleness for Arbitrum: %q", err)
	}

	rateStale := *abi.ConvertType(out[0], new(bool)).(*bool)

	if !rateStale {
		// Nothing to do
		return nil
	}

	// Get total number of ODAO members
	count, err := trustednode.GetMemberCount(t.rp, nil)
	if err != nil {
		return fmt.Errorf("Failed to get member count: %q", err)
	}

	// Find out which index we are
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		addr, err := trustednode.GetMemberAt(t.rp, i, nil)
		if err != nil {
			return fmt.Errorf("Failed to get member at %d: %q", i, err)
		}

		if bytes.Compare(addr.Bytes(), opts.From.Bytes()) == 0 {
			index = i
			break
		}
	}

	// Get current block number
	blockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get block number: %q", err)
	}

	// Calculate whose turn it is to submit
	indexToSubmit := (blockNumber / BlocksPerTurn) % count

	if index == indexToSubmit {

		// Get the current network recommended max fee
		suggestedMaxFee, err := rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return fmt.Errorf("error getting recommended base fee from the network for Arbitrum price submission: %w", err)
		}

		// Constants for Arbitrum
		bufferMultiplier := big.NewInt(4)
		dataLength := big.NewInt(36)
		arbitrumGasLimit := big.NewInt(40000)
		arbitrumMaxFeePerGas := eth.GweiToWei(0.1)

		// Gas limit calculation on Arbitrum
		maxSubmissionCost := big.NewInt(6)
		maxSubmissionCost.Mul(maxSubmissionCost, dataLength)
		maxSubmissionCost.Add(maxSubmissionCost, big.NewInt(1400))
		maxSubmissionCost.Mul(maxSubmissionCost, suggestedMaxFee)  // (1400 + 6 * dataLength) * baseFee
		maxSubmissionCost.Mul(maxSubmissionCost, bufferMultiplier) // Multiply by the buffer constant for safety

		// Provide enough ETH for the L2 and roundtrip TX's
		value := big.NewInt(0)
		value.Mul(arbitrumGasLimit, arbitrumMaxFeePerGas)
		value.Add(value, maxSubmissionCost)
		opts.Value = value

		// Temporary gas calculations until this gets put into a binding
		input, err := priceMessenger.ABI.Pack("submitRate", maxSubmissionCost, arbitrumGasLimit, arbitrumMaxFeePerGas)
		if err != nil {
			return fmt.Errorf("Could not encode input data for Arbitrum price submission: %w", err)
		}

		// Estimate gas limit
		gasLimit, err := t.rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       priceMessenger.Address,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Value:    opts.Value,
			Data:     input,
		})
		if err != nil {
			return fmt.Errorf("Error estimating gas limit of submitArbitrumPrice: %w", err)
		}

		// Get the safe gas limit
		safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		if gasLimit > rocketpool.MaxGasLimit {
			gasLimit = rocketpool.MaxGasLimit
		}
		if safeGasLimit > rocketpool.MaxGasLimit {
			safeGasLimit = rocketpool.MaxGasLimit
		}
		gasInfo := rocketpool.GasInfo{
			EstGasLimit:  gasLimit,
			SafeGasLimit: safeGasLimit,
		}

		// Print the gas info
		maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasInfo.SafeGasLimit

		t.log.Println("Submitting rate to Arbitrum...")

		// Submit rates
		tx, err := priceMessenger.Transact(opts, "submitRate", maxSubmissionCost, arbitrumGasLimit, arbitrumMaxFeePerGas)
		if err != nil {
			return fmt.Errorf("Failed to submit Arbitrum rate: %q", err)
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
		if err != nil {
			return err
		}

		// Log
		t.log.Printlnf("Successfully submitted Arbitrum price for block %d.", blockNumber)

	}

	return nil
}

// Checks if zkSync Era rate is stale and if it's our turn to submit, calls submitRate on the messenger
func (t *submitRplPrice) submitZkSyncEraPrice() error {
	priceMessengerAddress := t.cfg.Smartnode.GetZkSyncEraMessengerAddress()

	if priceMessengerAddress == "" {
		// No price messenger deployed on the current network
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Construct the price messenger contract instance
	parsed, err := abi.JSON(strings.NewReader(zkSyncEraMessengerAbi))
	if err != nil {
		return fmt.Errorf("Failed decoding ABI: %q", err)
	}

	addr := common.HexToAddress(priceMessengerAddress)
	priceMessengerContract := bind.NewBoundContract(addr, parsed, t.ec, t.ec, t.ec)
	priceMessenger := rocketpool.Contract{
		Contract: priceMessengerContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   t.ec,
	}

	// Check if the rate is stale
	var out []interface{}
	err = priceMessengerContract.Call(nil, &out, "rateStale")

	if err != nil {
		return fmt.Errorf("Failed to query rate staleness for zkSync Era: %q", err)
	}

	rateStale := *abi.ConvertType(out[0], new(bool)).(*bool)

	if !rateStale {
		// Nothing to do
		return nil
	}

	// Get total number of ODAO members
	count, err := trustednode.GetMemberCount(t.rp, nil)
	if err != nil {
		return fmt.Errorf("Failed to get member count: %q", err)
	}

	// Find out which index we are
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		addr, err := trustednode.GetMemberAt(t.rp, i, nil)
		if err != nil {
			return fmt.Errorf("Failed to get member at %d: %q", i, err)
		}

		if bytes.Compare(addr.Bytes(), opts.From.Bytes()) == 0 {
			index = i
			break
		}
	}

	// Get current block number
	blockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get block number: %q", err)
	}

	// Calculate whose turn it is to submit
	indexToSubmit := (blockNumber / BlocksPerTurn) % count

	if index == indexToSubmit {

		// Constants for zkSync Era
		l1GasPerPubdataByte := big.NewInt(17)
		fairL2GasPrice := eth.GweiToWei(0.5)
		l2GasLimit := big.NewInt(750000)
		gasPerPubdataByte := big.NewInt(800)
		maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))

		// Value calculation on zkSync Era
		pubdataPrice := big.NewInt(0).Mul(l1GasPerPubdataByte, maxFee)
		minL2GasPrice := big.NewInt(0).Add(pubdataPrice, gasPerPubdataByte)
		minL2GasPrice.Sub(minL2GasPrice, big.NewInt(1))
		minL2GasPrice.Div(minL2GasPrice, gasPerPubdataByte)
		gasPrice := big.NewInt(0).Set(fairL2GasPrice)
		if minL2GasPrice.Cmp(gasPrice) > 0 {
			gasPrice.Set(minL2GasPrice)
		}
		txValue := big.NewInt(0).Mul(l2GasLimit, gasPrice)
		opts.Value = txValue

		// Temporary gas calculations until this gets put into a binding
		input, err := priceMessenger.ABI.Pack("submitRate", l2GasLimit, gasPerPubdataByte)
		if err != nil {
			return fmt.Errorf("Could not encode input data for zkSync Era price submission: %w", err)
		}

		// Estimate gas limit
		gasLimit, err := t.rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       priceMessenger.Address,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Value:    opts.Value,
			Data:     input,
		})
		if err != nil {
			return fmt.Errorf("Error estimating gas limit of submitZkSyncEraPrice: %w", err)
		}

		// Get the safe gas limit
		safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		if gasLimit > rocketpool.MaxGasLimit {
			gasLimit = rocketpool.MaxGasLimit
		}
		if safeGasLimit > rocketpool.MaxGasLimit {
			safeGasLimit = rocketpool.MaxGasLimit
		}
		gasInfo := rocketpool.GasInfo{
			EstGasLimit:  gasLimit,
			SafeGasLimit: safeGasLimit,
		}

		// Print the gas info
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasInfo.SafeGasLimit

		t.log.Println("Submitting rate to zkSync Era...")

		// Submit rates
		tx, err := priceMessenger.Transact(opts, "submitRate", l2GasLimit, gasPerPubdataByte)
		if err != nil {
			return fmt.Errorf("Failed to submit zkSync Era rate: %q", err)
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
		if err != nil {
			return err
		}

		// Log
		t.log.Printlnf("Successfully submitted zkSync Era price for block %d.", blockNumber)

	}

	return nil
}

// Checks if Base rate is stale and if it's our turn to submit, calls submitRate on the messenger
func (t *submitRplPrice) submitBasePrice() error {
	priceMessengerAddress := t.cfg.Smartnode.GetBaseMessengerAddress()

	if priceMessengerAddress == "" {
		// No price messenger deployed on the current network
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Construct the price messenger contract instance - this is the same as the Optimism messenger
	parsed, err := abi.JSON(strings.NewReader(OptimismMessengerAbi))
	if err != nil {
		return fmt.Errorf("Failed decoding ABI: %q", err)
	}

	addr := common.HexToAddress(priceMessengerAddress)
	priceMessengerContract := bind.NewBoundContract(addr, parsed, t.ec, t.ec, t.ec)
	priceMessenger := rocketpool.Contract{
		Contract: priceMessengerContract,
		Address:  &addr,
		ABI:      &parsed,
		Client:   t.ec,
	}

	// Check if the rate is stale
	var out []interface{}
	err = priceMessengerContract.Call(nil, &out, "rateStale")

	if err != nil {
		return fmt.Errorf("Failed to query rate staleness for Base: %q", err)
	}

	rateStale := *abi.ConvertType(out[0], new(bool)).(*bool)

	if !rateStale {
		// Nothing to do
		return nil
	}

	// Get total number of ODAO members
	count, err := trustednode.GetMemberCount(t.rp, nil)
	if err != nil {
		return fmt.Errorf("Failed to get member count: %q", err)
	}

	// Find out which index we are
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		addr, err := trustednode.GetMemberAt(t.rp, i, nil)
		if err != nil {
			return fmt.Errorf("Failed to get member at %d: %q", i, err)
		}

		if bytes.Compare(addr.Bytes(), opts.From.Bytes()) == 0 {
			index = i
			break
		}
	}

	// Get current block number
	blockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get block number: %q", err)
	}

	// Calculate whose turn it is to submit
	indexToSubmit := (blockNumber / BlocksPerTurn) % count

	if index == indexToSubmit {

		// Temporary gas calculations until this gets put into a binding
		input, err := priceMessenger.ABI.Pack("submitRate")
		if err != nil {
			return fmt.Errorf("Could not encode input data: %w", err)
		}

		// Estimate gas limit
		gasLimit, err := t.rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       priceMessenger.Address,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Value:    opts.Value,
			Data:     input,
		})
		if err != nil {
			return fmt.Errorf("Error estimating gas limit of submitBasePrice: %w", err)
		}

		// Get the safe gas limit
		safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		if gasLimit > rocketpool.MaxGasLimit {
			gasLimit = rocketpool.MaxGasLimit
		}
		if safeGasLimit > rocketpool.MaxGasLimit {
			safeGasLimit = rocketpool.MaxGasLimit
		}
		gasInfo := rocketpool.GasInfo{
			EstGasLimit:  gasLimit,
			SafeGasLimit: safeGasLimit,
		}

		// Print the gas info
		maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasInfo.SafeGasLimit

		t.log.Println("Submitting rate to Base...")

		// Submit rates
		tx, err := priceMessenger.Transact(opts, "submitRate")
		if err != nil {
			return fmt.Errorf("Failed to submit rate: %q", err)
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
		if err != nil {
			return err
		}

		// Log
		t.log.Printlnf("Successfully submitted Base price for block %d.", blockNumber)

	}

	return nil
}
