package auction

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Settings
const LotDetailsBatchSize = 10

// Lot details
type LotDetails struct {
	Index               uint64   `json:"index"`
	Exists              bool     `json:"exists"`
	StartBlock          uint64   `json:"startBlock"`
	EndBlock            uint64   `json:"endBlock"`
	StartPrice          *big.Int `json:"startPrice"`
	ReservePrice        *big.Int `json:"reservePrice"`
	PriceAtCurrentBlock *big.Int `json:"priceAtCurrentBlock"`
	PriceByTotalBids    *big.Int `json:"priceByTotalBids"`
	CurrentPrice        *big.Int `json:"currentPrice"`
	TotalRPLAmount      *big.Int `json:"totalRplAmount"`
	ClaimedRPLAmount    *big.Int `json:"claimedRplAmount"`
	RemainingRPLAmount  *big.Int `json:"remainingRplAmount"`
	TotalBidAmount      *big.Int `json:"totalBidAmount"`
	AddressBidAmount    *big.Int `json:"addressBidAmount"`
	Cleared             bool     `json:"cleared"`
	RPLRecovered        bool     `json:"rplRecovered"`
}

// Get all lot details
func GetLots(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]LotDetails, error) {

	// Get lot count
	lotCount, err := GetLotCount(rp, opts)
	if err != nil {
		return []LotDetails{}, err
	}

	// Load lot details in batches
	details := make([]LotDetails, lotCount)
	for bsi := uint64(0); bsi < lotCount; bsi += LotDetailsBatchSize {

		// Get batch start & end index
		lsi := bsi
		lei := bsi + LotDetailsBatchSize
		if lei > lotCount {
			lei = lotCount
		}

		// Load details
		var wg errgroup.Group
		for li := lsi; li < lei; li++ {
			li := li
			wg.Go(func() error {
				lotDetails, err := GetLotDetails(rp, li, opts)
				if err == nil {
					details[li] = lotDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []LotDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get all lot details with bids from an address
func GetLotsWithBids(rp *rocketpool.RocketPool, bidder common.Address, opts *bind.CallOpts) ([]LotDetails, error) {

	// Get lot count
	lotCount, err := GetLotCount(rp, opts)
	if err != nil {
		return []LotDetails{}, err
	}

	// Load lot details in batches
	details := make([]LotDetails, lotCount)
	for bsi := uint64(0); bsi < lotCount; bsi += LotDetailsBatchSize {

		// Get batch start & end index
		lsi := bsi
		lei := bsi + LotDetailsBatchSize
		if lei > lotCount {
			lei = lotCount
		}

		// Load details
		var wg errgroup.Group
		for li := lsi; li < lei; li++ {
			li := li
			wg.Go(func() error {
				lotDetails, err := GetLotDetailsWithBids(rp, li, bidder, opts)
				if err == nil {
					details[li] = lotDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []LotDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get a lot's details
func GetLotDetails(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (LotDetails, error) {

	// Data
	var wg errgroup.Group
	var exists bool
	var startBlock uint64
	var endBlock uint64
	var startPrice *big.Int
	var reservePrice *big.Int
	var priceAtCurrentBlock *big.Int
	var priceByTotalBids *big.Int
	var currentPrice *big.Int
	var totalRplAmount *big.Int
	var claimedRplAmount *big.Int
	var remainingRplAmount *big.Int
	var totalBidAmount *big.Int
	var cleared bool
	var rplRecovered bool

	// Load data
	wg.Go(func() error {
		var err error
		exists, err = GetLotExists(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		startBlock, err = GetLotStartBlock(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		endBlock, err = GetLotEndBlock(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		startPrice, err = GetLotStartPrice(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		reservePrice, err = GetLotReservePrice(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		priceAtCurrentBlock, err = GetLotPriceAtCurrentBlock(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		priceByTotalBids, err = GetLotPriceByTotalBids(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		currentPrice, err = GetLotCurrentPrice(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		totalRplAmount, err = GetLotTotalRPLAmount(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		claimedRplAmount, err = GetLotClaimedRPLAmount(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		remainingRplAmount, err = GetLotRemainingRPLAmount(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		totalBidAmount, err = GetLotTotalBidAmount(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		cleared, err = GetLotIsCleared(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		rplRecovered, err = GetLotRPLRecovered(rp, lotIndex, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return LotDetails{}, err
	}

	// Return
	return LotDetails{
		Index:               lotIndex,
		Exists:              exists,
		StartBlock:          startBlock,
		EndBlock:            endBlock,
		StartPrice:          startPrice,
		ReservePrice:        reservePrice,
		PriceAtCurrentBlock: priceAtCurrentBlock,
		PriceByTotalBids:    priceByTotalBids,
		CurrentPrice:        currentPrice,
		TotalRPLAmount:      totalRplAmount,
		ClaimedRPLAmount:    claimedRplAmount,
		RemainingRPLAmount:  remainingRplAmount,
		TotalBidAmount:      totalBidAmount,
		Cleared:             cleared,
		RPLRecovered:        rplRecovered,
	}, nil

}

// Get a lot's details with address bid amounts
func GetLotDetailsWithBids(rp *rocketpool.RocketPool, lotIndex uint64, bidder common.Address, opts *bind.CallOpts) (LotDetails, error) {

	// Data
	var wg errgroup.Group
	var details LotDetails
	var addressBidAmount *big.Int

	// Load data
	wg.Go(func() error {
		var err error
		details, err = GetLotDetails(rp, lotIndex, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		addressBidAmount, err = GetLotAddressBidAmount(rp, lotIndex, bidder, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return LotDetails{}, err
	}

	// Return
	details.AddressBidAmount = addressBidAmount
	return details, nil

}

// Get the total RPL balance of the auction contract
func GetTotalRPLBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRplBalance := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, totalRplBalance, "getTotalRPLBalance"); err != nil {
		return nil, fmt.Errorf("Could not get auction contract total RPL balance: %w", err)
	}
	return *totalRplBalance, nil
}

// Get the allotted RPL balance of the auction contract
func GetAllottedRPLBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	allottedRplBalance := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, allottedRplBalance, "getAllottedRPLBalance"); err != nil {
		return nil, fmt.Errorf("Could not get auction contract allotted RPL balance: %w", err)
	}
	return *allottedRplBalance, nil
}

// Get the remaining RPL balance of the auction contract
func GetRemainingRPLBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	remainingRplBalance := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, remainingRplBalance, "getRemainingRPLBalance"); err != nil {
		return nil, fmt.Errorf("Could not get auction contract remaining RPL balance: %w", err)
	}
	return *remainingRplBalance, nil
}

// Get the number of lots for auction
func GetLotCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return 0, err
	}
	lotCount := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotCount, "getLotCount"); err != nil {
		return 0, fmt.Errorf("Could not get lot count: %w", err)
	}
	return (*lotCount).Uint64(), nil
}

// Lot details
func GetLotExists(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (bool, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return false, err
	}
	lotExists := new(bool)
	if err := rocketAuctionManager.Call(opts, lotExists, "getLotExists", big.NewInt(int64(lotIndex))); err != nil {
		return false, fmt.Errorf("Could not get lot %d exists status: %w", lotIndex, err)
	}
	return *lotExists, nil
}
func GetLotStartBlock(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (uint64, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return 0, err
	}
	lotStartBlock := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotStartBlock, "getLotStartBlock", big.NewInt(int64(lotIndex))); err != nil {
		return 0, fmt.Errorf("Could not get lot %d start block: %w", lotIndex, err)
	}
	return (*lotStartBlock).Uint64(), nil
}
func GetLotEndBlock(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (uint64, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return 0, err
	}
	lotEndBlock := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotEndBlock, "getLotEndBlock", big.NewInt(int64(lotIndex))); err != nil {
		return 0, fmt.Errorf("Could not get lot %d end block: %w", lotIndex, err)
	}
	return (*lotEndBlock).Uint64(), nil
}
func GetLotStartPrice(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotStartPrice := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotStartPrice, "getLotStartPrice", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d start price: %w", lotIndex, err)
	}
	return *lotStartPrice, nil
}
func GetLotReservePrice(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotReservePrice := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotReservePrice, "getLotReservePrice", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d reserve price: %w", lotIndex, err)
	}
	return *lotReservePrice, nil
}
func GetLotTotalRPLAmount(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotTotalRplAmount := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotTotalRplAmount, "getLotTotalRPLAmount", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d total RPL amount: %w", lotIndex, err)
	}
	return *lotTotalRplAmount, nil
}
func GetLotTotalBidAmount(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotTotalBidAmount := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotTotalBidAmount, "getLotTotalBidAmount", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d total ETH bid amount: %w", lotIndex, err)
	}
	return *lotTotalBidAmount, nil
}
func GetLotRPLRecovered(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (bool, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return false, err
	}
	lotRplRecovered := new(bool)
	if err := rocketAuctionManager.Call(opts, lotRplRecovered, "getLotRPLRecovered", big.NewInt(int64(lotIndex))); err != nil {
		return false, fmt.Errorf("Could not get lot %d RPL recovered status: %w", lotIndex, err)
	}
	return *lotRplRecovered, nil
}
func GetLotPriceAtCurrentBlock(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotPriceAtCurrentBlock := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotPriceAtCurrentBlock, "getLotPriceAtCurrentBlock", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d price by current block: %w", lotIndex, err)
	}
	return *lotPriceAtCurrentBlock, nil
}
func GetLotPriceByTotalBids(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotPriceByTotalBids := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotPriceByTotalBids, "getLotPriceByTotalBids", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d price by total bids: %w", lotIndex, err)
	}
	return *lotPriceByTotalBids, nil
}
func GetLotCurrentPrice(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotCurrentPrice := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotCurrentPrice, "getLotCurrentPrice", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d current price: %w", lotIndex, err)
	}
	return *lotCurrentPrice, nil
}
func GetLotClaimedRPLAmount(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotClaimedRplAmount := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotClaimedRplAmount, "getLotClaimedRPLAmount", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d claimed RPL amount: %w", lotIndex, err)
	}
	return *lotClaimedRplAmount, nil
}
func GetLotRemainingRPLAmount(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotRemainingRplAmount := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotRemainingRplAmount, "getLotRemainingRPLAmount", big.NewInt(int64(lotIndex))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d remaining RPL amount: %w", lotIndex, err)
	}
	return *lotRemainingRplAmount, nil
}
func GetLotIsCleared(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.CallOpts) (bool, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return false, err
	}
	lotIsCleared := new(bool)
	if err := rocketAuctionManager.Call(opts, lotIsCleared, "getLotIsCleared", big.NewInt(int64(lotIndex))); err != nil {
		return false, fmt.Errorf("Could not get lot %d cleared status: %w", lotIndex, err)
	}
	return *lotIsCleared, nil
}

// Get the price of a lot at a specific block
func GetLotPriceAtBlock(rp *rocketpool.RocketPool, lotIndex, blockNumber uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lotPriceAtBlock := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lotPriceAtBlock, "getLotPriceAtBlock", big.NewInt(int64(lotIndex)), big.NewInt(int64(blockNumber))); err != nil {
		return nil, fmt.Errorf("Could not get lot %d price at block: %w", lotIndex, err)
	}
	return *lotPriceAtBlock, nil
}

// Get the ETH amount bid on a lot by an address
func GetLotAddressBidAmount(rp *rocketpool.RocketPool, lotIndex uint64, bidder common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, opts)
	if err != nil {
		return nil, err
	}
	lot := new(*big.Int)
	if err := rocketAuctionManager.Call(opts, lot, "getLotAddressBidAmount", big.NewInt(int64(lotIndex)), bidder); err != nil {
		return nil, fmt.Errorf("Could not get lot %d address ETH bid amount: %w", lotIndex, err)
	}
	return *lot, nil
}

// Estimate the gas of CreateLot
func EstimateCreateLotGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketAuctionManager.GetTransactionGasInfo(opts, "createLot")
}

// Create a new lot
func CreateLot(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	lotCount, err := GetLotCount(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	tx, err := rocketAuctionManager.Transact(opts, "createLot")
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("Could not create lot: %w", err)
	}
	return lotCount, tx.Hash(), nil
}

// Estimate the gas of PlaceBid
func EstimatePlaceBidGas(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketAuctionManager.GetTransactionGasInfo(opts, "placeBid", big.NewInt(int64(lotIndex)))
}

// Place a bid on a lot
func PlaceBid(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketAuctionManager.Transact(opts, "placeBid", big.NewInt(int64(lotIndex)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not place bid on lot %d: %w", lotIndex, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ClaimBid
func EstimateClaimBidGas(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketAuctionManager.GetTransactionGasInfo(opts, "claimBid", big.NewInt(int64(lotIndex)))
}

// Claim RPL from a lot that was bid on
func ClaimBid(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketAuctionManager.Transact(opts, "claimBid", big.NewInt(int64(lotIndex)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not claim bid from lot %d: %w", lotIndex, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of RecoverUnclaimedRPL
func EstimateRecoverUnclaimedRPLGas(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketAuctionManager.GetTransactionGasInfo(opts, "recoverUnclaimedRPL", big.NewInt(int64(lotIndex)))
}

// Recover unclaimed RPL from a lot
func RecoverUnclaimedRPL(rp *rocketpool.RocketPool, lotIndex uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketAuctionManager, err := getRocketAuctionManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketAuctionManager.Transact(opts, "recoverUnclaimedRPL", big.NewInt(int64(lotIndex)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not recover unclaimed RPL from lot %d: %w", lotIndex, err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketAuctionManagerLock sync.Mutex

func getRocketAuctionManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketAuctionManagerLock.Lock()
	defer rocketAuctionManagerLock.Unlock()
	return rp.GetContract("rocketAuctionManager", opts)
}
