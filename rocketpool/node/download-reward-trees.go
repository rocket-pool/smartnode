package node

import (
	"fmt"
	"os"

	"github.com/rocket-pool/smartnode/rocketpool/common/log"
	rprewards "github.com/rocket-pool/smartnode/rocketpool/common/rewards"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/state"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// Manage download rewards trees task
type DownloadRewardsTrees struct {
	sp  *services.ServiceProvider
	log log.ColorLogger
}

// Create manage fee recipient task
func NewDownloadRewardsTrees(sp *services.ServiceProvider, logger log.ColorLogger) *DownloadRewardsTrees {
	return &DownloadRewardsTrees{
		sp:  sp,
		log: logger,
	}
}

// Manage fee recipient
func (t *DownloadRewardsTrees) Run(state *state.NetworkState) error {
	// Get services
	cfg := t.sp.GetConfig()
	rp := t.sp.GetRocketPool()
	nodeAddress, _ := t.sp.GetWallet().GetAddress()

	// Check if the user opted into downloading rewards files
	if cfg.Smartnode.RewardsTreeMode.Value.(cfgtypes.RewardsMode) != cfgtypes.RewardsMode_Download {
		return nil
	}

	// Log
	t.log.Println("Checking for new rewards tree files to download...")

	// Get the current interval
	currentIndex := state.NetworkDetails.RewardIndex

	// Check for missing intervals
	missingIntervals := []uint64{}
	for i := uint64(0); i < currentIndex; i++ {
		// Check if the tree file exists
		treeFilePath := cfg.Smartnode.GetRewardsTreePath(i, true)
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
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, missingInterval, nil)
		if err != nil {
			return fmt.Errorf("error getting interval %d info: %w", missingInterval, err)
		}
		err = rprewards.DownloadRewardsFile(cfg, &intervalInfo, true)
		if err != nil {
			fmt.Println()
			return err
		}
		fmt.Println("done!")
	}

	return nil
}
