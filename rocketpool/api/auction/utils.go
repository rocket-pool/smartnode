package auction

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "golang.org/x/sync/errgroup"
)


// Settings
const LotCountDetailsBatchSize = 10


// Lot count details
type lotCountDetails struct {
    AddressHasBid bool
    BiddingEnded bool
    HasRemainingRpl bool
    RplRecovered bool
}


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


// Get all lot count details
func getAllLotCountDetails(rp *rocketpool.RocketPool, bidderAddress common.Address) ([]lotCountDetails, error) {

    // Data
    var wg1 errgroup.Group
    var lotCount uint64
    var currentBlock uint64

    // Get lot count
    wg1.Go(func() error {
        var err error
        lotCount, err = auction.GetLotCount(rp, nil)
        return err
    })

    // Get current block
    wg1.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []lotCountDetails{}, err
    }

    // Load details in batches
    details := make([]lotCountDetails, lotCount)
    for bsi := uint64(0); bsi < lotCount; bsi += LotCountDetailsBatchSize {

        // Get batch start & end index
        lsi := bsi
        lei := bsi + LotCountDetailsBatchSize
        if lei > lotCount { lei = lotCount }

        // Load details
        var wg errgroup.Group
        for li := lsi; li < lei; li++ {
            li := li
            wg.Go(func() error {
                lotDetails, err := getLotCountDetails(rp, bidderAddress, li, currentBlock)
                if err == nil { details[li] = lotDetails }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []lotCountDetails{}, err
        }

    }

    // Return
    return details, nil

}


// Get a lot's count details
func getLotCountDetails(rp *rocketpool.RocketPool, bidderAddress common.Address, lotIndex, currentBlock uint64) (lotCountDetails, error) {

    // Data
    var wg errgroup.Group
    var addressBidAmount *big.Int
    var endBlock uint64
    var remainingRpl *big.Int
    var rplRecovered bool

    // Get address bid amount
    wg.Go(func() error {
        var err error
        addressBidAmount, err = auction.GetLotAddressBidAmount(rp, lotIndex, bidderAddress, nil)
        return err
    })

    // Get lot end block
    wg.Go(func() error {
        var err error
        endBlock, err = auction.GetLotEndBlock(rp, lotIndex, nil)
        return err
    })

    // Get lot remaining RPL amount
    wg.Go(func() error {
        var err error
        remainingRpl, err = auction.GetLotRemainingRPLAmount(rp, lotIndex, nil)
        return err
    })

    // Get lot RPL recovered status
    wg.Go(func() error {
        var err error
        rplRecovered, err = auction.GetLotRPLRecovered(rp, lotIndex, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return lotCountDetails{}, err
    }

    // Return
    return lotCountDetails{
        AddressHasBid: (addressBidAmount.Cmp(big.NewInt(0)) > 0),
        BiddingEnded: (currentBlock >= endBlock),
        HasRemainingRpl: (remainingRpl.Cmp(big.NewInt(0)) > 0),
        RplRecovered: rplRecovered,
    }, nil

}

