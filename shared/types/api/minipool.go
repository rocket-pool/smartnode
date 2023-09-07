package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
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
	Address         common.Address        `json:"address"`
	ValidatorPubkey types.ValidatorPubkey `json:"validatorPubkey"`
	Status          struct {
		Status      types.MinipoolStatus `json:"status"`
		StatusBlock uint64               `json:"statusBlock"`
		StatusTime  time.Time            `json:"statusTime"`
		IsVacant    bool                 `json:"isVacant"`
	} `json:"status"`
	DepositType types.MinipoolDeposit `json:"depositType"`
	Node        struct {
		Address         common.Address `json:"address"`
		Fee             float64        `json:"fee"`
		DepositBalance  *big.Int       `json:"depositBalance"`
		RefundBalance   *big.Int       `json:"refundBalance"`
		DepositAssigned bool           `json:"depositAssigned"`
	}
	User struct {
		DepositBalance      *big.Int  `json:"depositBalance"`
		DepositAssigned     bool      `json:"depositAssigned"`
		DepositAssignedTime time.Time `json:"depositAssignedTime"`
	} `json:"user"`
	Balances struct {
		Eth            *big.Int `json:"eth"`
		Reth           *big.Int `json:"reth"`
		Rpl            *big.Int `json:"rpl"`
		FixedSupplyRpl *big.Int `json:"fixedSupplyRpl"`
	} `json:"balances"`
	NodeShareOfEthBalance *big.Int `json:"nodeShareOfEthBalance"`
	Validator             struct {
		Exists      bool     `json:"exists"`
		Active      bool     `json:"active"`
		Index       string   `json:"index"`
		Balance     *big.Int `json:"balance"`
		NodeBalance *big.Int `json:"nodeBalance"`
	} `json:"validator"`
	CanStake            bool                  `json:"canStake"`
	CanPromote          bool                  `json:"canPromote"`
	Queue               minipool.QueueDetails `json:"queue"`
	RefundAvailable     bool                  `json:"refundAvailable"`
	WithdrawalAvailable bool                  `json:"withdrawalAvailable"`
	CloseAvailable      bool                  `json:"closeAvailable"`
	Finalised           bool                  `json:"finalised"`
	UseLatestDelegate   bool                  `json:"useLatestDelegate"`
	Delegate            common.Address        `json:"delegate"`
	PreviousDelegate    common.Address        `json:"previousDelegate"`
	EffectiveDelegate   common.Address        `json:"effectiveDelegate"`
	TimeUntilDissolve   time.Duration         `json:"timeUntilDissolve"`
	Penalties           uint64                `json:"penalties"`
	ReduceBondTime      time.Time             `json:"reduceBondTime"`
	ReduceBondCancelled bool                  `json:"reduceBondCancelled"`
}

type MinipoolRefundDetails struct {
	Address                   common.Address `json:"address"`
	InsufficientRefundBalance bool           `json:"insufficientRefundBalance"`
	CanRefund                 bool           `json:"canRefund"`
}
type GetMinipoolRefundDetailsForNodeResponse struct {
	Status  string                  `json:"status"`
	Error   string                  `json:"error"`
	Details []MinipoolRefundDetails `json:"details"`
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

type MinipoolStakeDetails struct {
	Address            common.Address       `json:"address"`
	State              types.MinipoolStatus `json:"state"`
	InvalidState       bool                 `json:"invalidState"`
	RemainingTime      time.Duration        `json:"remainingTime"`
	StillInScrubPeriod bool                 `json:"stillInScrubPeriod"`
	CanStake           bool                 `json:"canStake"`
}
type GetMinipoolStakeDetailsForNodeResponse struct {
	Status  string                 `json:"status"`
	Error   string                 `json:"error"`
	Details []MinipoolStakeDetails `json:"details"`
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
	AlreadyInWindow       bool                  `json:"alreadyInWindow"`
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

type MinipoolReduceBondDetails struct {
	Address               common.Address        `json:"address"`
	MinipoolVersionTooLow bool                  `json:"minipoolVersionTooLow"`
	Balance               uint64                `json:"balance"`
	BalanceTooLow         bool                  `json:"balanceTooLow"`
	OutOfWindow           bool                  `json:"outOfWindow"`
	BeaconState           beacon.ValidatorState `json:"beaconState"`
	InvalidElState        bool                  `json:"invalidElState"`
	InvalidBeaconState    bool                  `json:"invalidBeaconState"`
	CanReduce             bool                  `json:"canReduce"`
}
type GetMinipoolReduceBondDetailsForNodeResponse struct {
	Status                string                      `json:"status"`
	Error                 string                      `json:"error"`
	BondReductionDisabled bool                        `json:"bondReductionDisabled"`
	Details               []MinipoolReduceBondDetails `json:"details"`
}

type MinipoolRescueDissolvedDetails struct {
	Address            common.Address        `json:"address"`
	CanRescue          bool                  `json:"canRescue"`
	IsFinalized        bool                  `json:"isFinalized"`
	MinipoolState      types.MinipoolStatus  `json:"minipoolStatus"`
	InvalidElState     bool                  `json:"invalidElState"`
	MinipoolVersion    uint8                 `json:"minipoolVersion"`
	BeaconBalance      *big.Int              `json:"beaconBalance"`
	BeaconState        beacon.ValidatorState `json:"beaconState"`
	InvalidBeaconState bool                  `json:"invalidBeaconState"`
	HasFullBalance     bool                  `json:"hasFullBalance"`
}
type GetMinipoolRescueDissolvedDetailsForNodeResponse struct {
	Status  string                           `json:"status"`
	Error   string                           `json:"error"`
	Details []MinipoolRescueDissolvedDetails `json:"details"`
}
