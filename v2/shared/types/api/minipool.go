package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
)

type MinipoolDetails struct {
	Address         common.Address         `json:"address"`
	ValidatorPubkey beacon.ValidatorPubkey `json:"validatorPubkey"`
	Version         uint8                  `json:"version"`
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
type MinipoolStatusData struct {
	Minipools      []MinipoolDetails `json:"minipools"`
	LatestDelegate common.Address    `json:"latestDelegate"`
}

type MinipoolRefundDetails struct {
	Address                   common.Address `json:"address"`
	InsufficientRefundBalance bool           `json:"insufficientRefundBalance"`
	CanRefund                 bool           `json:"canRefund"`
}
type MinipoolRefundDetailsData struct {
	Details []MinipoolRefundDetails `json:"details"`
}

type MinipoolDissolveDetails struct {
	Address       common.Address `json:"address"`
	CanDissolve   bool           `json:"canDissolve"`
	InvalidStatus bool           `json:"invalidStatus"`
}
type MinipoolDissolveDetailsData struct {
	Details []MinipoolDissolveDetails `json:"details"`
}

type MinipoolExitDetails struct {
	Address       common.Address `json:"address"`
	CanExit       bool           `json:"canExit"`
	InvalidStatus bool           `json:"invalidStatus"`
}
type MinipoolExitDetailsData struct {
	Details []MinipoolExitDetails `json:"details"`
}

type MinipoolCanChangeWithdrawalCredentialsData struct {
	CanChange bool `json:"canChange"`
}

type MinipoolProcessWithdrawalData struct {
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
type MinipoolDelegateDetailsData struct {
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
type MinipoolCloseDetailsData struct {
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
type MinipoolDistributeDetailsData struct {
	Details []MinipoolDistributeDetails `json:"details"`
}

type MinipoolStakeDetails struct {
	Address            common.Address       `json:"address"`
	State              types.MinipoolStatus `json:"state"`
	InvalidState       bool                 `json:"invalidState"`
	RemainingTime      time.Duration        `json:"remainingTime"`
	StillInScrubPeriod bool                 `json:"stillInScrubPeriod"`
	CanStake           bool                 `json:"canStake"`
}
type MinipoolStakeDetailsData struct {
	Details []MinipoolStakeDetails `json:"details"`
}

type MinipoolPromoteDetails struct {
	Address    common.Address `json:"address"`
	CanPromote bool           `json:"canPromote"`
}
type MinipoolPromoteDetailsData struct {
	Details []MinipoolPromoteDetails `json:"details"`
}

type MinipoolVanityArtifactsData struct {
	NodeAddress            common.Address `json:"nodeAddress"`
	MinipoolFactoryAddress common.Address `json:"minipoolFactoryAddress"`
	InitHash               common.Hash    `json:"initHash"`
}

type MinipoolBeginReduceBondDetails struct {
	Address               common.Address        `json:"address"`
	NodeDepositBalance    *big.Int              `json:"nodeDepositBalance"`
	NodeFee               *big.Int              `json:"nodeFee"`
	MinipoolVersionTooLow bool                  `json:"minipoolVersionTooLow"`
	Balance               uint64                `json:"balance"`
	BalanceTooLow         bool                  `json:"balanceTooLow"`
	AlreadyInWindow       bool                  `json:"alreadyInWindow"`
	MatchRequest          *big.Int              `json:"matchRequest"`
	BeaconState           beacon.ValidatorState `json:"beaconState"`
	InvalidElState        bool                  `json:"invalidElState"`
	InvalidBeaconState    bool                  `json:"invalidBeaconState"`
	AlreadyCancelled      bool                  `json:"alreadyCancelled"`
	NodeDepositTooLow     bool                  `json:"nodeDepositTooLow"`
	CanReduce             bool                  `json:"canReduce"`
}
type MinipoolBeginReduceBondDetailsData struct {
	BondReductionDisabled       bool                             `json:"bondReductionDisabled"`
	IsFeeDistributorInitialized bool                             `json:"isFeeDistributorInitialized"`
	BondReductionWindowStart    time.Duration                    `json:"bondReductionWindowStart"`
	BondReductionWindowLength   time.Duration                    `json:"bondReductionWindowLength"`
	Details                     []MinipoolBeginReduceBondDetails `json:"details"`
}

type MinipoolReduceBondDetails struct {
	Address               common.Address        `json:"address"`
	NodeDepositBalance    *big.Int              `json:"nodeDepositBalance"`
	NodeFee               *big.Int              `json:"nodeFee"`
	MinipoolVersionTooLow bool                  `json:"minipoolVersionTooLow"`
	Balance               uint64                `json:"balance"`
	BalanceTooLow         bool                  `json:"balanceTooLow"`
	OutOfWindow           bool                  `json:"outOfWindow"`
	BeaconState           beacon.ValidatorState `json:"beaconState"`
	InvalidElState        bool                  `json:"invalidElState"`
	InvalidBeaconState    bool                  `json:"invalidBeaconState"`
	AlreadyCancelled      bool                  `json:"alreadyCancelled"`
	CanReduce             bool                  `json:"canReduce"`
}
type MinipoolReduceBondDetailsData struct {
	BondReductionDisabled       bool                        `json:"bondReductionDisabled"`
	IsFeeDistributorInitialized bool                        `json:"isFeeDistributorInitialized"`
	BondReductionWindowStart    time.Duration               `json:"bondReductionWindowStart"`
	BondReductionWindowLength   time.Duration               `json:"bondReductionWindowLength"`
	Details                     []MinipoolReduceBondDetails `json:"details"`
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
type MinipoolRescueDissolvedDetailsData struct {
	Details []MinipoolRescueDissolvedDetails `json:"details"`
}
