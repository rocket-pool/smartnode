package minipool

import (
    "bytes"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Validate that a minipool belongs to a node
func validateMinipoolOwner(mp *minipool.Minipool, nodeAddress common.Address) error {
    owner, err := mp.GetNodeAddress()
    if err != nil {
        return err
    }
    if !bytes.Equal(owner.Bytes(), nodeAddress.Bytes()) {
        return fmt.Errorf("Minipool %s does not belong to the node", mp.Address.Hex())
    }
    return nil
}


// Get all node minipool details
func getNodeMinipoolDetails(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]api.MinipoolDetails, error) {

    // Get minipool addresses
    addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAddress)
    if err != nil {
        return []api.MinipoolDetails{}, err
    }

    // Data
    var wg errgroup.Group
    details := make([]api.MinipoolDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg.Go(func() error {
            mpDetails, err := getMinipoolDetails(rp, address)
            if err == nil { details[mi] = mpDetails }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []api.MinipoolDetails{}, err
    }

    // Return
    return details, nil

}


// Get a minipool's details
func getMinipoolDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address) (api.MinipoolDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return api.MinipoolDetails{}, err
    }

    // Data
    var wg errgroup.Group
    details := api.MinipoolDetails{Address: minipoolAddress}

    // Load data
    wg.Go(func() error {
        var err error
        details.ValidatorPubkey, err = minipool.GetMinipoolPubkey(rp, minipoolAddress)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Status, err = mp.GetStatusDetails()
        return err
    })
    wg.Go(func() error {
        var err error
        details.DepositType, err = mp.GetDepositType()
        return err
    })
    wg.Go(func() error {
        var err error
        details.Node, err = mp.GetNodeDetails()
        return err
    })
    wg.Go(func() error {
        var err error
        details.User, err = mp.GetUserDetails()
        return err
    })
    wg.Go(func() error {
        var err error
        details.Staking, err = mp.GetStakingDetails()
        return err
    })
    wg.Go(func() error {
        var err error
        details.Balances, err = tokens.GetBalances(rp, minipoolAddress)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return api.MinipoolDetails{}, err
    }

    // Return
    return details, nil

}

