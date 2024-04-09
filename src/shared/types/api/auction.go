package api

import (
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
)

type AuctionStatusData struct {
	TotalRplBalance     *big.Int `json:"totalRPLBalance"`
	AllottedRplBalance  *big.Int `json:"allottedRPLBalance"`
	RemainingRplBalance *big.Int `json:"remainingRPLBalance"`
	CanCreateLot        bool     `json:"canCreateLot"`
	LotCounts           struct {
		ClaimAvailable       int `json:"claimAvailable"`
		BiddingAvailable     int `json:"biddingAvailable"`
		RplRecoveryAvailable int `json:"rplRecoveryAvailable"`
	} `json:"lotCounts"`
}

type AuctionLotDetails struct {
	Index                uint64   `json:"index"`
	Exists               bool     `json:"exists"`
	StartBlock           uint64   `json:"startBlock"`
	EndBlock             uint64   `json:"endBlock"`
	StartPrice           *big.Int `json:"startPrice"`
	ReservePrice         *big.Int `json:"reservePrice"`
	PriceAtCurrentBlock  *big.Int `json:"priceAtCurrentBlock"`
	PriceByTotalBids     *big.Int `json:"priceByTotalBids"`
	CurrentPrice         *big.Int `json:"currentPrice"`
	TotalRplAmount       *big.Int `json:"totalRplAmount"`
	ClaimedRplAmount     *big.Int `json:"claimedRplAmount"`
	RemainingRplAmount   *big.Int `json:"remainingRplAmount"`
	TotalBidAmount       *big.Int `json:"totalBidAmount"`
	IsCleared            bool     `json:"isCleared"`
	RplRecovered         bool     `json:"rplRecovered"`
	ClaimAvailable       bool     `json:"claimAvailable"`
	BiddingAvailable     bool     `json:"biddingAvailable"`
	RplRecoveryAvailable bool     `json:"rplRecoveryAvailable"`
	NodeBidAmount        *big.Int `json:"nodeBidAmount"`
}
type AuctionLotsData struct {
	Lots []AuctionLotDetails `json:"lots"`
}

type AuctionCreateLotData struct {
	CanCreate           bool                 `json:"canCreate"`
	InsufficientBalance bool                 `json:"insufficientBalance"`
	CreateLotDisabled   bool                 `json:"createLotDisabled"`
	TxInfo              *eth.TransactionInfo `json:"txInfo"`
}

type AuctionBidOnLotData struct {
	CanBid           bool                 `json:"canBid"`
	DoesNotExist     bool                 `json:"doesNotExist"`
	BiddingEnded     bool                 `json:"biddingEnded"`
	RplExhausted     bool                 `json:"rplExhausted"`
	BidOnLotDisabled bool                 `json:"bidOnLotDisabled"`
	TxInfo           *eth.TransactionInfo `json:"txInfo"`
}

type AuctionClaimFromLotData struct {
	CanClaim         bool                 `json:"canClaim"`
	DoesNotExist     bool                 `json:"doesNotExist"`
	NoBidFromAddress bool                 `json:"noBidFromAddress"`
	NotCleared       bool                 `json:"notCleared"`
	TxInfo           *eth.TransactionInfo `json:"txInfo"`
}

type AuctionRecoverRplFromLotData struct {
	CanRecover          bool                 `json:"canRecover"`
	DoesNotExist        bool                 `json:"doesNotExist"`
	BiddingNotEnded     bool                 `json:"biddingNotEnded"`
	NoUnclaimedRpl      bool                 `json:"noUnclaimedRpl"`
	RplAlreadyRecovered bool                 `json:"rplAlreadyRecovered"`
	TxInfo              *eth.TransactionInfo `json:"txInfo"`
}
