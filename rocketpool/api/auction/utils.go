package auction

import (
    "context"
    "math/big"

    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "golang.org/x/sync/errgroup"
)


// Check if bidding has ended for a lot
func getLotBiddingEnded(rp *rocketpool.RocketPool, lotIndex uint64) (bool, error) {

    // Data
    var wg errgroup.Group
    var currentBlock uint64
    var lotEndBlock uint64

    // Get current block
    wg.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get lot end block
    wg.Go(func() error {
        var err error
        lotEndBlock, err = auction.GetLotEndBlock(rp, lotIndex, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (currentBlock >= lotEndBlock), nil

}


// Check whether sufficient remaining RPL is available to create a lot
func getSufficientRemainingRPLForLot(rp *rocketpool.RocketPool) (bool, error) {

    // Data
    var wg errgroup.Group
    var remainingRplBalance *big.Int
    var lotMinimumEthValue *big.Int
    var rplPrice *big.Int

    // Get data
    wg.Go(func() error {
        var err error
        remainingRplBalance, err = auction.GetRemainingRPLBalance(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        lotMinimumEthValue, err = protocol.GetLotMinimumEthValue(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        rplPrice, err = network.GetRPLPrice(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Calculate lot minimum RPL amount
    var tmp big.Int
    var lotMinimumRplAmount big.Int
    tmp.Mul(lotMinimumEthValue, eth.EthToWei(1))
    lotMinimumRplAmount.Quo(&tmp, rplPrice)

    // Return
    return (remainingRplBalance.Cmp(&lotMinimumRplAmount) >= 0), nil

}

