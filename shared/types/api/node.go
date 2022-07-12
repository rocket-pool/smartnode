package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
)

type NodeStatusResponse struct {
	Status                   string          `json:"status"`
	Error                    string          `json:"error"`
	AccountAddress           common.Address  `json:"accountAddress"`
	WithdrawalAddress        common.Address  `json:"withdrawalAddress"`
	PendingWithdrawalAddress common.Address  `json:"pendingWithdrawalAddress"`
	Registered               bool            `json:"registered"`
	Trusted                  bool            `json:"trusted"`
	TimezoneLocation         string          `json:"timezoneLocation"`
	AccountBalances          tokens.Balances `json:"accountBalances"`
	WithdrawalBalances       tokens.Balances `json:"withdrawalBalances"`
	RplStake                 *big.Int        `json:"rplStake"`
	EffectiveRplStake        *big.Int        `json:"effectiveRplStake"`
	MinimumRplStake          *big.Int        `json:"minimumRplStake"`
	MaximumRplStake          *big.Int        `json:"maximumRplStake"`
	CollateralRatio          float64         `json:"collateralRatio"`
	VotingDelegate           common.Address  `json:"votingDelegate"`
	IsInSmoothingPool        bool            `json:"isInSmoothingPool"`
	MinipoolLimit            uint64          `json:"minipoolLimit"`
	MinipoolCounts           struct {
		Total               int `json:"total"`
		Initialized         int `json:"initialized"`
		Prelaunch           int `json:"prelaunch"`
		Staking             int `json:"staking"`
		Withdrawable        int `json:"withdrawable"`
		Dissolved           int `json:"dissolved"`
		RefundAvailable     int `json:"refundAvailable"`
		WithdrawalAvailable int `json:"withdrawalAvailable"`
		CloseAvailable      int `json:"closeAvailable"`
		Finalised           int `json:"finalised"`
	} `json:"minipoolCounts"`
	IsMergeUpdateDeployed       bool                      `json:"isMergeUpdateDeployed"`
	IsFeeDistributorInitialized bool                      `json:"isFeeDistributorInitialized"`
	FeeDistributorAddress       common.Address            `json:"feeDistributorAddress"`
	FeeDistributorBalance       *big.Int                  `json:"feeDistributorBalance"`
	PenalizedMinipools          map[common.Address]uint64 `json:"penalizedMinipools"`
}

type CanRegisterNodeResponse struct {
	Status               string             `json:"status"`
	Error                string             `json:"error"`
	CanRegister          bool               `json:"canRegister"`
	AlreadyRegistered    bool               `json:"alreadyRegistered"`
	RegistrationDisabled bool               `json:"registrationDisabled"`
	GasInfo              rocketpool.GasInfo `json:"gasInfo"`
}
type RegisterNodeResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanSetNodeWithdrawalAddressResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	CanSet  bool               ` json:"canSet"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetNodeWithdrawalAddressResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanConfirmNodeWithdrawalAddressResponse struct {
	Status     string             `json:"status"`
	Error      string             `json:"error"`
	CanConfirm bool               `json:"canConfirm"`
	GasInfo    rocketpool.GasInfo `json:"gasInfo"`
}
type ConfirmNodeWithdrawalAddressResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetNodeWithdrawalAddressResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type GetNodePendingWithdrawalAddressResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type CanSetNodeTimezoneResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	CanSet  bool               `json:"canSet"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetNodeTimezoneResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeSwapRplResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanSwap             bool               `json:"canSwap"`
	InsufficientBalance bool               `json:"insufficientBalance"`
	GasInfo             rocketpool.GasInfo `json:"GasInfo"`
}
type NodeSwapRplApproveGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeSwapRplApproveResponse struct {
	Status        string      `json:"status"`
	Error         string      `json:"error"`
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type NodeSwapRplSwapResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	SwapTxHash common.Hash `json:"swapTxHash"`
}
type NodeSwapRplAllowanceResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Allowance *big.Int `json:"allowance"`
}

type CanNodeStakeRplResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanStake            bool               `json:"canStake"`
	InsufficientBalance bool               `json:"insufficientBalance"`
	InConsensus         bool               `json:"inConsensus"`
	GasInfo             rocketpool.GasInfo `json:"gasInfo"`
}
type NodeStakeRplApproveGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeStakeRplApproveResponse struct {
	Status        string      `json:"status"`
	Error         string      `json:"error"`
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type NodeStakeRplStakeResponse struct {
	Status      string      `json:"status"`
	Error       string      `json:"error"`
	StakeTxHash common.Hash `json:"stakeTxHash"`
}
type NodeStakeRplAllowanceResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Allowance *big.Int `json:"allowance"`
}

