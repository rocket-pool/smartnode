package services

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/urfave/cli"
)


// Settings
const EthClientSyncTimeout = 15 // 15 seconds
const BeaconClientSyncTimeout = 15 // 15 seconds
var checkNodePasswordInterval, _ = time.ParseDuration("15s")
var checkNodeWalletInterval, _ = time.ParseDuration("15s")
var checkRocketStorageInterval, _ = time.ParseDuration("15s")
var checkNodeRegisteredInterval, _ = time.ParseDuration("15s")
var ethClientSyncPollInterval, _ = time.ParseDuration("2s")
var beaconClientSyncPollInterval, _ = time.ParseDuration("2s")


//
// Service requirements
//


func RequireNodePassword(c *cli.Context) error {
    nodePasswordSet, err := getNodePasswordSet(c)
    if err != nil {
        return err
    }
    if !nodePasswordSet {
        return errors.New("The node password has not been set. Please initialize the node and try again.")
    }
    return nil
}


func RequireNodeWallet(c *cli.Context) error {
    if err := RequireNodePassword(c); err != nil {
        return err
    }
    nodeWalletInitialized, err := getNodeWalletInitialized(c)
    if err != nil {
        return err
    }
    if !nodeWalletInitialized {
        return errors.New("The node wallet has not been initialized. Please initialize the node and try again.")
    }
    return nil
}


func RequireEthClientSynced(c *cli.Context) error {
    ethClientSynced, err := waitEthClientSynced(c, false, EthClientSyncTimeout)
    if err != nil {
        return err
    }
    if !ethClientSynced {
        return errors.New("The Eth 1.0 node is currently syncing. Please try again later.")
    }
    return nil
}


func RequireBeaconClientSynced(c *cli.Context) error {
    beaconClientSynced, err := waitBeaconClientSynced(c, false, BeaconClientSyncTimeout)
    if err != nil {
        return err
    }
    if !beaconClientSynced {
        return errors.New("The Eth 2.0 node is currently syncing. Please try again later.")
    }
    return nil
}


func RequireRocketStorage(c *cli.Context) error {
    if err := RequireEthClientSynced(c); err != nil {
        return err
    }
    rocketStorageLoaded, err := getRocketStorageLoaded(c)
    if err != nil {
        return err
    }
    if !rocketStorageLoaded {
        return errors.New("The Rocket Pool storage contract was not found; the configured address may be incorrect, or the Eth 1.0 node may not be synced. Please try again later.")
    }
    return nil
}


func RequireNodeRegistered(c *cli.Context) error {
    if err := RequireNodeWallet(c); err != nil {
        return err
    }
    if err := RequireRocketStorage(c); err != nil {
        return err
    }
    nodeRegistered, err := getNodeRegistered(c)
    if err != nil {
        return err
    }
    if !nodeRegistered {
        return errors.New("The node is not registered with Rocket Pool. Please register and try again.")
    }
    return nil
}


//
// Service synchronization
//


func WaitNodePassword(c *cli.Context, verbose bool) error {
    for {
        nodePasswordSet, err := getNodePasswordSet(c)
        if err != nil {
            return err
        }
        if nodePasswordSet {
            return nil
        }
        if verbose {
            fmt.Printf("The node password has not been set, retrying in %s...\n", checkNodePasswordInterval.String())
        }
        time.Sleep(checkNodePasswordInterval)
    }
}


func WaitNodeWallet(c *cli.Context, verbose bool) error {
    if err := WaitNodePassword(c, verbose); err != nil {
        return err
    }
    for {
        nodeWalletInitialized, err := getNodeWalletInitialized(c)
        if err != nil {
            return err
        }
        if nodeWalletInitialized {
            return nil
        }
        if verbose {
            fmt.Printf("The node wallet has not been initialized, retrying in %s...\n", checkNodeWalletInterval.String())
        }
        time.Sleep(checkNodeWalletInterval)
    }
}


func WaitEthClientSynced(c *cli.Context, verbose bool) error {
    _, err := waitEthClientSynced(c, verbose, 0)
    return err
}


func WaitBeaconClientSynced(c *cli.Context, verbose bool) error {
    _, err := waitBeaconClientSynced(c, verbose, 0)
    return err
}


