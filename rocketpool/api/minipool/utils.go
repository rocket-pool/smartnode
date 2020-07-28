package minipool

import (
    "bytes"
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Validate that a minipool belongs to a node
func validateMinipoolOwner(mp *minipool.Minipool, nodeAddress common.Address) error {
    owner, err := mp.GetNodeAddress(nil)
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

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var currentBlock int64
    var withdrawalDelay int64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
        return err
    })

    // Get current block
    wg1.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Int64()
        }
        return err
    })

    // Get withdrawal delay
    wg1.Go(func() error {
        var err error
        withdrawalDelay, err = settings.GetMinipoolWithdrawalDelay(rp, nil)
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []api.MinipoolDetails{}, err
    }

    // Data
    var wg2 errgroup.Group
    details := make([]api.MinipoolDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg2.Go(func() error {
            mpDetails, err := getMinipoolDetails(rp, address, currentBlock, withdrawalDelay)
            if err == nil { details[mi] = mpDetails }
            return err
        })
    }

    // Wait for data
    if err := wg2.Wait(); err != nil {
        return []api.MinipoolDetails{}, err
    }

    // Return
    return details, nil

}


// Get a minipool's details
func getMinipoolDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, currentBlock, withdrawalDelay int64) (api.MinipoolDetails, error) {

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
        details.DepositType, err = mp.GetDepositType(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Node, err = mp.GetNodeDetails(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.User, err = mp.GetUserDetails(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Staking, err = mp.GetStakingDetails(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        details.Balances, err = tokens.GetBalances(rp, minipoolAddress, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return api.MinipoolDetails{}, err
    }

    // Update & return
    details.RefundAvailable = (details.Node.RefundBalance.Cmp(big.NewInt(0)) > 0)
    details.WithdrawalAvailable = (details.Status.Status == types.Withdrawable && (currentBlock - details.Status.StatusBlock) >= withdrawalDelay)
    details.CloseAvailable = (details.Status.Status == types.Dissolved)
    return details, nil

}

