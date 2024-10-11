package node

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

type AutoInitVotingPower struct {
	ctx            context.Context
	sp             *services.ServiceProvider
	logger         *slog.Logger
	cfg            *config.SmartNodeConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.IBeaconClient
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	nodeAddress    common.Address
}

// Auto Initialize Vote Power
func NewAutoInitVotingPower(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger) *AutoInitVotingPower {
	cfg := sp.GetConfig()
	log := logger.With(slog.String(keys.TaskKey, "Auto Initialize Vote Power"))
	maxFee, maxPriorityFee := getAutoTxInfo(cfg, log)
	return &AutoInitVotingPower{
		ctx:            ctx,
		sp:             sp,
		logger:         log,
		cfg:            cfg,
		w:              sp.GetWallet(),
		rp:             sp.GetRocketPool(),
		bc:             sp.GetBeaconClient(),
		gasThreshold:   cfg.AutoInitVPThreshold.Value,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
	}
}

func (t *AutoInitVotingPower) Run(state *state.NetworkState) error {
	// Log
	t.logger.Info("Checking for node initialized voting power...")

	// Create Node Binding
	t.nodeAddress, _ = t.w.GetAddress()
	node, err := node.NewNode(t.rp, t.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node binding for node %s: %w", t.nodeAddress.Hex(), err)
	}

	// Check if voting is initialized
	err = t.rp.Query(nil, nil, node.IsVotingInitialized)
	if err != nil {
		return fmt.Errorf("error checking if voting is initialized for node %s: %w", t.nodeAddress.Hex(), err)
	}
	votingInitialized := node.IsVotingInitialized.Get()

	// Create the tx and submit if voting isn't initialized
	if !votingInitialized {
		txSubmission, err := t.createInitializeVotingTx()
		if err != nil {
			return fmt.Errorf("error preparing submission to initialize voting for node %s: %w", t.nodeAddress.Hex(), err)
		}
		err = t.initializeVotingPower(txSubmission)
		if err != nil {
			return fmt.Errorf("error initializing voting power for node %s: %w", t.nodeAddress.Hex(), err)
		}
		return nil
	}

	return nil
}

func (t *AutoInitVotingPower) createInitializeVotingTx() (*eth.TransactionSubmission, error) {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Bindings
	node, err := node.NewNode(t.rp, t.nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", t.nodeAddress.Hex(), err)
	}
	// Get the tx info
	txInfo, err := node.InitializeVoting(opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating the gas required to initialize voting for node %s: %w", t.nodeAddress.Hex(), err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return nil, fmt.Errorf("simulating initialize voting tx for node %s failed: %s", t.nodeAddress.Hex(), txInfo.SimulationResult.SimulationError)
	}

	submission, err := eth.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating submission to initialize voting for node %s: %w", t.nodeAddress.Hex(), err)
	}
	return submission, nil
}

func (t *AutoInitVotingPower) initializeVotingPower(submission *eth.TransactionSubmission) error {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(t.logger)
		if err != nil {
			return err
		}
	}

	// Lower the priority fee when the suggested maxfee is lower than the user requested priority fee
	if maxFee.Cmp(t.maxPriorityFee) < 0 {
		t.maxPriorityFee = new(big.Int).Div(maxFee, big.NewInt(2))
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(submission.TxInfo.SimulationResult, true, t.gasThreshold, t.logger, opts.GasFeeCap, 0) {
		return nil
	}

	// Print TX info and wait for them to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.logger, submission.TxInfo, opts)
	if err != nil {
		return err
	}

	return nil
}
