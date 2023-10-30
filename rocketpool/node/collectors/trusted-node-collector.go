package collectors

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	oracleutils "github.com/rocket-pool/rocketpool-go/utils/oracle"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/config"
	"golang.org/x/sync/errgroup"
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

	// Cached data
	cacheTime     time.Time
	cachedMetrics []prometheus.Metric

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new NodeCollector instance
func NewTrustedNodeCollector(sp *services.ServiceProvider, stateLocker *StateLocker) *TrustedNodeCollector {
	subsystem := "trusted_node"
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
		stateLocker: stateLocker,
		logPrefix:   "ODAO Stats Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *TrustedNodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.proposalCount
	channel <- collector.unvotedProposalCount
	channel <- collector.proposalTable
	channel <- collector.ethBalance
	channel <- collector.balancesParticipation
	channel <- collector.pricesParticipation
}

// Caches slow to process metrics so it doesn't have to be processed every second
func (collector *TrustedNodeCollector) collectSlowMetrics(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, memberIds map[common.Address]string) {
	intervalSize, err := cfg.GetEventLogInterval()
	if err != nil {
		collector.logError(fmt.Errorf("error getting event log interval size: %w", err))
		return
	}

	pCalc, err := oracleutils.NewTrustedNodeParticipationCalculator(rp)
	if err != nil {
		collector.logError(fmt.Errorf("error creating oracle DAO participation calculator: %w", err))
		return
	}

	// Create a new cached metrics array to populate
	collector.cachedMetrics = make([]prometheus.Metric, 0)

	// Sync
	var wg errgroup.Group

	var balancesParticipation map[common.Address]bool
	var pricesParticipation map[common.Address]bool

	// Get the balances participation data
	wg.Go(func() error {
		var err error
		balancesParticipation, err = pCalc.GetTrustedNodeLatestBalancesParticipation(rp, big.NewInt(int64(intervalSize)), nil)
		if err != nil {
			return fmt.Errorf("Error getting trusted node balances participation data: %w", err)
		}
		return nil
	})

	// Get the prices participation data
	wg.Go(func() error {
		var err error
		pricesParticipation, err = pCalc.GetTrustedNodeLatestPricesParticipation(rp, big.NewInt(int64(intervalSize)), nil)
		if err != nil {
			return fmt.Errorf("Error getting trusted node prices participation data: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	// Balances participation
	for member, status := range balancesParticipation {
		value := float64(0)
		if status {
			value = 1
		}
		collector.cachedMetrics = append(collector.cachedMetrics, prometheus.MustNewConstMetric(collector.balancesParticipation, prometheus.GaugeValue, value, memberIds[member]))
	}

	// Prices participation
	for member, status := range pricesParticipation {
		value := float64(0)
		if status {
			value = 1
		}
		collector.cachedMetrics = append(collector.cachedMetrics, prometheus.MustNewConstMetric(collector.pricesParticipation, prometheus.GaugeValue, value, memberIds[member]))
	}
}

// Collect the latest metric values and pass them to Prometheus
func (collector *TrustedNodeCollector) Collect(channel chan<- prometheus.Metric) {
	// Services
	rp := collector.sp.GetRocketPool()
	cfg := collector.sp.GetConfig()
	if !cfg.EnableODaoMetrics.Value.(bool) {
		return
	}

	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	pMgr, err := proposals.NewDaoProposalManager(rp)
	if err != nil {
		collector.logError(fmt.Errorf("error creating DAO proposal manager binding: %w", err))
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
		collector.logError(fmt.Errorf("error getting DAO proposal count: %w", err))
		return
	}

	// Get the DAO proposals
	_, oDaoProps, err := pMgr.GetProposals(pMgr.ProposalCount.Formatted(), true, nil)
	if err != nil {
		collector.logError(fmt.Errorf("error getting DAO proposals: %w", err))
		return
	}

	// Map the member IDs by address
	addresses := make([]common.Address, len(state.OracleDaoMemberDetails))
	memberIds := make(map[common.Address]string)
	for i, member := range state.OracleDaoMemberDetails {
		addresses[i] = member.Address
		memberIds[member.Address] = member.ID
	}

	// Only collect fresh participation metrics from chain every 60 seconds as it updates infrequently and takes longer to collect
	if time.Since(collector.cacheTime) > (time.Second * 60) {
		collector.collectSlowMetrics(rp, cfg, memberIds)
		collector.cacheTime = time.Now()
	}

	// Get member ETH balances
	ethBalances := make(map[string]float64)
	balances, err := rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		collector.logError(fmt.Errorf("error getting Oracle DAO member balances: %w", err))
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
	nodeAddress, hasNodeAddress := collector.sp.GetWallet().GetAddress()
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
				collector.logError(fmt.Errorf("error getting Oracle DAO voting status: %w", err))
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
		collector.unvotedProposalCount, prometheus.GaugeValue, unvotedCount)
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, pendingCount, "pending")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, activeCount, "active")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, succeededCount, "succeeded")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, executedCount, "executed")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, cancelledCount, "cancelled")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, defeatedCount, "defeated")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, expiredCount, "expired")

	// Update balance metrics
	for memberId, balance := range ethBalances {
		channel <- prometheus.MustNewConstMetric(
			collector.ethBalance, prometheus.GaugeValue, balance, memberId)
	}

	// Update proposal metrics
	for _, proposal := range oDaoProps {
		if proposal.State.Formatted() != types.ProposalState_Active {
			continue
		}
		channel <- prometheus.MustNewConstMetric(
			collector.proposalTable, prometheus.GaugeValue, proposal.VotesFor.Formatted(), strconv.FormatUint(proposal.ID, 10), "for")
		channel <- prometheus.MustNewConstMetric(
			collector.proposalTable, prometheus.GaugeValue, proposal.VotesAgainst.Formatted(), strconv.FormatUint(proposal.ID, 10), "against")
		channel <- prometheus.MustNewConstMetric(
			collector.proposalTable, prometheus.GaugeValue, proposal.VotesRequired.Formatted(), strconv.FormatUint(proposal.ID, 10), "required")
	}

	// Include cached metrics
	for _, metric := range collector.cachedMetrics {
		channel <- metric
	}
}

// Log error messages
func (collector *TrustedNodeCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
