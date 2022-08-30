package node

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
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
func (d *downloadRewardsTrees) run() error {

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

	// Get the unclaimed intervals
	unclaimed, _, err := rprewards.GetClaimStatus(d.rp, nodeAccount.Address)
	if err != nil {
		return err
	}

	// Check for missing intervals
	missingIntervals := []rprewards.IntervalInfo{}
	for _, interval := range unclaimed {
		intervalInfo, err := rprewards.GetIntervalInfo(d.rp, d.cfg, nodeAccount.Address, interval)
		if err != nil {
			return err
		}
		if !intervalInfo.TreeFileExists {
			d.log.Printlnf("You are missing the rewards tree file for interval %d.", intervalInfo.Index)
			missingIntervals = append(missingIntervals, intervalInfo)
		}
	}

	if len(missingIntervals) == 0 {
		return nil
	}

	// Download missing intervals
	for _, missingInterval := range missingIntervals {
		fmt.Printf("Downloading interval %d file... ", missingInterval.Index)
		err := rprewards.DownloadRewardsFile(d.cfg, missingInterval.Index, missingInterval.CID, true)
		if err != nil {
			fmt.Println()
			return err
		}
		fmt.Println("done!")
	}

	return nil

}

// Downloads a single rewards file
func (d *downloadRewardsTrees) downloadRewardsFile(interval uint64, cid string, compressedUrls []string, uncompressedUrls []string) ([]byte, error) {

	for i, url := range compressedUrls {
		d.log.Printf("Downloading %s... ", url)
		resp, err := http.Get(url)
		if err != nil {
			d.log.Printlnf("failed (%s)", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			// Not found, try the uncompressed URL
			// NOTE: this can go after Kiln
			uncompressedUrl := uncompressedUrls[i]
			d.log.Printf("not found, trying uncompressed URL (%s)... ", uncompressedUrl)
			resp, err = http.Get(url)
			if err != nil {
				d.log.Printlnf("failed (%s)", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				d.log.Println("failed with status %s", resp.Status)
				continue
			} else {
				// Got it uncompressed, return the body
				bytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					d.log.Printlnf("error reading response bytes: %s", err.Error())
					continue
				}

				d.log.Println("done!")
				return bytes, nil
			}

		} else if resp.StatusCode != http.StatusOK {
			d.log.Printlnf("failed with status %s", resp.Status)
			continue
		} else {
			// If we got here, we have a successful download
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				d.log.Printlnf("error reading response bytes: %s", err.Error())
				continue
			}

			// Decompress it
			decompressedBytes, err := d.decompressFile(bytes)
			if err != nil {
				d.log.Println(err.Error())
				continue
			}

			d.log.Println("done!")
			return decompressedBytes, nil
		}
	}

	return nil, fmt.Errorf("Error downloading rewards file for interval %d: all URLs failed.", interval)

}

// Decompresses a rewards file
func (d *downloadRewardsTrees) decompressFile(compressedBytes []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating compression decoder: %w", err)
	}

	decompressedBytes, err := decoder.DecodeAll(compressedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("error decompressing rewards file: %w", err)
	}

	return decompressedBytes, nil
}