func WaitRocketStorage(c *cli.Context, verbose bool) error {
    if err := WaitEthClientSynced(c, verbose); err != nil {
        return err
    }
    for {
        rocketStorageLoaded, err := getRocketStorageLoaded(c)
        if err != nil {
            return err
        }
        if rocketStorageLoaded {
            return nil
        }
        if verbose {
            fmt.Printf("The Rocket Pool storage contract was not found, retrying in %s...\n", checkRocketStorageInterval.String())
        }
        time.Sleep(checkRocketStorageInterval)
    }
}


func WaitNodeRegistered(c *cli.Context, verbose bool) error {
    if err := WaitNodeWallet(c, verbose); err != nil {
        return err
    }
    if err := WaitRocketStorage(c, verbose); err != nil {
        return err
    }
    for {
        nodeRegistered, err := getNodeRegistered(c)
        if err != nil {
            return err
        }
        if nodeRegistered {
            return nil
        }
        if verbose {
            fmt.Printf("The node is not registered with Rocket Pool, retrying in %s...\n", checkNodeRegisteredInterval.String())
        }
        time.Sleep(checkNodeRegisteredInterval)
    }
}


//
// Helpers
//


// Check if the node password is set
func getNodePasswordSet(c *cli.Context) (bool, error) {
    pm, err := GetPasswordManager(c)
    if err != nil {
        return false, err
    }
    return pm.IsPasswordSet(), nil
}


// Check if the node wallet is initialized
func getNodeWalletInitialized(c *cli.Context) (bool, error) {
    w, err := GetWallet(c)
    if err != nil {
        return false, err
    }
    return w.IsInitialized(), nil
}


// Check if the RocketStorage contract is loaded
func getRocketStorageLoaded(c *cli.Context) (bool, error) {
    cfg, err := GetConfig(c)
    if err != nil {
        return false, err
    }
    ec, err := GetEthClient(c)
    if err != nil {
        return false, err
    }
    code, err := ec.CodeAt(context.Background(), common.HexToAddress(cfg.Rocketpool.StorageAddress), nil)
    if err != nil {
        return false, err
    }
    return (len(code) > 0), nil
}


// Check if the node is registered
func getNodeRegistered(c *cli.Context) (bool, error) {
    w, err := GetWallet(c)
    if err != nil {
        return false, err
    }
    rp, err := GetRocketPool(c)
    if err != nil {
        return false, err
    }
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return false, err
    }
    return node.GetNodeExists(rp, nodeAccount.Address, nil)
}


// Wait for the eth client to sync
// timeout of 0 indicates no timeout
func waitEthClientSynced(c *cli.Context, verbose bool, timeout int64) (bool, error) {

    // Get eth client
    ec, err := GetEthClient(c)
    if err != nil {
        return false, err
    }

    // Get start settings
    startTime := time.Now().Unix()
    statusPrinted := false

    // Wait for sync
    for {

        // Check timeout
        if (timeout > 0) && (time.Now().Unix() - startTime > timeout) {
            return false, nil
        }

        // Get sync progress
        progress, err := ec.SyncProgress(context.Background())
        if err != nil {
            return false, err
        }

        // Check sync progress
        if progress != nil {
            if verbose {
                if statusPrinted { fmt.Print("\r") }
                fmt.Printf("Eth 1.0 node syncing: %.2f%%   ", (float64(progress.CurrentBlock - progress.StartingBlock) / float64(progress.HighestBlock - progress.StartingBlock)) * 100)
                statusPrinted = true
            }
        } else {
            if statusPrinted { fmt.Print("\n") }
            return true, nil
        }

        // Pause before next poll
        time.Sleep(ethClientSyncPollInterval)

    }

}


// Wait for the beacon client to sync
// timeout of 0 indicates no timeout
func waitBeaconClientSynced(c *cli.Context, verbose bool, timeout int64) (bool, error) {

    // Get beacon client
    bc, err := GetBeaconClient(c)
    if err != nil {
        return false, err
    }

    // Get start settings
    startTime := time.Now().Unix()
    statusPrinted := false

    // Wait for sync
    for {

        // Check timeout
        if (timeout > 0) && (time.Now().Unix() - startTime > timeout) {
            return false, nil
        }

        // Get sync status
        syncStatus, err := bc.GetSyncStatus()
        if err != nil {
            return false, err
        }

        // Check sync status
        if syncStatus.Syncing {
            if verbose {
                if statusPrinted { fmt.Print("\r") }
                fmt.Print("Eth 2.0 node syncing...   ")
                statusPrinted = true
            }
        } else {
            if statusPrinted { fmt.Print("\n") }
            return true, nil
        }

        // Pause before next poll
        time.Sleep(beaconClientSyncPollInterval)

    }

}

