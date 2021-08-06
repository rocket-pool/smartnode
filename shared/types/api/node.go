package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
)


type NodeStatusResponse struct {
    Status string                               `json:"status"`
    Error string                                `json:"error"`
    AccountAddress common.Address               `json:"accountAddress"`
    WithdrawalAddress common.Address            `json:"withdrawalAddress"`
    PendingWithdrawalAddress common.Address     `json:"pendingWithdrawalAddress"`
    Registered bool                             `json:"registered"`
    Trusted bool                                `json:"trusted"`
    TimezoneLocation string                     `json:"timezoneLocation"`
    AccountBalances tokens.Balances             `json:"accountBalances"`
    WithdrawalBalances tokens.Balances          `json:"withdrawalBalances"`
    RplStake *big.Int                           `json:"rplStake"`
    EffectiveRplStake *big.Int                  `json:"effectiveRplStake"`
    MinimumRplStake *big.Int                    `json:"minimumRplStake"`
    MaximumRplStake *big.Int                    `json:"maximumRplStake"`
    CollateralRatio float64                     `json:"collateralRatio"`
    MinipoolLimit uint64                        `json:"minipoolLimit"`
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
        Finalised int                       `json:"finalised"`
    }                                           `json:"minipoolCounts"`
}


type CanRegisterNodeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanRegister bool                    `json:"canRegister"`
    AlreadyRegistered bool              `json:"alreadyRegistered"`
    RegistrationDisabled bool           `json:"registrationDisabled"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
}
type RegisterNodeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanSetNodeWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanSet bool                        ` json:"canSet"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
}
type SetNodeWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanConfirmNodeWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanConfirm bool                     `json:"canConfirm"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
}
type ConfirmNodeWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type GetNodeWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    Address common.Address              `json:"address"`
}


type GetNodePendingWithdrawalAddressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    Address common.Address              `json:"address"`
}


type CanSetNodeTimezoneResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanSet bool                         `json:"canSet"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
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
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
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
    InConsensus bool                    `json:"inConsensus"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
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
    InConsensus bool                    `json:"inConsensus"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
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
    InConsensus bool                    `json:"inConsensus"`
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
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
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
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
}
type NodeBurnResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}

type NodeSyncProgressResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    Eth1Progress float64                `json:"eth1Progress"`
    Eth2Progress float64                `json:"eth2Progress"`
    Eth1Synced bool                     `json:"eth1Synced"`
    Eth2Synced bool                     `json:"eth2Synced"`
}


type CanNodeClaimRplResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    RplAmount *big.Int                  `json:"rplAmount"`
    GasInfo rocketpool.GasInfo          `json:"gasInfo"`
}
type NodeClaimRplResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}

type NodeRewardsResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    RewardsInterval time.Duration       `json:"rewardsInterval"`
    LastCheckpoint time.Time            `json:"lastCheckpoint"`
    Trusted bool                        `json:"trusted"`
    EstimatedRewards float64            `json:"estimatedRewards"`
    CumulativeRewards float64           `json:"cumulativeRewards"`
    EstimatedTrustedRewards float64     `json:"estimatedTrustedRewards"`
    CumulativeTrustedRewards float64    `json:"cumulativeTrustedRewards"`
    TxHash common.Hash                  `json:"txHash"`
}