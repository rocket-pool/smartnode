package api

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
)


// Node status response type
type NodeStatusResponse struct {
    AccountAddress common.Address       `json:"accountAddress"`
    AccountBalanceEtherWei *big.Int     `json:"accountBalanceEtherWei"`
    AccountBalanceRethWei *big.Int      `json:"accountBalanceRethWei"`
    AccountBalanceRplWei *big.Int       `json:"accountBalanceRplWei"`
    ContractAddress common.Address      `json:"contractAddress"`
    ContractBalanceEtherWei *big.Int    `json:"contractBalanceEtherWei"`
    ContractBalanceRplWei *big.Int      `json:"contractBalanceRplWei"`
    Registered bool                     `json:"registered"`
    Active bool                         `json:"active"`
    Trusted bool                        `json:"trusted"`
    Timezone string                     `json:"timezone"`
}

