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
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
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
    gasThreshold float64
    maxFee *big.Int
    maxPriorityFee *big.Int
    gasLimit uint64
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

    // Check if auto-staking is disabled
    gasThreshold := cfg.Smartnode.RplClaimGasThreshold
    if gasThreshold == 0 {
        logger.Println("RPL claim gas threshold is set to 0, automatic staking of prelaunch minipools will be disabled.")
    }

    // Get the user-requested max fee
    maxFee, err := cfg.GetMaxFee()
    if err != nil {
        return nil, fmt.Errorf("Error getting max fee in configuration: %w", err)
    }

    // Get the user-requested max fee
    maxPriorityFee, err := cfg.GetMaxPriorityFee()
    if err != nil {
        return nil, fmt.Errorf("Error getting max priority fee in configuration: %w", err)
    }
    if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
        logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.");
        maxPriorityFee = big.NewInt(2)
    }

    // Get the user-requested gas limit
    gasLimit, err := cfg.GetGasLimit()
    if err != nil {
        return nil, fmt.Errorf("Error getting gas limit in configuration: %w", err)
    }

    // Return task
    return &stakePrelaunchMinipools{
        c: c,
        log: logger,
        cfg: cfg,
        w: w,
        rp: rp,
        bc: bc,
        d: d,
        gasThreshold: gasThreshold,
        maxFee: maxFee,
        maxPriorityFee: maxPriorityFee,
        gasLimit: gasLimit,
    }, nil

}


// Stake prelaunch minipools
func (t *stakePrelaunchMinipools) run() error {

    // Reload the wallet (in case a call to `node deposit` changed it)
    if err := t.w.Reload(); err != nil {
        return err
    }

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
    successCount := 0
    for _, mp := range minipools {
        success, err := t.stakeMinipool(mp, eth2Config)
        if err != nil {
            t.log.Println(fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err))
            return err
        }
        if success {
            successCount++
        }
    }

    // Restart validator process if any minipools were staked successfully
    if successCount > 0 {
        if err := t.restartValidator(); err != nil {
            return err
        }
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
    statuses := make([]minipool.StatusDetails, len(minipools))

    // Load minipool statuses
    for mi, mp := range minipools {
        mi, mp := mi, mp
        wg.Go(func() error {
            status, err := mp.GetStatusDetails(nil)
            if err == nil { statuses[mi] = status }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []*minipool.Minipool{}, err
    }

    // Get the scrub period
    scrubPeriodSeconds, err := trustednode.GetScrubPeriod(t.rp, nil)
    if err != nil{
        return []*minipool.Minipool{}, err
    }
    scrubPeriod := time.Duration(scrubPeriodSeconds) * time.Second

    // Get the time of the latest block
    latestEth1Block, err := t.rp.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return []*minipool.Minipool{}, fmt.Errorf("Can't get the latest block time: %w", err)
    }
    latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

    // Filter minipools by status
    prelaunchMinipools := []*minipool.Minipool{}
    for mi, mp := range minipools {
        if statuses[mi].Status == rptypes.Prelaunch {
            creationTime := statuses[mi].StatusTime
            remainingTime := creationTime.Add(scrubPeriod).Sub(latestBlockTime)
            if remainingTime < 0 {
                prelaunchMinipools = append(prelaunchMinipools, mp)
            } else {
                t.log.Printlnf("Minipool %s has %s left until it can be staked.", mp.Address.Hex(), remainingTime)
            }
        }
    }

    // Return
    return prelaunchMinipools, nil

}


// Stake a minipool
func (t *stakePrelaunchMinipools) stakeMinipool(mp *minipool.Minipool, eth2Config beacon.Eth2Config) (bool, error) {

    // Log
    t.log.Printlnf("Staking minipool %s...", mp.Address.Hex())

    // Get minipool withdrawal credentials
    withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(t.rp, mp.Address, nil)
    if err != nil {
        return false, err
    }

    // Get the validator key for the minipool
    validatorPubkey, err := minipool.GetMinipoolPubkey(t.rp, mp.Address, nil)
    if err != nil {
        return false, err
    }
    validatorKey, err := t.w.GetValidatorKeyByPubkey(validatorPubkey)
    if err != nil {
        return false, err
    }

    // Get validator deposit data
    depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config)
    if err != nil {
        return false, err
    }

    // Get transactor
    opts, err := t.w.GetNodeAccountTransactor()
    if err != nil {
        return false, err
    }

    // Get the gas limit
    signature := rptypes.BytesToValidatorSignature(depositData.Signature)
    gasInfo, err := mp.EstimateStakeGas(signature, depositDataRoot, opts)
    if err != nil {
        return false, fmt.Errorf("Could not estimate the gas required to stake the minipool: %w", err)
    }
    var gas *big.Int 
    if t.gasLimit != 0 {
        gas = new(big.Int).SetUint64(t.gasLimit)
    } else {
        gas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
    }

    // Get the max fee
    maxFee := t.maxFee
    if maxFee == nil || maxFee.Uint64() == 0 {
        maxFee, err = rpgas.GetHeadlessMaxFeeWei()
        if err != nil {
            return false, err
        }
    }
    
    // Print the gas info
    if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
        // Check for the timeout buffer
        prelaunchTime, err := mp.GetStatusTime(nil)
        if err != nil {
            t.log.Printlnf("Error checking minipool launch time: %s\nStaking now for safety...", err.Error())
        }
        isDue, timeUntilDue, err := api.IsTransactionDue(t.rp, prelaunchTime)
        if err != nil {
            t.log.Printlnf("Error checking if minipool is due: %s\nStaking now for safety...", err.Error())
        }
        if !isDue {
            t.log.Printlnf("Time until staking will be forced for safety: %s", timeUntilDue)
            return false, nil
        } else {
            t.log.Println("NOTICE: The minipool has exceeded half of the timeout period, so it will be force-staked at the current gas price.")
        }
    }

    opts.GasFeeCap = maxFee
    opts.GasTipCap = t.maxPriorityFee
    opts.GasLimit = gas.Uint64()

    // Stake minipool
    hash, err := mp.Stake(
        signature,
        depositDataRoot,
        opts,
    )
    if err != nil {
        return false, err
    }

    // Print TX info and wait for it to be mined
    err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
    if err != nil {
        return false, err
    }

    // Log
    t.log.Printlnf("Successfully staked minipool %s.", mp.Address.Hex())

    // Return
    return true, nil

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

