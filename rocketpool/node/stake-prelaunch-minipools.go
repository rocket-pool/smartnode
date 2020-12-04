package node

import (
    "context"
    "errors"
    "fmt"
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
    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
    "github.com/rocket-pool/smartnode/shared/utils/log"
    "github.com/rocket-pool/smartnode/shared/utils/validator"
)


// Settings
const ValidatorContainerSuffix = "_validator"
var stakePrelaunchMinipoolsInterval, _ = time.ParseDuration("5m")
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


// Start stake prelaunch minipools task
func (t *stakePrelaunchMinipools) Start() {
    go (func() {
        for {
            if err := t.run(); err != nil {
                t.log.Println(err)
            }
            time.Sleep(stakePrelaunchMinipoolsInterval)
        }
    })()
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

    // Data
    var wg errgroup.Group
    var withdrawalCredentials common.Hash
    var eth2Config beacon.Eth2Config

    // Get Rocket pool withdrawal credentials
    wg.Go(func() error {
        var err error
        withdrawalCredentials, err = network.GetWithdrawalCredentials(t.rp, nil)
        return err
    })

    // Get eth2 config
    wg.Go(func() error {
        var err error
        eth2Config, err = t.bc.GetEth2Config()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    // Log
    t.log.Printlnf("%d minipools are ready for staking...", len(minipools))

    // Stake minipools
    for _, mp := range minipools {
        if err := t.stakeMinipool(mp, withdrawalCredentials, eth2Config); err != nil {
            t.log.Println(fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err))
        }
    }

    // Restart validator container
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
func (t *stakePrelaunchMinipools) stakeMinipool(mp *minipool.Minipool, withdrawalCredentials common.Hash, eth2Config beacon.Eth2Config) error {

    // Log
    t.log.Printlnf("Staking minipool %s...", mp.Address.Hex())

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
    if err := t.w.Save(); err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Successfully staked minipool %s.", mp.Address.Hex())

    // Return
    return nil

}


// Restart validator container
func (t *stakePrelaunchMinipools) restartValidator() error {

    // Get validator container name
    if t.cfg.Smartnode.ProjectName == "" {
        return errors.New("Rocket Pool docker project name not set")
    }
    containerName := t.cfg.Smartnode.ProjectName + ValidatorContainerSuffix

    // Log
    t.log.Printlnf("Restarting validator container (%s)...", containerName)

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

    // Log
    t.log.Println("Successfully restarted validator container.")

    // Return
    return nil

}

