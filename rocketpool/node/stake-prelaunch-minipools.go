package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

// Settings
const ValidatorContainerSuffix = "_validator"
const BeaconContainerSuffix = "_eth2"
var validatorRestartTimeout, _ = time.ParseDuration("5s")


// Stake prelaunch minipools task
type stakePrelaunchMinipools struct {
    c *cli.Context
    log log.ColorLogger
    cfg config.RocketPoolConfig
    w *wallet.Wallet
    rp *rocketpool.RocketPool
    bc beacon.Client
    d *client.Client
}


// Create stake prelaunch minipools task
func newStakePrelaunchMinipools(c *cli.Context, logger log.ColorLogger) (*stakePrelaunchMinipools, error) {

    // Get services
    cfg, err := services.GetConfig(c)
    if err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }
    d, err := services.GetDocker(c)
    if err != nil { return nil, err }

    // Return task
    return &stakePrelaunchMinipools{
        c: c,
        log: logger,
        cfg: cfg,
        w: w,
        rp: rp,
        bc: bc,
        d: d,
    }, nil

}


// Stake prelaunch minipools
func (t *stakePrelaunchMinipools) run() error {

    // Wait for eth client to sync
    if err := services.WaitEthClientSynced(t.c, true); err != nil {
        return err
    }

    // Log
    t.log.Println("Checking for minipools to launch...")

    // Get node account
    nodeAccount, err := t.w.GetNodeAccount()
    if err != nil {
        return err
    }

    // Get prelaunch minipools
    minipools, err := t.getPrelaunchMinipools(nodeAccount.Address)
    if err != nil {
        return err
    }
    if len(minipools) == 0 {
        return nil
    }

    // Get eth2 config
    eth2Config, err := t.bc.GetEth2Config()
    if err != nil {
        return err
    }

    // Log
    t.log.Printlnf("%d minipool(s) are ready for staking...", len(minipools))

    // Stake minipools
    for _, mp := range minipools {
        if err := t.stakeMinipool(mp, eth2Config); err != nil {
            t.log.Println(fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err))
        }
    }

    // Restart validator process
    if err := t.restartValidator(); err != nil {
        return err
    }

    // Return
    return nil

}


