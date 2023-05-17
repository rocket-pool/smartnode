package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

type MinipoolStatusResponse struct {
	Status         string            `json:"status"`
	Error          string            `json:"error"`
	Minipools      []MinipoolDetails `json:"minipools"`
	LatestDelegate common.Address    `json:"latestDelegate"`
}
type MinipoolDetails struct {
	Address               common.Address         `json:"address"`
	ValidatorPubkey       types.ValidatorPubkey  `json:"validatorPubkey"`
	Status                minipool.StatusDetails `json:"status"`
	DepositType           types.MinipoolDeposit  `json:"depositType"`
	Node                  minipool.NodeDetails   `json:"node"`
	User                  minipool.UserDetails   `json:"user"`
	Balances              tokens.Balances        `json:"balances"`
	NodeShareOfETHBalance *big.Int               `json:"nodeShareOfETHBalance"`
	Validator             ValidatorDetails       `json:"validator"`
	CanStake              bool                   `json:"canStake"`
	CanPromote            bool                   `json:"canPromote"`
	Queue                 minipool.QueueDetails  `json:"queue"`
	RefundAvailable       bool                   `json:"refundAvailable"`
	WithdrawalAvailable   bool                   `json:"withdrawalAvailable"`
	CloseAvailable        bool                   `json:"closeAvailable"`
	Finalised             bool                   `json:"finalised"`
	UseLatestDelegate     bool                   `json:"useLatestDelegate"`
	Delegate              common.Address         `json:"delegate"`
	PreviousDelegate      common.Address         `json:"previousDelegate"`
	EffectiveDelegate     common.Address         `json:"effectiveDelegate"`
	TimeUntilDissolve     time.Duration          `json:"timeUntilDissolve"`
	Penalties             uint64                 `json:"penalties"`
	ReduceBondTime        time.Time              `json:"reduceBondTime"`
	ReduceBondCancelled   bool                   `json:"reduceBondCancelled"`
}
type ValidatorDetails struct {
	Exists      bool     `json:"exists"`
	Active      bool     `json:"active"`
	Index       string   `json:"index"`
	Balance     *big.Int `json:"balance"`
	NodeBalance *big.Int `json:"nodeBalance"`
}
type MinipoolBalanceDistributionDetails struct {
	Address            common.Address       `json:"address"`
	Balance            *big.Int             `json:"balance"`
	Refund             *big.Int             `json:"refund"`
	NodeShareOfBalance *big.Int             `json:"nodeShareOfBalance"`
	MinipoolVersion    uint8                `json:"minipoolVersion"`
	Status             types.MinipoolStatus `json:"status"`
	IsFinalized        bool                 `json:"isFinalized"`
	CanDistribute      bool                 `json:"canDistribute"`
	GasInfo            rocketpool.GasInfo   `json:"gasInfo"`
}

type CanRefundMinipoolResponse struct {
	Status                    string             `json:"status"`
	Error                     string             `json:"error"`
	CanRefund                 bool               `json:"canRefund"`
	InsufficientRefundBalance bool               `json:"insufficientRefundBalance"`
	GasInfo                   rocketpool.GasInfo `json:"gasInfo"`
}
type RefundMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanDissolveMinipoolResponse struct {
	Status        string             `json:"status"`
	Error         string             `json:"error"`
	CanDissolve   bool               `json:"canDissolve"`
	InvalidStatus bool               `json:"invalidStatus"`
	GasInfo       rocketpool.GasInfo `json:"gasInfo"`
}
type DissolveMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanExitMinipoolResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	CanExit       bool   `json:"canExit"`
	InvalidStatus bool   `json:"invalidStatus"`
}
type ExitMinipoolResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type CanChangeWithdrawalCredentialsResponse struct {
	Status    string `json:"status"`
	Error     string `json:"error"`
	CanChange bool   `json:"canChange"`
}
type ChangeWithdrawalCredentialsResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ImportKeyResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type CanProcessWithdrawalResponse struct {
	Status        string             `json:"status"`
	Error         string             `json:"error"`
	CanWithdraw   bool               `json:"canWithdraw"`
	InvalidStatus bool               `json:"invalidStatus"`
	GasInfo       rocketpool.GasInfo `json:"gasInfo"`
}
type ProcessWithdrawalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanProcessWithdrawalAndFinaliseResponse struct {
	Status        string             `json:"status"`
	Error         string             `json:"error"`
	CanWithdraw   bool               `json:"canWithdraw"`
	InvalidStatus bool               `json:"invalidStatus"`
	GasInfo       rocketpool.GasInfo `json:"gasInfo"`
}
type ProcessWithdrawalAndFinaliseResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type MinipoolCloseDetails struct {
	Address            common.Address        `json:"address"`
	IsFinalized        bool                  `json:"isFinalized"`
	MinipoolStatus     types.MinipoolStatus  `json:"minipoolStatus"`
	MinipoolVersion    uint8                 `json:"minipoolVersion"`
	Distributed        bool                  `json:"distributed"`
	CanClose           bool                  `json:"canClose"`
	Balance            *big.Int              `json:"balance"`
	Refund             *big.Int              `json:"refund"`
	UserDepositBalance *big.Int              `json:"userDepositBalance"`
	BeaconState        beacon.ValidatorState `json:"beaconState"`
	NodeShare          *big.Int              `json:"nodeShare"`
	GasInfo            rocketpool.GasInfo    `json:"gasInfo"`
}

