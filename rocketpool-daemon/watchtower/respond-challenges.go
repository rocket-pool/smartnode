package watchtower

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/utils"
)

// Respond to challenges task
type RespondChallenges struct {
	sp  *services.ServiceProvider
	log log.ColorLogger
}

// Create respond to challenges task
func NewRespondChallenges(sp *services.ServiceProvider, logger log.ColorLogger, m *state.NetworkStateManager) *RespondChallenges {
	return &RespondChallenges{
		sp:  sp,
		log: logger,
	}
}

// Respond to challenges
func (t *RespondChallenges) Run() error {
	// Get services
	cfg := t.sp.GetConfig()
	w := t.sp.GetWallet()
	rp := t.sp.GetRocketPool()
	nodeAddress, _ := w.GetAddress()

	// Log
	t.log.Println("Checking for challenges to respond to...")
	member, err := oracle.NewOracleDaoMember(rp, nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO member binding: %w", err)
	}

	// Check for active challenges
	err = rp.Query(nil, nil, member.IsChallenged)
	if err != nil {
		return fmt.Errorf("error checking if member is challenged: %w", err)
	}
	if !member.IsChallenged.Get() {
		return nil
	}

	// Log
	t.log.Printlnf("Node %s has an active challenge against it, responding...", nodeAddress.Hex())

	// Get transactor
	opts, err := w.GetTransactor()
	if err != nil {
		return err
	}

	// Create an oDAO manager
	odaoMgr, err := oracle.NewOracleDaoManager(rp)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}

	// Get the tx info
	txInfo, err := odaoMgr.DecideChallenge(nodeAddress, opts)
	if err != nil {
		return fmt.Errorf("error getting DecideChallenge TX info: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating DecideChallenge TX failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log & return
	t.log.Printlnf("Successfully responded to challenge against node %s.", nodeAddress.Hex())
	return nil
}
