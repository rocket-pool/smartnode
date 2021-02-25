package auction

import (
    "context"

    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
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