// Get prelaunch minipools
func (t *stakePrelaunchMinipools) getPrelaunchMinipools(nodeAddress common.Address) ([]*minipool.Minipool, error) {

    // Get node minipool addresses
    addresses, err := minipool.GetNodeMinipoolAddresses(t.rp, nodeAddress, nil)
    if err != nil {
        return []*minipool.Minipool{}, err
    }

    // Create minipool contracts
    minipools := make([]*minipool.Minipool, len(addresses))
    for mi, address := range addresses {
        mp, err := minipool.NewMinipool(t.rp, address)
        if err != nil {
            return []*minipool.Minipool{}, err
        }
        minipools[mi] = mp
    }

    // Data
    var wg errgroup.Group
    statuses := make([]rptypes.MinipoolStatus, len(minipools))

    // Load minipool statuses
    for mi, mp := range minipools {
        mi, mp := mi, mp
        wg.Go(func() error {
            status, err := mp.GetStatus(nil)
            if err == nil { statuses[mi] = status }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []*minipool.Minipool{}, err
    }

    // Filter minipools by status
    prelaunchMinipools := []*minipool.Minipool{}
    for mi, mp := range minipools {
        if statuses[mi] == rptypes.Prelaunch {
            prelaunchMinipools = append(prelaunchMinipools, mp)
        }
    }

    // Return
    return prelaunchMinipools, nil

}


// Stake a minipool
func (t *stakePrelaunchMinipools) stakeMinipool(mp *minipool.Minipool,  eth2Config beacon.Eth2Config) error {

    // Log
    t.log.Printlnf("Staking minipool %s...", mp.Address.Hex())

    // Get minipool withdrawal credentials
    withdrawalCredentials, err := mp.GetWithdrawalCredentials(nil)
    if err != nil {
        return err
    }

    // Create new validator key
    validatorKey, err := t.w.CreateValidatorKey()
    if err != nil {
        return err
    }

    // Get validator deposit data
    depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config)
    if err != nil {
        return err
    }

    // Get transactor
    opts, err := t.w.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
    signature := rptypes.BytesToValidatorSignature(depositData.Signature)

    // Get the gas estimates
    gasInfo, err := mp.EstimateStakeGas(
        pubKey,
        signature,
        depositDataRoot,
        opts,
    )
    if err != nil {
        return fmt.Errorf("Could not estimate the gas required to stake the minipool: %w", err)
    }
    gasPrice := gasInfo.ReqGasPrice
    if gasPrice == nil {
        gasPrice = gasInfo.EstGasPrice
    }
    
    // Print the total TX cost
    var gas *big.Int 
    if gasInfo.ReqGasLimit != 0 {
        gas = new(big.Int).SetUint64(gasInfo.ReqGasLimit)
    } else {
        gas = new(big.Int).SetUint64(gasInfo.EstGasLimit)
    }
    totalGasWei := new(big.Int).Mul(gasPrice, gas)
    t.log.Printf("Staking the minipool will use a gas price of %.6f Gwei, for a total of %.6f ETH.",
        eth.WeiToGwei(gasPrice),
        math.RoundDown(eth.WeiToEth(totalGasWei), 6))

    // Stake minipool
    hash, err := mp.Stake(
        pubKey,
        signature,
        depositDataRoot,
        opts,
    )
    if err != nil {
        return err
    }

    // Print TX info
    txWatchUrl := t.cfg.Smartnode.TxWatchUrl
    hashString := hash.String()

    t.log.Printf("Transaction has been submitted with hash %s.\n", hashString)
    if txWatchUrl != "" {
        t.log.Printf("You may follow its progress by visiting:\n")
        t.log.Printf("%s/%s\n\n", txWatchUrl, hashString)
    }
    t.log.Println("Waiting for the transaction to be mined...")

    // Wait for the TX to be mined
    if _, err = utils.WaitForTransaction(t.rp.Client, hash); err != nil {
        return err
    }

    // Save wallet
    if err := t.w.Save(); err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Successfully staked minipool %s.", mp.Address.Hex())

    // Return
    return nil

}


// Restart validator process
func (t *stakePrelaunchMinipools) restartValidator() error {

    // Restart validator container
    if isInsideContainer() {

        // Get validator container name & client type label
        var containerName string
        var clientTypeLabel string
        if t.cfg.Smartnode.ProjectName == "" {
            return errors.New("Rocket Pool docker project name not set")
        }
        switch clientType := t.bc.GetClientType(); clientType {
            case beacon.SplitProcess:
                containerName = t.cfg.Smartnode.ProjectName + ValidatorContainerSuffix
                clientTypeLabel = "validator"
            case beacon.SingleProcess:
                containerName = t.cfg.Smartnode.ProjectName + BeaconContainerSuffix
                clientTypeLabel = "beacon"
            default:
                return fmt.Errorf("Can't restart the validator, unknown client type '%d'", clientType)
        }

        // Log
        t.log.Printlnf("Restarting %s container (%s)...", clientTypeLabel, containerName)

        // Get all containers
        containers, err := t.d.ContainerList(context.Background(), types.ContainerListOptions{All: true})
        if err != nil {
            return fmt.Errorf("Could not get docker containers: %w", err)
        }

        // Get validator container ID
        var validatorContainerId string
        for _, container := range containers {
            if container.Names[0] == "/" + containerName {
                validatorContainerId = container.ID
                break
            }
        }
        if validatorContainerId == "" {
            return errors.New("Validator container not found")
        }

        // Restart validator container
        if err := t.d.ContainerRestart(context.Background(), validatorContainerId, &validatorRestartTimeout); err != nil {
            return fmt.Errorf("Could not restart validator container: %w", err)
        }

    // Restart external validator process
    } else {

        // Get validator restart command
        restartCommand := os.ExpandEnv(t.cfg.Smartnode.ValidatorRestartCommand)

        // Log
        t.log.Printlnf("Restarting validator process with command '%s'...", restartCommand)

        // Run validator restart command bound to os stdout/stderr
        cmd := exec.Command(restartCommand)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("Could not restart validator process: %w", err)
        }

    }

    // Log & return
    t.log.Println("Successfully restarted validator")
    return nil

}


// Check if path exists
func pathExists(path string) bool {

    // Check for file info at path
    if _, err := os.Stat(path); err == nil {
        return true;
    }

    // Assume that the path does not exist; this may result in false negatives (e.g. due to permissions)
    return false;

}


// Check whether process is running inside a container
func isInsideContainer() bool {
    containerMarkerPaths := []string {
        "/.dockerenv", // Docker
        "/run/.containerenv", // Podman
    }
    for _, path := range containerMarkerPaths {
        if pathExists(path) {
            return true;
        }
    }
    return false;
}

