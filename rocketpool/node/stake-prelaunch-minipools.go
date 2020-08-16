package node

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    rptypes "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
    "github.com/rocket-pool/smartnode/shared/utils/validator"
)


// Settings
const ValidatorContainerName = "rocketpool_validator_1"
var stakePrelaunchMinipoolsInterval, _ = time.ParseDuration("1m")
var validatorRestartTimeout, _ = time.ParseDuration("5s")


// Start stake prelaunch minipools task
func startStakePrelaunchMinipools(c *cli.Context) error {

    // Get services
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }
    w, err := services.GetWallet(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return err }
    d, err := services.GetDocker(c)
    if err != nil { return err }

    // Stake prelaunch minipools at interval
    go (func() {
        for {
            if err := stakePrelaunchMinipools(c, w, rp, bc, d); err != nil {
                log.Println(err)
            }
            time.Sleep(stakePrelaunchMinipoolsInterval)
        }
    })()

    // Return
    return nil

}


// Stake prelaunch minipools
func stakePrelaunchMinipools(c *cli.Context, w *wallet.Wallet, rp *rocketpool.RocketPool, bc beacon.Client, d *client.Client) error {

    // Wait for eth client to sync
    if err := services.WaitEthClientSynced(c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return err
    }

    // Get prelaunch minipools
    minipools, err := getPrelaunchMinipools(rp, nodeAccount.Address)
    if err != nil {
        return err
    }
    if len(minipools) == 0 {
        return nil
    }

    // Data
    var wg errgroup.Group
    var withdrawalCredentials common.Hash
    var eth2Config beacon.Eth2Config

    // Get Rocket pool withdrawal credentials
    wg.Go(func() error {
        var err error
        withdrawalCredentials, err = network.GetWithdrawalCredentials(rp, nil)
        return err
    })

    // Get eth2 config
    wg.Go(func() error {
        var err error
        eth2Config, err = bc.GetEth2Config()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    // Log
    log.Printf("%d minipools are ready for staking...\n", len(minipools))

    // Stake minipools
    for _, mp := range minipools {
        if err := stakeMinipool(w, mp, withdrawalCredentials, eth2Config); err != nil {
            log.Println(fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err))
        }
    }

    // Restart validator container
    if err := restartValidator(d); err != nil {
        return err
    }

    // Return
    return nil

}


// Get prelaunch minipools
func getPrelaunchMinipools(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]*minipool.Minipool, error) {

    // Get node minipool addresses
    addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
    if err != nil {
        return []*minipool.Minipool{}, err
    }

    // Create minipool contracts
    minipools := make([]*minipool.Minipool, len(addresses))
    for mi, address := range addresses {
        mp, err := minipool.NewMinipool(rp, address)
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
func stakeMinipool(w *wallet.Wallet, mp *minipool.Minipool, withdrawalCredentials common.Hash, eth2Config beacon.Eth2Config) error {

    // Log
    log.Printf("Staking minipool %s...\n", mp.Address.Hex())

    // Create new validator key
    validatorKey, err := w.CreateValidatorKey()
    if err != nil {
        return err
    }

    // Get validator deposit data
    depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config)
    if err != nil {
        return err
    }

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Stake minipool
    if _, err := mp.Stake(
        rptypes.BytesToValidatorPubkey(depositData.PublicKey),
        rptypes.BytesToValidatorSignature(depositData.Signature),
        depositDataRoot,
        opts,
    ); err != nil {
        return err
    }

    // Save wallet
    if err := w.Save(); err != nil {
        return err
    }

    // Log
    log.Printf("Successfully staked minipool %s.\n", mp.Address.Hex())

    // Return
    return nil

}


// Restart validator container
func restartValidator(d *client.Client) error {

    // Log
    log.Println("Restarting validator container...")

    // Get all containers
    containers, err := d.ContainerList(context.Background(), types.ContainerListOptions{All: true})
    if err != nil {
        return fmt.Errorf("Could not get docker containers: %w", err)
    }

    // Get validator container ID
    var validatorContainerId string
    for _, container := range containers {
        if container.Names[0] == "/" + ValidatorContainerName {
            validatorContainerId = container.ID
            break
        }
    }
    if validatorContainerId == "" {
        return errors.New("Validator container not found")
    }

    // Restart validator container
    if err := d.ContainerRestart(context.Background(), validatorContainerId, &validatorRestartTimeout); err != nil {
        return fmt.Errorf("Could not restart validator container: %w", err)
    }

    // Log
    log.Println("Successfully restarted validator container...")

    // Return
    return nil

}

