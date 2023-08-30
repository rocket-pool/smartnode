package api

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/core"
)

type AuctionStatusResponse struct {
	Status              string   `json:"status"`
	Error               string   `json:"error"`
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

type AuctionLotsResponse struct {
	Status string       `json:"status"`
	Error  string       `json:"error"`
	Lots   []LotDetails `json:"lots"`
}
type LotDetails struct {
	Details              auction.AuctionLotDetails `json:"details"`
	ClaimAvailable       bool                      `json:"claimAvailable"`
	BiddingAvailable     bool                      `json:"biddingAvailable"`
	RPLRecoveryAvailable bool                      `json:"rplRecoveryAvailable"`
	NodeBidAmount        *big.Int                  `json:"nodeBidAmount"`
}

type CreateLotResponse struct {
	Status              string                `json:"status"`
	Error               string                `json:"error"`
	CanCreate           bool                  `json:"canCreate"`
	InsufficientBalance bool                  `json:"insufficientBalance"`
	CreateLotDisabled   bool                  `json:"createLotDisabled"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}

type BidOnLotResponse struct {
	Status           string                `json:"status"`
	Error            string                `json:"error"`
	CanBid           bool                  `json:"canBid"`
	DoesNotExist     bool                  `json:"doesNotExist"`
	BiddingEnded     bool                  `json:"biddingEnded"`
	RPLExhausted     bool                  `json:"rplExhausted"`
	BidOnLotDisabled bool                  `json:"bidOnLotDisabled"`
	TxInfo           *core.TransactionInfo `json:"txInfo"`
}

type ClaimFromLotResponse struct {
	Status           string                `json:"status"`
	Error            string                `json:"error"`
	CanClaim         bool                  `json:"canClaim"`
	DoesNotExist     bool                  `json:"doesNotExist"`
	NoBidFromAddress bool                  `json:"noBidFromAddress"`
	NotCleared       bool                  `json:"notCleared"`
	TxInfo           *core.TransactionInfo `json:"txInfo"`
}

type RecoverRPLFromLotResponse struct {
	Status              string                `json:"status"`
	Error               string                `json:"error"`
	CanRecover          bool                  `json:"canRecover"`
	DoesNotExist        bool                  `json:"doesNotExist"`
	BiddingNotEnded     bool                  `json:"biddingNotEnded"`
	NoUnclaimedRPL      bool                  `json:"noUnclaimedRpl"`
	RPLAlreadyRecovered bool                  `json:"rplAlreadyRecovered"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}
