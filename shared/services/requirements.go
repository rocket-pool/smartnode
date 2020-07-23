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
const ClientSyncTimeout = 15 // 15 seconds
var checkNodePasswordInterval, _ = time.ParseDuration("15s")
var checkNodeAccountInterval, _ = time.ParseDuration("15s")
var checkRocketStorageInterval, _ = time.ParseDuration("15s")
var checkNodeRegisteredInterval, _ = time.ParseDuration("15s")


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


func RequireNodeAccount(c *cli.Context) error {
    if err := RequireNodePassword(c); err != nil {
        return err
    }
    nodeAccountExists, err := getNodeAccountExists(c)
    if err != nil {
        return err
    }
    if !nodeAccountExists {
        return errors.New("The node account has not been created. Please initialize the node and try again.")
    }
    return nil
}


func RequireClientSynced(c *cli.Context) error {
    clientSynced, err := waitClientSynced(c, false, ClientSyncTimeout)
    if err != nil {
        return err
    }
    if !clientSynced {
        return errors.New("The Eth 1.0 node is currently syncing. Please try again later.")
    }
    return nil
}


func RequireRocketStorage(c *cli.Context) error {
    if err := RequireClientSynced(c); err != nil {
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
    if err := RequireNodeAccount(c); err != nil {
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


func WaitNodeAccount(c *cli.Context, verbose bool) error {
    if err := WaitNodePassword(c, verbose); err != nil {
        return err
    }
    for {
        nodeAccountExists, err := getNodeAccountExists(c)
        if err != nil {
            return err
        }
        if nodeAccountExists {
            return nil
        }
        if verbose {
            fmt.Printf("The node account has not been created, retrying in %s...\n", checkNodeAccountInterval.String())
        }
        time.Sleep(checkNodeAccountInterval)
    }
}


func WaitClientSynced(c *cli.Context, verbose bool) error {
    if _, err := waitClientSynced(c, verbose, 0); err != nil {
        return err
    }
    return nil
}


func WaitRocketStorage(c *cli.Context, verbose bool) error {
    if err := WaitClientSynced(c, verbose); err != nil {
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
    if err := WaitNodeAccount(c, verbose); err != nil {
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
    return pm.PasswordExists(), nil
}


// Check if the node account exists
func getNodeAccountExists(c *cli.Context) (bool, error) {
    am, err := GetAccountManager(c)
    if err != nil {
        return false, err
    }
    return am.NodeAccountExists(), nil
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
    am, err := GetAccountManager(c)
    if err != nil {
        return false, err
    }
    rp, err := GetRocketPool(c)
    if err != nil {
        return false, err
    }
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return false, err
    }
    return node.GetNodeExists(rp, nodeAccount.Address)
}


// Wait for the eth client to sync
// timeout of 0 indicates no timeout
func waitClientSynced(c *cli.Context, verbose bool, timeout int64) (bool, error) {

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
        if progress == nil {
            if statusPrinted { fmt.Print("\n") }
            return true, nil
        } else {
            if statusPrinted { fmt.Print("\r") }
            if verbose {
                fmt.Printf("Node syncing: %.2f%%  ", (float64(progress.CurrentBlock - progress.StartingBlock) / float64(progress.HighestBlock - progress.StartingBlock)) * 100)
                statusPrinted = true
            }
        }

    }

}

