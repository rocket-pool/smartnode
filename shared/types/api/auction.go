package api

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/core"
)

type AuctionStatusData struct {
	TotalRPLBalance     *big.Int `json:"totalRPLBalance"`
	AllottedRPLBalance  *big.Int `json:"allottedRPLBalance"`
	RemainingRPLBalance *big.Int `json:"remainingRPLBalance"`
	CanCreateLot        bool     `json:"canCreateLot"`
	LotCounts           struct {
		ClaimAvailable       int `json:"claimAvailable"`
		BiddingAvailable     int `json:"biddingAvailable"`
		RPLRecoveryAvailable int `json:"rplRecoveryAvailable"`
	} `json:"lotCounts"`
}

type AuctionLotDetails struct {
	Details              auction.AuctionLotDetails `json:"details"`
	ClaimAvailable       bool                      `json:"claimAvailable"`
	BiddingAvailable     bool                      `json:"biddingAvailable"`
	RPLRecoveryAvailable bool                      `json:"rplRecoveryAvailable"`
	NodeBidAmount        *big.Int                  `json:"nodeBidAmount"`
}
type AuctionLotsData struct {
	Lots []AuctionLotDetails `json:"lots"`
}

type CreateLotData struct {
	CanCreate           bool                  `json:"canCreate"`
	InsufficientBalance bool                  `json:"insufficientBalance"`
	CreateLotDisabled   bool                  `json:"createLotDisabled"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}

type BidOnLotData struct {
	CanBid           bool                  `json:"canBid"`
	DoesNotExist     bool                  `json:"doesNotExist"`
	BiddingEnded     bool                  `json:"biddingEnded"`
	RPLExhausted     bool                  `json:"rplExhausted"`
	BidOnLotDisabled bool                  `json:"bidOnLotDisabled"`
	TxInfo           *core.TransactionInfo `json:"txInfo"`
}

type ClaimFromLotData struct {
	CanClaim         bool                  `json:"canClaim"`
	DoesNotExist     bool                  `json:"doesNotExist"`
	NoBidFromAddress bool                  `json:"noBidFromAddress"`
	NotCleared       bool                  `json:"notCleared"`
	TxInfo           *core.TransactionInfo `json:"txInfo"`
}

type RecoverRplFromLotData struct {
	CanRecover          bool                  `json:"canRecover"`
	DoesNotExist        bool                  `json:"doesNotExist"`
	BiddingNotEnded     bool                  `json:"biddingNotEnded"`
	NoUnclaimedRPL      bool                  `json:"noUnclaimedRpl"`
	RPLAlreadyRecovered bool                  `json:"rplAlreadyRecovered"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}
