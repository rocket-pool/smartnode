package node

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Manage download rewards trees task
type downloadRewardsTrees struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
	d   *client.Client
	bc  beacon.Client
}

// Create manage fee recipient task
func newDownloadRewardsTrees(c *cli.Context, logger log.ColorLogger) (*downloadRewardsTrees, error) {

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
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &downloadRewardsTrees{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		rp:  rp,
		d:   d,
		bc:  bc,
	}, nil

}

// Manage fee recipient
func (d *downloadRewardsTrees) run(state *state.NetworkState) error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(d.c, true); err != nil {
		return err
	}

	// Check if the user opted into downloading rewards files
	if d.cfg.Smartnode.RewardsTreeMode.Value.(cfgtypes.RewardsMode) != cfgtypes.RewardsMode_Download {
		return nil
	}

	// Log
	d.log.Println("Checking for new rewards tree files to download...")

	// Get node account
	nodeAccount, err := d.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get the current interval
	currentIndex := state.NetworkDetails.RewardIndex

	// Check for missing intervals
	missingIntervals := []uint64{}
	for i := uint64(0); i < currentIndex; i++ {
		// Check if the tree file exists
		treeFilePath := d.cfg.Smartnode.GetRewardsTreePath(i, true, config.RewardsExtensionJSON)
		_, err = os.Stat(treeFilePath)
		if os.IsNotExist(err) {
			d.log.Printlnf("You are missing the rewards tree file for interval %d.", i)
			missingIntervals = append(missingIntervals, i)
		} else if err != nil {
			return fmt.Errorf("error checking if rewards interval %d file exists: %w", i, err)
		}
	}

	if len(missingIntervals) == 0 {
		return nil
	}

	// Download missing intervals
	for _, missingInterval := range missingIntervals {
		fmt.Printf("Downloading interval %d file... ", missingInterval)
		intervalInfo, err := rprewards.GetIntervalInfo(d.rp, d.cfg, nodeAccount.Address, missingInterval, nil)
		if err != nil {
			return fmt.Errorf("error getting interval %d info: %w", missingInterval, err)
		}
		err = intervalInfo.DownloadRewardsFile(d.cfg, true)
		if err != nil {
			fmt.Println()
			return err
		}
		fmt.Println("done!")
	}

	return nil

}
