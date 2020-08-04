package watchtower

import (
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


// Settings
var submitWithdrawableMinipoolsInterval, _ = time.ParseDuration("1m")


// Withdrawable minipool info
type minipoolWithdrawableDetails struct {
    Address common.Address
    StartBalance *big.Int
    EndBalance *big.Int
    Withdrawable bool
}


// Start submit withdrawable minipools task
func startSubmitWithdrawableMinipools(c *cli.Context) error {

    // Get services
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return err }

    // Submit withdrawable minipools at interval
    go (func() {
        for {
            if err := submitWithdrawableMinipools(c, am, rp, bc); err != nil {
                log.Println(err)
            }
            time.Sleep(submitWithdrawableMinipoolsInterval)
        }
    })()

    // Return
    return nil

}


// Submit withdrawable minipools
func submitWithdrawableMinipools(c *cli.Context, am *accounts.AccountManager, rp *rocketpool.RocketPool, bc beacon.Client) error {

    // Wait for eth client to sync
    if err := services.WaitClientSynced(c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return err
    }

    // Data
    var wg errgroup.Group
    var nodeTrusted bool
    var submitWithdrawableEnabled bool

    // Get data
    wg.Go(func() error {
        var err error
        nodeTrusted, err = node.GetNodeTrusted(rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        submitWithdrawableEnabled, err = settings.GetMinipoolSubmitWithdrawableEnabled(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    // Check node trusted status & settings
    if !(nodeTrusted && submitWithdrawableEnabled) {
        return nil
    }

    // Get minipool withdrawable details
    minipools, err := getNetworkMinipoolWithdrawableDetails(rp, bc, nodeAccount.Address)
    if err != nil {
        return err
    }
    if len(minipools) == 0 {
        return nil
    }

    // Log
    log.Printf("%d minipools are withdrawable...\n", len(minipools))

    // Submit minipools withdrawable status
    for _, details := range minipools {
        if err := submitWithdrawableMinipool(am, rp, details); err != nil {
            log.Println(fmt.Errorf("Could not submit minipool %s withdrawable status: %w", details.Address.Hex(), err))
        }
    }

    // Return
    return nil

}


// Get all minipool withdrawable details
func getNetworkMinipoolWithdrawableDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address) ([]minipoolWithdrawableDetails, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var eth2Config beacon.Eth2Config
    var beaconHead beacon.BeaconHead

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(rp, nil)
        return err
    })

    // Get eth2 config
    wg1.Go(func() error {
        var err error
        eth2Config, err = bc.GetEth2Config()
        return err
    })

    // Get beacon head
    wg1.Go(func() error {
        var err error
        beaconHead, err = bc.GetBeaconHead()
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []minipoolWithdrawableDetails{}, err
    }

    // Data
    var wg2 errgroup.Group
    minipools := make([]minipoolWithdrawableDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg2.Go(func() error {
            mpDetails, err := getMinipoolWithdrawableDetails(rp, bc, nodeAddress, address, eth2Config, beaconHead)
            if err == nil { minipools[mi] = mpDetails }
            return err
        })
    }

    // Wait for data
    if err := wg2.Wait(); err != nil {
        return []minipoolWithdrawableDetails{}, err
    }

    // Filter by withdrawable status
    withdrawableMinipools := []minipoolWithdrawableDetails{}
    for _, details := range minipools {
        if details.Withdrawable {
            withdrawableMinipools = append(withdrawableMinipools, details)
        }
    }

    // Return
    return withdrawableMinipools, nil

}


// Get minipool withdrawable details
func getMinipoolWithdrawableDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, minipoolAddress common.Address, eth2Config beacon.Eth2Config, beaconHead beacon.BeaconHead) (minipoolWithdrawableDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return minipoolWithdrawableDetails{}, err
    }

    // Data
    var wg errgroup.Group
    var status types.MinipoolStatus
    var userDepositTime uint64
    var pubkey types.ValidatorPubkey

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus(nil)
        return err
    })
    wg.Go(func() error {
        userDepositAssignedTime, err := mp.GetUserDepositAssignedTime(nil)
        if err == nil {
            userDepositTime = uint64(userDepositAssignedTime.Unix())
        }
        return err
    })
    wg.Go(func() error {
        var err error
        pubkey, err = minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return minipoolWithdrawableDetails{}, err
    }

    // Check minipool status
    if status != types.Staking {
        return minipoolWithdrawableDetails{}, nil
    }

    // Get & check validator status
    validator, err := bc.GetValidatorStatus(pubkey, nil)
    if err != nil {
        return minipoolWithdrawableDetails{}, err
    }
    if !validator.Exists || validator.WithdrawableEpoch > beaconHead.Epoch {
        return minipoolWithdrawableDetails{}, nil
    }

    // Get start epoch
    startEpoch := epochAt(eth2Config, userDepositTime)
    if startEpoch < validator.ActivationEpoch {
        startEpoch = validator.ActivationEpoch
    } else if startEpoch > beaconHead.Epoch {
        startEpoch = beaconHead.Epoch
    }

    // Get validator status at start epoch
    validatorStart, err := bc.GetValidatorStatus(pubkey, &beacon.ValidatorStatusOptions{Epoch: startEpoch})
    if err != nil {
        return minipoolWithdrawableDetails{}, err
    }
    if !validatorStart.Exists {
        return minipoolWithdrawableDetails{}, fmt.Errorf("Could not get validator %s balance at epoch %d", pubkey.Hex(), startEpoch)
    }

    // Get validator balances at start epoch and current epoch
    startBalance := eth.GweiToWei(float64(validatorStart.Balance))
    endBalance := eth.GweiToWei(float64(validator.Balance))

    // Check for existing node submission
    nodeSubmittedMinipool, err := rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("minipool.withdrawable.submitted.node"), nodeAddress.Bytes(), minipoolAddress.Bytes(), startBalance.Bytes(), endBalance.Bytes()))
    if err != nil {
        return minipoolWithdrawableDetails{}, err
    }
    if nodeSubmittedMinipool {
        return minipoolWithdrawableDetails{}, nil
    }

    // Return
    return minipoolWithdrawableDetails{
        Address: minipoolAddress,
        StartBalance: startBalance,
        EndBalance: endBalance,
        Withdrawable: true,
    }, nil

}


// Submit minipool withdrawable status
func submitWithdrawableMinipool(am *accounts.AccountManager, rp *rocketpool.RocketPool, details minipoolWithdrawableDetails) error {

    // Log
    log.Printf("Submitting minipool %s withdrawable status...\n", details.Address.Hex())

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Dissolve
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, details.Address, details.StartBalance, details.EndBalance, opts); err != nil {
        return err
    }

    // Log
    log.Printf("Successfully submitted minipool %s withdrawable status.\n", details.Address.Hex())

    // Return
    return nil

}