type CanNodeWithdrawRplResponse struct {
	Status                       string             `json:"status"`
	Error                        string             `json:"error"`
	CanWithdraw                  bool               `json:"canWithdraw"`
	InsufficientBalance          bool               `json:"insufficientBalance"`
	MinipoolsUndercollateralized bool               `json:"minipoolsUndercollateralized"`
	WithdrawalDelayActive        bool               `json:"withdrawalDelayActive"`
	InConsensus                  bool               `json:"inConsensus"`
	GasInfo                      rocketpool.GasInfo `json:"gasInfo"`
}
type NodeWithdrawRplResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeDepositResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanDeposit             bool               `json:"canDeposit"`
	InsufficientBalance    bool               `json:"insufficientBalance"`
	InsufficientRplStake   bool               `json:"insufficientRplStake"`
	InvalidAmount          bool               `json:"invalidAmount"`
	UnbondedMinipoolsAtMax bool               `json:"unbondedMinipoolsAtMax"`
	DepositDisabled        bool               `json:"depositDisabled"`
	InConsensus            bool               `json:"inConsensus"`
	MinipoolAddress        common.Address     `json:"minipoolAddress"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type NodeDepositResponse struct {
	Status          string                  `json:"status"`
	Error           string                  `json:"error"`
	TxHash          common.Hash             `json:"txHash"`
	MinipoolAddress common.Address          `json:"minipoolAddress"`
	ValidatorPubkey rptypes.ValidatorPubkey `json:"validatorPubkey"`
	ScrubPeriod     time.Duration           `json:"scrubPeriod"`
}

type CanNodeSendResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanSend             bool               `json:"canSend"`
	InsufficientBalance bool               `json:"insufficientBalance"`
	GasInfo             rocketpool.GasInfo `json:"gasInfo"`
}
type NodeSendResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeBurnResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanBurn                bool               `json:"canBurn"`
	InsufficientBalance    bool               `json:"insufficientBalance"`
	InsufficientCollateral bool               `json:"insufficientCollateral"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type NodeBurnResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeSyncProgressResponse struct {
	Status       string              `json:"status"`
	Error        string              `json:"error"`
	EcStatus     ClientManagerStatus `json:"ecStatus"`
	Eth2Progress float64             `json:"eth2Progress"`
	Eth2Synced   bool                `json:"eth2Synced"`
}

type CanNodeClaimRplResponse struct {
	Status    string             `json:"status"`
	Error     string             `json:"error"`
	RplAmount *big.Int           `json:"rplAmount"`
	GasInfo   rocketpool.GasInfo `json:"gasInfo"`
}
type NodeClaimRplResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeRewardsResponse struct {
	Status                      string        `json:"status"`
	Error                       string        `json:"error"`
	NodeRegistrationTime        time.Time     `json:"nodeRegistrationTime"`
	RewardsInterval             time.Duration `json:"rewardsInterval"`
	LastCheckpoint              time.Time     `json:"lastCheckpoint"`
	Trusted                     bool          `json:"trusted"`
	Registered                  bool          `json:"registered"`
	EffectiveRplStake           float64       `json:"effectiveRplStake"`
	TotalRplStake               float64       `json:"totalRplStake"`
	TrustedRplBond              float64       `json:"trustedRplBond"`
	EstimatedRewards            float64       `json:"estimatedRewards"`
	CumulativeRplRewards        float64       `json:"cumulativeRplRewards"`
	CumulativeEthRewards        float64       `json:"cumulativeEthRewards"`
	EstimatedTrustedRplRewards  float64       `json:"estimatedTrustedRplRewards"`
	CumulativeTrustedRplRewards float64       `json:"cumulativeTrustedRplRewards"`
	UnclaimedRplRewards         float64       `json:"unclaimedRplRewards"`
	UnclaimedEthRewards         float64       `json:"unclaimedEthRewards"`
	UnclaimedTrustedRplRewards  float64       `json:"unclaimedTrustedRplRewards"`
	BeaconRewards               float64       `json:"beaconRewards"`
	IsMergeUpdateDeployed       bool          `json:"isMergeUpdateDeployed"`
	TxHash                      common.Hash   `json:"txHash"`
}

type DepositContractInfoResponse struct {
	Status                string         `json:"status"`
	Error                 string         `json:"error"`
	RPDepositContract     common.Address `json:"rpDepositContract"`
	RPNetwork             uint64         `json:"rpNetwork"`
	BeaconDepositContract common.Address `json:"beaconDepositContract"`
	BeaconNetwork         uint64         `json:"beaconNetwork"`
	SufficientSync        bool           `json:"sufficientSync"`
}

type NodeSignResponse struct {
	Status     string `json:"status"`
	Error      string `json:"error"`
	SignedData string `json:"signedData"`
}

type EstimateSetSnapshotDelegateGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}

type SetSnapshotDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type EstimateClearSnapshotDelegateGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}

type ClearSnapshotDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeIsFeeDistributorInitializedResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	IsInitialized bool   `json:"isInitialized"`
}
type NodeInitializeFeeDistributorGasResponse struct {
	Status      string             `json:"status"`
	Error       string             `json:"error"`
	Distributor common.Address     `json:"distributor"`
	GasInfo     rocketpool.GasInfo `json:"gasInfo"`
}
type NodeInitializeFeeDistributorResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
type NodeCanDistributeResponse struct {
	Status         string             `json:"status"`
	Error          string             `json:"error"`
	Balance        *big.Int           `json:"balance"`
	AverageNodeFee float64            `json:"averageNodeFee"`
	GasInfo        rocketpool.GasInfo `json:"gasInfo"`
}
type NodeDistributeResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeGetRewardsInfoResponse struct {
	Status             string                 `json:"status"`
	Error              string                 `json:"error"`
	ClaimedIntervals   []uint64               `json:"claimedIntervals"`
	UnclaimedIntervals []rewards.IntervalInfo `json:"unclaimedIntervals"`
	InvalidIntervals   []rewards.IntervalInfo `json:"invalidIntervals"`
	RplStake           *big.Int               `json:"rplStake"`
	RplPrice           *big.Int               `json:"rplPrice"`
	ActiveMinipools    int                    `json:"activeMinipools"`
}

type CanNodeClaimRewardsResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeClaimRewardsResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeClaimAndStakeRewardsResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeClaimAndStakeRewardsResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetSmoothingPoolRegistrationStatusResponse struct {
	Status                  string        `json:"status"`
	Error                   string        `json:"error"`
	NodeRegistered          bool          `json:"nodeRegistered"`
	TimeLeftUntilChangeable time.Duration `json:"timeLeftUntilChangeable"`
}
type CanSetSmoothingPoolRegistrationStatusResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetSmoothingPoolRegistrationStatusResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
