package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type PDAOProposalsResponse struct {
	Status    string                `json:"status"`
	Error     string                `json:"error"`
	Proposals []dao.ProposalDetails `json:"proposals"`
}

type PDAOProposalResponse struct {
	Status   string              `json:"status"`
	Error    string              `json:"error"`
	Proposal dao.ProposalDetails `json:"proposal"`
}

type CanCancelPDAOProposalResponse struct {
	Status          string             `json:"status"`
	Error           string             `json:"error"`
	CanCancel       bool               `json:"canCancel"`
	DoesNotExist    bool               `json:"doesNotExist"`
	InvalidState    bool               `json:"invalidState"`
	InvalidProposer bool               `json:"invalidProposer"`
	GasInfo         rocketpool.GasInfo `json:"gasInfo"`
}
type CancelPDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanVoteOnPDAOProposalResponse struct {
	Status            string             `json:"status"`
	Error             string             `json:"error"`
	CanVote           bool               `json:"canVote"`
	DoesNotExist      bool               `json:"doesNotExist"`
	InvalidState      bool               `json:"invalidState"`
	InsufficientPower bool               `json:"insufficientPower"`
	AlreadyVoted      bool               `json:"alreadyVoted"`
	VotingPower       *big.Int           `json:"votingPower"`
	GasInfo           rocketpool.GasInfo `json:"gasInfo"`
}
type VoteOnPDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanExecutePDAOProposalResponse struct {
	Status       string             `json:"status"`
	Error        string             `json:"error"`
	CanExecute   bool               `json:"canExecute"`
	DoesNotExist bool               `json:"doesNotExist"`
	InvalidState bool               `json:"invalidState"`
	GasInfo      rocketpool.GasInfo `json:"gasInfo"`
}
type ExecutePDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetPDAOSettingsResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Auction struct {
		IsCreateLotEnabled    bool     `json:"isCreateLotEnabled"`
		IsBidOnLotEnabled     bool     `json:"isBidOnLotEnabled"`
		LotMinimumEthValue    *big.Int `json:"lotMinimumEthValue"`
		LotMaximumEthValue    *big.Int `json:"lotMaximumEthValue"`
		LotDuration           uint64   `json:"lotDuration"`
		LotStartingPriceRatio float64  `json:"lotStartingPriceRatio"`
		LotReservePriceRatio  float64  `json:"lotReservePriceRatio"`
	} `json:"auction"`

	Deposit struct {
		IsDepositingEnabled                    bool     `json:"isDepositingEnabled"`
		AreDepositAssignmentsEnabled           bool     `json:"areDepositAssignmentsEnabled"`
		MinimumDeposit                         *big.Int `json:"minimumDeposit"`
		MaximumDepositPoolSize                 *big.Int `json:"maximumDepositPoolSize"`
		MaximumAssignmentsPerDeposit           uint64   `json:"maximumAssignmentsPerDeposit"`
		MaximumSocialisedAssignmentsPerDeposit uint64   `json:"maximumSocialisedAssignmentsPerDeposit"`
		DepositFee                             float64  `json:"depositFee"`
	} `json:"deposit"`

	Inflation struct {
		IntervalRate float64   `json:"intervalRate"`
		StartTime    time.Time `json:"startTime"`
	} `json:"inflation"`

	Minipool struct {
		IsSubmitWithdrawableEnabled bool          `json:"isSubmitWithdrawableEnabled"`
		LaunchTimeout               time.Duration `json:"launchTimeout"`
		IsBondReductionEnabled      bool          `json:"isBondReductionEnabled"`
		MaximumCount                uint64        `json:"maximumCount"`
		UserDistributeWindowStart   time.Duration `json:"userDistributeWindowStart"`
		UserDistributeWindowLength  time.Duration `json:"userDistributeWindowLength"`
	} `json:"minipool"`

	Network struct {
		OracleDaoConsensusThreshold float64  `json:"oracleDaoConsensusThreshold"`
		NodePenaltyThreshold        float64  `json:"nodePenaltyThreshold"`
		PerPenaltyRate              float64  `json:"perPenaltyRate"`
		IsSubmitBalancesEnabled     bool     `json:"isSubmitBalancesEnabled"`
		SubmitBalancesFrequency     uint64   `json:"submitBalancesFrequency"`
		IsSubmitPricesEnabled       bool     `json:"isSubmitPricesEnabled"`
		SubmitPricesFrequency       uint64   `json:"submitPricesFrequency"`
		MinimumNodeFee              float64  `json:"minimumNodeFee"`
		TargetNodeFee               float64  `json:"targetNodeFee"`
		MaximumNodeFee              float64  `json:"maximumNodeFee"`
		NodeFeeDemandRange          *big.Int `json:"nodeFeeDemandRange"`
		TargetRethCollateralRate    float64  `json:"targetRethCollateralRate"`
		IsSubmitRewardsEnabled      bool     `json:"isSubmitRewardsEnabled"`
	} `json:"network"`

	Node struct {
		IsRegistrationEnabled              bool    `json:"isRegistrationEnabled"`
		IsSmoothingPoolRegistrationEnabled bool    `json:"isSmoothingPoolRegistrationEnabled"`
		IsDepositingEnabled                bool    `json:"isDepositingEnabled"`
		AreVacantMinipoolsEnabled          bool    `json:"areVacantMinipoolsEnabled"`
		MinimumPerMinipoolStake            float64 `json:"minimumPerMinipoolStake"`
		MaximumPerMinipoolStake            float64 `json:"maximumPerMinipoolStake"`
	} `json:"node"`

	Proposals struct {
		VoteTime        time.Duration `json:"voteTime"`
		VoteDelayTime   time.Duration `json:"voteDelayTime"`
		ExecuteTime     time.Duration `json:"executeTime"`
		ProposalBond    *big.Int      `json:"proposalBond"`
		ChallengeBond   *big.Int      `json:"challengeBond"`
		ChallengePeriod time.Duration `json:"challengePeriod"`
		Quorum          float64       `json:"quorum"`
		VetoQuorum      float64       `json:"vetoQuorum"`
		MaxBlockAge     uint64        `json:"maxBlockAge"`
	} `json:"proposals"`

	Rewards struct {
		IntervalTime time.Duration `json:"intervalTime"`
	} `json:"rewards"`
}

type CanProposePDAOSettingResponse struct {
	Status          string             `json:"status"`
	Error           string             `json:"error"`
	CanPropose      bool               `json:"canPropose"`
	InsufficientRpl bool               `json:"proposalCooldownActive"`
	StakedRpl       *big.Int           `json:"stakedRpl"`
	LockedRpl       *big.Int           `json:"lockedRpl"`
	ProposalBond    *big.Int           `json:"proposalBond"`
	BlockNumber     uint32             `json:"blockNumber"`
	Pollard         string             `json:"pollard"`
	GasInfo         rocketpool.GasInfo `json:"gasInfo"`
}
type ProposePDAOSettingResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
