package node

import (
	"fmt"
	"os"

	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rprewards "github.com/rocket-pool/smartnode/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Manage download rewards trees task
type DownloadRewardsTrees struct {
	sp  *services.ServiceProvider
	log **log.Logger
	cfg *config.SmartNodeConfig
	rp  *rocketpool.RocketPool
}

// Create manage fee recipient task
func NewDownloadRewardsTrees(sp *services.ServiceProvider, logger *log.Logger) *DownloadRewardsTrees {
	return &DownloadRewardsTrees{
		sp:  sp,
		log: &logger,
		cfg: sp.GetConfig(),
		rp:  sp.GetRocketPool(),
	}
}

// Manage fee recipient
func (t *DownloadRewardsTrees) Run(state *state.NetworkState) error {
	// Check if the user opted into downloading rewards files
	if t.cfg.RewardsTreeMode.Value != config.RewardsMode_Download {
		return nil
	}

	// Log
	t.log.Println("Checking for new rewards tree files to download...")

	// Get the current interval
	currentIndex := state.NetworkDetails.RewardIndex
	nodeAddress, _ := t.sp.GetWallet().GetAddress()

	// Check for missing intervals
	missingIntervals := []uint64{}
	for i := uint64(0); i < currentIndex; i++ {
		// Check if the tree file exists
		treeFilePath := t.cfg.GetRewardsTreePath(i)
		_, err := os.Stat(treeFilePath)
		if os.IsNotExist(err) {
			t.log.Printlnf("You are missing the rewards tree file for interval %d.", i)
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
		intervalInfo, err := rprewards.GetIntervalInfo(t.rp, t.cfg, nodeAddress, missingInterval, nil)
		if err != nil {
			return fmt.Errorf("error getting interval %d info: %w", missingInterval, err)
		}
		err = rprewards.DownloadRewardsFile(t.cfg, &intervalInfo)
		if err != nil {
			fmt.Println()
			return err
		}
		fmt.Println("done!")
	}

	return nil
}
