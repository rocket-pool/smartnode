package api

import (
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/auction"
)


type AuctionStatusResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
}


type AuctionLotsResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    Lots []auction.LotDetails   `json:"lots"`
}


type CanCreateLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    CanCreate bool              `json:"canCreate"`
    InsufficientBalance bool    `json:"insufficientBalance"`
    CreateLotDisabled bool      `json:"createLotDisabled"`
}
type CreateLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    LotId uint64                `json:"lotId"`
    TxHash common.Hash          `json:"txHash"`
}


type CanBidOnLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    CanBid bool                 `json:"canBid"`
    DoesNotExist bool           `json:"doesNotExist"`
    BiddingEnded bool           `json:"biddingEnded"`
    RPLExhausted bool           `json:"rplExhausted"`
    BidOnLotDisabled bool       `json:"bidOnLotDisabled"`
}
type BidOnLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    TxHash common.Hash          `json:"txHash"`
}


type CanClaimFromLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    CanClaim bool               `json:"canClaim"`
    DoesNotExist bool           `json:"doesNotExist"`
    NoBidFromAddress bool       `json:"noBidFromAddress"`
    NotCleared bool             `json:"notCleared"`
}
type ClaimFromLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    TxHash common.Hash          `json:"txHash"`
}


type CanRecoverRPLFromLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    CanRecover bool             `json:"canRecover"`
    DoesNotExist bool           `json:"doesNotExist"`
    NotCleared bool             `json:"notCleared"`
    NoUnclaimedRPL bool         `json:"noUnclaimedRpl"`
    RPLAlreadyRecovered bool    `json:"rplAlreadyRecovered"`
}
type RecoverRPLFromLotResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    TxHash common.Hash          `json:"txHash"`
}

