package node

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Manage download rewards trees task
type DownloadRewardsTrees struct {
	sp     *services.ServiceProvider
	logger *slog.Logger
	cfg    *config.SmartNodeConfig
	rp     *rocketpool.RocketPool
}

// Create manage fee recipient task
func NewDownloadRewardsTrees(sp *services.ServiceProvider, logger *log.Logger) *DownloadRewardsTrees {
	return &DownloadRewardsTrees{
		sp:     sp,
		logger: logger.With(slog.String(keys.RoutineKey, "Rewards Tree Download")),
		cfg:    sp.GetConfig(),
		rp:     sp.GetRocketPool(),
	}
}

// Manage fee recipient
func (t *DownloadRewardsTrees) Run(state *state.NetworkState) error {
	// Check if the user opted into downloading rewards files
	if t.cfg.RewardsTreeMode.Value != config.RewardsMode_Download {
		return nil
	}

	// Log
	t.logger.Info("Starting check for new rewards tree files to download.")

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
			t.logger.Info("You are missing a rewards tree file.", slog.Uint64(keys.IntervalKey, i))
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
		t.logger.Info("Downloading file... ", slog.Uint64(keys.IntervalKey, missingInterval))
		intervalInfo, err := rprewards.GetIntervalInfo(t.rp, t.cfg, nodeAddress, missingInterval, nil)
		if err != nil {
			return fmt.Errorf("error getting interval %d info: %w", missingInterval, err)
		}
		err = rprewards.DownloadRewardsFile(t.cfg, &intervalInfo)
		if err != nil {
			return err
		}
		t.logger.Info("Download successful.")
	}

	return nil
}
