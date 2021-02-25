package minipool

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "golang.org/x/sync/errgroup"
    "github.com/urfave/cli"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)


func getLeader(c *cli.Context) (*api.MinipoolStatusResponse, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Response
    response := api.MinipoolStatusResponse{}

    details, err := GetAllMinipoolDetails(rp, bc)
    if err != nil {
        return nil, err
    }
    response.Minipools = details

    // Return response
    return &response, nil
}


func GetAllMinipoolDetails(rp *rocketpool.RocketPool, bc beacon.Client) ([]api.MinipoolDetails, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var currentEpoch uint64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(rp, nil)
        return err
    })

    // Get current epoch
    wg1.Go(func() error {
        head, err := bc.GetBeaconHead()
        if err == nil {
            currentEpoch = head.Epoch
        }
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []api.MinipoolDetails{}, err
    }

    // Load details in batches
    details := make([]api.MinipoolDetails, len(addresses))
    for bsi := 0; bsi < len(addresses); bsi += MinipoolDetailsBatchSize {

        // Get batch start & end index
        msi := bsi
        mei := bsi + MinipoolDetailsBatchSize
        if mei > len(addresses) { mei = len(addresses) }

        // Get minipool validator statuses
        validators, err := rputils.GetMinipoolValidators(rp, bc, addresses[msi:mei], nil, nil)
        if err != nil {
            continue
        }

        // Load details
        var wg errgroup.Group
        for mi := msi; mi < mei; mi++ {
            mi := mi
            wg.Go(func() error {
                address := addresses[mi]
                validator := validators[address]
                mpDetails, err := getMinipoolBalance(rp, address, validator, currentEpoch)
                if err == nil { details[mi] = mpDetails }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []api.MinipoolDetails{}, err
        }

    }

    // Return
    return details, nil
}


func getMinipoolBalance(rp *rocketpool.RocketPool, minipoolAddress common.Address, validator beacon.ValidatorStatus, currentEpoch uint64) (api.MinipoolDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return api.MinipoolDetails{}, err
    }

    // Data
    var wg errgroup.Group
    details := api.MinipoolDetails { Address: minipoolAddress }

    // Load data
    wg.Go(func() error {
        var err error
        details.ValidatorPubkey, err = minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Status, err = mp.GetStatusDetails(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Node.Address, err = mp.GetNodeAddress(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Node.DepositBalance, err = mp.GetNodeDepositBalance(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.User.DepositBalance, err = mp.GetUserDepositBalance(nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return api.MinipoolDetails{}, err
    }

    // Get validator details if staking
    if details.Status.Status == types.Staking {
        // Validator details
        validatorDetails := api.ValidatorDetails{}

        // Set validator status details
        validatorActivated := false
        if validator.Exists {
            validatorDetails.Exists = true
            validatorDetails.Active = (validator.ActivationEpoch < currentEpoch && validator.ExitEpoch > currentEpoch)
            validatorActivated = (validator.ActivationEpoch < currentEpoch)
        }

        // use deposit balances if validator not activated
        if !validatorActivated {
            validatorDetails.Balance = new(big.Int)
            validatorDetails.Balance.Add(details.Node.DepositBalance, details.User.DepositBalance)
        } else {
            // Set validator balance
            validatorDetails.Balance = eth.GweiToWei(float64(validator.Balance))
        }
        details.Validator = validatorDetails
    }

    return details, nil
}