type GetMinipoolCloseDetailsForNodeResponse struct {
	Status                      string                 `json:"status"`
	Error                       string                 `json:"error"`
	IsFeeDistributorInitialized bool                   `json:"isFeeDistributorInitialized"`
	Details                     []MinipoolCloseDetails `json:"details"`
}
type CloseMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetDistributeBalanceDetailsResponse struct {
	Status  string                               `json:"status"`
	Error   string                               `json:"error"`
	Details []MinipoolBalanceDistributionDetails `json:"details"`
}
type CanDistributeBalanceResponse struct {
	Status          string               `json:"status"`
	Error           string               `json:"error"`
	MinipoolVersion uint8                `json:"minipoolVersion"`
	MinipoolStatus  types.MinipoolStatus `json:"minipoolStatus"`
	Balance         *big.Int             `json:"balance"`
	CanDistribute   bool                 `json:"canDistribute"`
	GasInfo         rocketpool.GasInfo   `json:"gasInfo"`
}
type EstimateDistributeBalanceGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type DistributeBalanceResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanFinaliseMinipoolResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type FinaliseMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanDelegateUpgradeResponse struct {
	Status                string             `json:"status"`
	Error                 string             `json:"error"`
	LatestDelegateAddress common.Address     `json:"latestDelegateAddress"`
	GasInfo               rocketpool.GasInfo `json:"gasInfo"`
}
type DelegateUpgradeResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanDelegateRollbackResponse struct {
	Status          string             `json:"status"`
	Error           string             `json:"error"`
	RollbackAddress common.Address     `json:"rollbackAddress"`
	GasInfo         rocketpool.GasInfo `json:"gasInfo"`
}
type DelegateRollbackResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanSetUseLatestDelegateResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetUseLatestDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanStakeMinipoolResponse struct {
	Status   string             `json:"status"`
	Error    string             `json:"error"`
	CanStake bool               `json:"canStake"`
	GasInfo  rocketpool.GasInfo `json:"gasInfo"`
}
type StakeMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanPromoteMinipoolResponse struct {
	Status     string             `json:"status"`
	Error      string             `json:"error"`
	CanPromote bool               `json:"canPromote"`
	GasInfo    rocketpool.GasInfo `json:"gasInfo"`
}
type PromoteMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetUseLatestDelegateResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Setting bool   `json:"setting"`
}

type GetDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type GetPreviousDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type GetEffectiveDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type GetVanityArtifactsResponse struct {
	Status                 string         `json:"status"`
	Error                  string         `json:"error"`
	NodeAddress            common.Address `json:"nodeAddress"`
	MinipoolFactoryAddress common.Address `json:"minipoolFactoryAddress"`
	InitHash               common.Hash    `json:"initHash"`
}

type CanBeginReduceBondAmountResponse struct {
	Status                string                `json:"status"`
	Error                 string                `json:"error"`
	BondReductionDisabled bool                  `json:"bondReductionDisabled"`
	MinipoolVersionTooLow bool                  `json:"minipoolVersionTooLow"`
	Balance               uint64                `json:"balance"`
	BalanceTooLow         bool                  `json:"balanceTooLow"`
	MatchRequest          *big.Int              `json:"matchRequest"`
	BeaconState           beacon.ValidatorState `json:"beaconState"`
	InvalidBeaconState    bool                  `json:"invalidBeaconState"`
	CanReduce             bool                  `json:"canReduce"`
	GasInfo               rocketpool.GasInfo    `json:"gasInfo"`
}
type BeginReduceBondAmountResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanReduceBondAmountResponse struct {
	Status          string             `json:"status"`
	Error           string             `json:"error"`
	MinipoolVersion uint8              `json:"minipoolVersion"`
	CanReduce       bool               `json:"canReduce"`
	GasInfo         rocketpool.GasInfo `json:"gasInfo"`
}
type ReduceBondAmountResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type MinipoolRescueDissolvedDetails struct {
	Address         common.Address        `json:"address"`
	CanRescue       bool                  `json:"canRescue"`
	IsFinalized     bool                  `json:"isFinalized"`
	MinipoolStatus  types.MinipoolStatus  `json:"minipoolStatus"`
	MinipoolVersion uint8                 `json:"minipoolVersion"`
	BeaconBalance   *big.Int              `json:"beaconBalance"`
	BeaconState     beacon.ValidatorState `json:"beaconState"`
	GasInfo         rocketpool.GasInfo    `json:"gasInfo"`
}

type GetMinipoolRescueDissolvedDetailsForNodeResponse struct {
	Status  string                           `json:"status"`
	Error   string                           `json:"error"`
	Details []MinipoolRescueDissolvedDetails `json:"details"`
}
type RescueDissolvedMinipoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
