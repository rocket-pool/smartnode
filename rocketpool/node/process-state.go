package node

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	v100_rewards "github.com/rocket-pool/rocketpool-go/legacy/v1.0.0/rewards"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	node_state "github.com/rocket-pool/smartnode/shared/services/node-state"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Process node state task
type processState struct {
	c                *cli.Context
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	state            *node_state.NodeState
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
}

// Create process node state task
func newProcessState(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger) (*processState, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Return task
	lock := &sync.Mutex{}
	return &processState{
		c:                c,
		log:              logger,
		errLog:           errorLogger,
		cfg:              cfg,
		w:                w,
		rp:               rp,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Records Collection]",
	}, nil

}

// Manage fee recipient
func (p *processState) run(state *state.NetworkState) error {

	if p.state == nil {
		p.log.Println("Loading cumulative node records...")

		// Attempt to to load the node state
		state, exists, err := node_state.LoadState(p.cfg)
		if err != nil {
			return fmt.Errorf("error loading node records: %w", err)
		}

		if !exists {
			// Create a new state, with the deploy block as the starter
			p.state = node_state.NewNodeState()
			deployBlockHash := crypto.Keccak256Hash([]byte("deploy.block"))
			latestKnownBlockBig, err := p.rp.RocketStorage.GetUint(nil, deployBlockHash)
			if err != nil {
				return fmt.Errorf("error getting Rocket Pool deployment block: %w", err)
			}
			p.state.NextBlockToCheck = latestKnownBlockBig.Uint64()
		} else {
			p.state = state
		}
	}

	// Check if the check is already running
	p.lock.Lock()
	if p.isRunning {
		p.log.Println("Node records collector is already running in the background.")
		p.lock.Unlock()
		return nil
	}
	p.lock.Unlock()

	// Run the check
	go func() {
		p.lock.Lock()
		p.isRunning = true
		p.lock.Unlock()
		p.printMessage("Starting node records collection in a separate thread.")

		err := p.processLogs()
		if err != nil {
			p.handleError(fmt.Errorf("%s %w", p.generationPrefix, err))
			return
		}

		p.lock.Lock()
		p.isRunning = false
		p.lock.Unlock()
	}()

	// Return
	return nil

}

// Process the logs from the start block to the chain head
func (p *processState) processLogs() error {

	// Get the node address
	account, err := p.w.GetNodeAccount()
	if err != nil {
		return fmt.Errorf("error getting node account: %w", err)
	}
	nodeAddress := account.Address

	// Get the latest block at the time of starting this check (it's fine that it goes stale if the check is long)
	latestBlock, err := p.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting latest block: %w", err)
	}

	// Get the event log limit for the EC
	logWindow, err := p.cfg.GetEventLogInterval()
	if err != nil {
		return fmt.Errorf("error getting the event log window length: %w", err)
	}
	interval := big.NewInt(int64(logWindow))

	// Get the blocks that each network upgrade occurred on
	upgradeBlocks := p.cfg.Smartnode.GetNetworkUpgradeBlockMaps()
	redstoneBlock := int64(-1)
	atlasBlock := int64(-1)
	if len(upgradeBlocks) > 0 {
		redstoneBlock = int64(upgradeBlocks[0])
	}
	if len(upgradeBlocks) > 1 {
		atlasBlock = int64(upgradeBlocks[1])
	}

	// Get the legacy contract addresses
	v100RewardsPoolAddress := p.cfg.Smartnode.GetV100RewardsPoolAddress()

	// Get all the events
	start := big.NewInt(0).SetUint64(p.state.NextBlockToCheck)
	for i := p.state.NextBlockToCheck; i < latestBlock; i += uint64(logWindow) {
		if redstoneBlock == -1 || int64(i) < redstoneBlock {
			// If this is pre-Redstone, get the vanilla RPL rewards events for this node
			claims, err := v100_rewards.GetRPLTokenClaimsForNode(p.rp, nodeAddress, start, interval, &v100RewardsPoolAddress, nil)
			if err != nil {
				return fmt.Errorf("error checking classic RPL rewards claim events for block %s: %w", start.String(), err)
			}
			for _, claim := range claims {
				p.state.RplRewards.AddReward(claim.Amount, claim.Time)
			}
		} else {
			if (atlasBlock != -1) && (int64(i)+int64(logWindow) >= atlasBlock) {
				// If this is post-Atlas, get the withdrawal info for each minipool somehow?
			}

			// Redstone and beyond, query the rewards files for RPL and ETH rewards
			// How about fee recipients?
		}

		start.Add(start, interval)
		p.state.NextBlockToCheck = i + uint64(logWindow)
		err := p.state.Save(p.cfg)
		if err != nil {
			return fmt.Errorf("error saving node records: %w", err)
		}
	}

	return nil

}

func (p *processState) handleError(err error) {
	p.errLog.Println(err)
	p.errLog.Println("*** Error during node records collection. ***")
	p.lock.Lock()
	p.isRunning = false
	p.lock.Unlock()
}

// Print a message
func (p *processState) printMessage(message string) {
	p.log.Printlnf("%s %s", p.generationPrefix, message)
}
