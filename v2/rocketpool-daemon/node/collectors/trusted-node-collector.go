package collectors

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Represents the collector for the user's trusted node
type TrustedNodeCollector struct {
	// The number of proposals per state
	proposalCount *prometheus.Desc

	// The number of votable proposals this node has yet to vote on
	unvotedProposalCount *prometheus.Desc

	// Tabular data for the votes for and against each proposal
	proposalTable *prometheus.Desc

	// The ETH balance of each trusted node
	ethBalance *prometheus.Desc

	// The balances submission participation of the ODAO members
	balancesParticipation *prometheus.Desc

	// The prices submission participation of the ODAO members
	pricesParticipation *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The logger
	logger *slog.Logger

	// Cached data
	cacheTime     time.Time
	cachedMetrics []prometheus.Metric

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new TrustedNodeCollector instance
func NewTrustedNodeCollector(logger *log.Logger, sp *services.ServiceProvider, stateLocker *StateLocker) *TrustedNodeCollector {
	subsystem := "trusted_node"
	sublogger := logger.With(slog.String(keys.RoutineKey, "ODAO Stats Collector"))
	return &TrustedNodeCollector{
		proposalCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposal_count"),
			"The number of proposals in each state",
			[]string{"state"}, nil,
		),
		unvotedProposalCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "unvoted_proposal_count"),
			"How many active proposals has this trusted node has not voted on",
			nil, nil,
		),
		proposalTable: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposal_table"),
			"Tabular data of each active proposal's for and against votes",
			[]string{"proposal", "category"}, nil,
		),
		ethBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_balance"),
			"The ETH balance of each trusted node",
			[]string{"member"}, nil,
		),
		balancesParticipation: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balances_participation"),
			"Whether each member has participated in the current balances update interval",
			[]string{"member"}, nil,
		),
		pricesParticipation: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prices_participation"),
			"Whether each member has participated in the current prices update interval",
			[]string{"member"}, nil,
		),
		sp:          sp,
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *TrustedNodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.proposalCount
	channel <- c.unvotedProposalCount
	channel <- c.proposalTable
	channel <- c.ethBalance
	channel <- c.balancesParticipation
	channel <- c.pricesParticipation
}

// Collect the latest metric values and pass them to Prometheus
func (c *TrustedNodeCollector) Collect(channel chan<- prometheus.Metric) {
	// Services
	rp := c.sp.GetRocketPool()
	cfg := c.sp.GetConfig()
	if !cfg.Metrics.EnableOdaoMetrics.Value {
		return
	}

	// Get the latest state
	state := c.stateLocker.GetState()
	if state == nil {
		return
	}

	pMgr, err := proposals.NewDaoProposalManager(rp)
	if err != nil {
		c.logger.Error("Error creating DAO proposal manager binding", log.Err(err))
		return
	}

	unvotedCount := float64(0)
	pendingCount := float64(0)
	activeCount := float64(0)
	succeededCount := float64(0)
	executedCount := float64(0)
	cancelledCount := float64(0)
	defeatedCount := float64(0)
	expiredCount := float64(0)

	// Get the number of DAO proposals
	err = rp.Query(nil, nil, pMgr.ProposalCount)
	if err != nil {
		c.logger.Error("Error getting DAO proposal count", log.Err(err))
		return
	}

	// Get the DAO proposals
	oDaoProps, _, err := pMgr.GetProposals(pMgr.ProposalCount.Formatted(), true, nil)
	if err != nil {
		c.logger.Error("Error getting DAO proposals", log.Err(err))
		return
	}

	// Map the member IDs by address
	addresses := make([]common.Address, len(state.OracleDaoMemberDetails))
	memberIds := make(map[common.Address]string)
	for i, member := range state.OracleDaoMemberDetails {
		addresses[i] = member.Address
		memberIds[member.Address] = member.ID
	}

	// Get member ETH balances
	ethBalances := make(map[string]float64)
	balances, err := rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		c.logger.Error("Error getting Oracle DAO member balances", log.Err(err))
		return
	}
	for i, member := range state.OracleDaoMemberDetails {
		ethBalances[member.ID] = eth.WeiToEth(balances[i])
	}

	// Calculate metrics
	activeProps := []*proposals.OracleDaoProposal{}
	for _, proposal := range oDaoProps {
		switch proposal.State.Formatted() {
		case types.ProposalState_Pending:
			pendingCount++
		case types.ProposalState_Active:
			activeCount++
			activeProps = append(activeProps, proposal)
		case types.ProposalState_Succeeded:
			succeededCount++
		case types.ProposalState_Executed:
			executedCount++
		case types.ProposalState_Cancelled:
			cancelledCount++
		case types.ProposalState_Defeated:
			defeatedCount++
		case types.ProposalState_Expired:
			expiredCount++
		}
	}

	// Get the local node's voting status
	nodeAddress, hasNodeAddress := c.sp.GetWallet().GetAddress()
	if hasNodeAddress {
		trusted := false
		for _, member := range state.OracleDaoMemberDetails {
			if member.Address == nodeAddress {
				trusted = true
				break
			}
		}

		if trusted {
			voteStatus := make([]bool, len(activeProps))
			err = rp.Query(func(mc *batch.MultiCaller) error {
				for i, prop := range activeProps {
					prop.GetMemberHasVoted(mc, &voteStatus[i], nodeAddress)
				}
				return nil
			}, nil)
			if err != nil {
				c.logger.Error("Error getting Oracle DAO voting status", log.Err(err))
				return
			}

			for _, voted := range voteStatus {
				if !voted {
					unvotedCount++
				}
			}
		}
	}

	// Update all the metrics
	channel <- prometheus.MustNewConstMetric(
		c.unvotedProposalCount, prometheus.GaugeValue, unvotedCount)
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, pendingCount, "pending")
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, activeCount, "active")
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, succeededCount, "succeeded")
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, executedCount, "executed")
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, cancelledCount, "cancelled")
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, defeatedCount, "defeated")
	channel <- prometheus.MustNewConstMetric(
		c.proposalCount, prometheus.GaugeValue, expiredCount, "expired")

	// Update balance metrics
	for memberId, balance := range ethBalances {
		channel <- prometheus.MustNewConstMetric(
			c.ethBalance, prometheus.GaugeValue, balance, memberId)
	}

	// Update proposal metrics
	for _, proposal := range oDaoProps {
		if proposal.State.Formatted() != types.ProposalState_Active {
			continue
		}
		channel <- prometheus.MustNewConstMetric(
			c.proposalTable, prometheus.GaugeValue, proposal.VotesFor.Formatted(), strconv.FormatUint(proposal.ID, 10), "for")
		channel <- prometheus.MustNewConstMetric(
			c.proposalTable, prometheus.GaugeValue, proposal.VotesAgainst.Formatted(), strconv.FormatUint(proposal.ID, 10), "against")
		channel <- prometheus.MustNewConstMetric(
			c.proposalTable, prometheus.GaugeValue, proposal.VotesRequired.Formatted(), strconv.FormatUint(proposal.ID, 10), "required")
	}

	// Include cached metrics
	for _, metric := range c.cachedMetrics {
		channel <- metric
	}
}
