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

type MinipoolDissolveDetails struct {
	Address       common.Address `json:"address"`
	CanDissolve   bool           `json:"canDissolve"`
	InvalidStatus bool           `json:"invalidStatus"`
}
type GetMinipoolDissolveDetailsForNodeResponse struct {
	Status  string                    `json:"status"`
	Error   string                    `json:"error"`
	Details []MinipoolDissolveDetails `json:"details"`
}

type MinipoolExitDetails struct {
	Address       common.Address `json:"address"`
	CanExit       bool           `json:"canExit"`
	InvalidStatus bool           `json:"invalidStatus"`
}
type GetMinipoolExitDetailsForNodeResponse struct {
	Status  string                `json:"status"`
	Error   string                `json:"error"`
	Details []MinipoolExitDetails `json:"details"`
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

type MinipoolDelegateDetails struct {
	Address                         common.Address `json:"address"`
	Delegate                        common.Address `json:"delegate"`
	EffectiveDelegate               common.Address `json:"effectiveDelegate"`
	PreviousDelegate                common.Address `json:"previousDelegate"`
	UseLatestDelegate               bool           `json:"useLatestDelegate"`
	RollbackVersionTooLow           bool           `json:"rollbackVersionTooLow"`
	VersionTooLowToDisableUseLatest bool           `json:"versionTooLowToDisableUseLatest"`
}
type GetMinipoolDelegateDetailsForNodeResponse struct {
	Status         string                    `json:"status"`
	Error          string                    `json:"error"`
	LatestDelegate common.Address            `json:"latestDelegate"`
	Details        []MinipoolDelegateDetails `json:"details"`
}

type MinipoolCloseDetails struct {
	Address                     common.Address        `json:"address"`
	IsFinalized                 bool                  `json:"isFinalized"`
	Status                      types.MinipoolStatus  `json:"status"`
	Version                     uint8                 `json:"version"`
	Distributed                 bool                  `json:"distributed"`
	CanClose                    bool                  `json:"canClose"`
	Balance                     *big.Int              `json:"balance"`
	EffectiveBalance            *big.Int              `json:"effectiveBalance"`
	Refund                      *big.Int              `json:"refund"`
	UserDepositBalance          *big.Int              `json:"userDepositBalance"`
	BeaconState                 beacon.ValidatorState `json:"beaconState"`
	NodeShareOfEffectiveBalance *big.Int              `json:"nodeShareOfEffectiveBalance"`
}
type GetMinipoolCloseDetailsForNodeResponse struct {
	Status                      string                 `json:"status"`
	Error                       string                 `json:"error"`
	IsFeeDistributorInitialized bool                   `json:"isFeeDistributorInitialized"`
	Details                     []MinipoolCloseDetails `json:"details"`
}

type MinipoolDistributeDetails struct {
	Address                         common.Address       `json:"address"`
	Balance                         *big.Int             `json:"balance"`
	Refund                          *big.Int             `json:"refund"`
	DistributableBalance            *big.Int             `json:"distributableBalance"`
	NodeShareOfDistributableBalance *big.Int             `json:"nodeShareOfDistributableBalance"`
	Version                         uint8                `json:"version"`
	Status                          types.MinipoolStatus `json:"status"`
	IsFinalized                     bool                 `json:"isFinalized"`
	CanDistribute                   bool                 `json:"canDistribute"`
}
type GetMinipoolDistributeDetailsForNodeResponse struct {
	Status  string                      `json:"status"`
	Error   string                      `json:"error"`
	Details []MinipoolDistributeDetails `json:"details"`
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

type MinipoolPromoteDetails struct {
	Address    common.Address `json:"address"`
	CanPromote bool           `json:"canPromote"`
}
type GetMinipoolPromoteDetailsForNodeResponse struct {
	Status  string                   `json:"status"`
	Error   string                   `json:"error"`
	Details []MinipoolPromoteDetails `json:"details"`
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

type MinipoolBeginReduceBondDetails struct {
	Address               common.Address        `json:"address"`
	MinipoolVersionTooLow bool                  `json:"minipoolVersionTooLow"`
	Balance               uint64                `json:"balance"`
	BalanceTooLow         bool                  `json:"balanceTooLow"`
	MatchRequest          *big.Int              `json:"matchRequest"`
	BeaconState           beacon.ValidatorState `json:"beaconState"`
	InvalidElState        bool                  `json:"invalidElState"`
	InvalidBeaconState    bool                  `json:"invalidBeaconState"`
	CanReduce             bool                  `json:"canReduce"`
}
type GetMinipoolBeginReduceBondDetailsForNodeResponse struct {
	Status                string                           `json:"status"`
	Error                 string                           `json:"error"`
	BondReductionDisabled bool                             `json:"bondReductionDisabled"`
	Details               []MinipoolBeginReduceBondDetails `json:"details"`
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
