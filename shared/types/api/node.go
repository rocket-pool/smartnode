package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
)


type NodeStatusResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    AccountAddress common.Address       `json:"accountAddress"`
    WithdrawalAddress common.Address    `json:"withdrawalAddress"`
    Registered bool                     `json:"registered"`
    Trusted bool                        `json:"trusted"`
    TimezoneLocation string             `json:"timezoneLocation"`
    AccountBalances tokens.Balances     `json:"accountBalances"`
    WithdrawalBalances tokens.Balances  `json:"withdrawalBalances"`
    RplStake *big.Int                   `json:"rplStake"`
    EffectiveRplStake *big.Int          `json:"effectiveRplStake"`
    MinimumRplStake *big.Int            `json:"minimumRplStake"`
    MinipoolLimit uint64                `json:"minipoolLimit"`
    MinipoolCounts struct {
        Total int                           `json:"total"`
        Initialized int                     `json:"initialized"`
        Prelaunch int                       `json:"prelaunch"`
        Staking int                         `json:"staking"`
        Withdrawable int                    `json:"withdrawable"`
        Dissolved int                       `json:"dissolved"`
        RefundAvailable int                 `json:"refundAvailable"`
        WithdrawalAvailable int             `json:"withdrawalAvailable"`
        CloseAvailable int                  `json:"closeAvailable"`
    }                                   `json:"minipoolCounts"`
}


type NodeLeaderResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Nodes []NodeRank                `json:"nodes"`
}

type NodeRank struct {
    Rank int                        `json:"rank"`
    Address common.Address          `json:"address"`
    Registered bool                 `json:"registered"`
    TimezoneLocation string         `json:"timezoneLocation"`
    Score *big.Int                  `json:"score"`
    Details []MinipoolDetails       `json:"details"`
}


type CanRegisterNodeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanRegister bool                    `json:"canRegister"`
    AlreadyRegistered bool              `json:"alreadyRegistered"`
    RegistrationDisabled bool           `json:"registrationDisabled"`
}
type RegisterNodeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type SetNodeWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type SetNodeTimezoneResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanNodeSwapRplResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanSwap bool                        `json:"canSwap"`
    InsufficientBalance bool            `json:"insufficientBalance"`
}
type NodeSwapRplApproveResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    ApproveTxHash common.Hash           `json:"approveTxHash"`
}
type NodeSwapRplSwapResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    SwapTxHash common.Hash              `json:"swapTxHash"`
}


type CanNodeStakeRplResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanStake bool                       `json:"canStake"`
    InsufficientBalance bool            `json:"insufficientBalance"`
}
type NodeStakeRplApproveResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    ApproveTxHash common.Hash           `json:"approveTxHash"`
}
type NodeStakeRplStakeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    StakeTxHash common.Hash             `json:"stakeTxHash"`
}


type CanNodeWithdrawRplResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanWithdraw bool                    `json:"canWithdraw"`
    InsufficientBalance bool            `json:"insufficientBalance"`
    MinipoolsUndercollateralized bool   `json:"minipoolsUndercollateralized"`
    WithdrawalDelayActive bool          `json:"withdrawalDelayActive"`
}
type NodeWithdrawRplResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanNodeDepositResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanDeposit bool                     `json:"canDeposit"`
    InsufficientBalance bool            `json:"insufficientBalance"`
    InsufficientRplStake bool           `json:"insufficientRplStake"`
    InvalidAmount bool                  `json:"invalidAmount"`
    UnbondedMinipoolsAtMax bool         `json:"unbondedMinipoolsAtMax"`
    DepositDisabled bool                `json:"depositDisabled"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
}
type NodeDepositResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}
type NodeDepositMinipoolResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    MinipoolAddress common.Address      `json:"minipoolAddress"`
}


type CanNodeSendResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanSend bool                        `json:"canSend"`
    InsufficientBalance bool            `json:"insufficientBalance"`
}
type NodeSendResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanNodeBurnResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanBurn bool                        `json:"canBurn"`
    InsufficientBalance bool            `json:"insufficientBalance"`
    InsufficientCollateral bool         `json:"insufficientCollateral"`
}
type NodeBurnResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}

