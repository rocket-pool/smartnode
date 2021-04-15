package metrics

import (
    "math/big"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


type settingsGauges struct {
    flags                        *prometheus.GaugeVec
    lotMinimumEth                prometheus.Gauge
    lotMaximumEth                prometheus.Gauge
    lotDuration                  prometheus.Gauge
    lotStartingPrice             prometheus.Gauge
    lotReservePrice              prometheus.Gauge
    minimumDeposit               prometheus.Gauge
    maximumDepositPoolSize       prometheus.Gauge
    maximumDepositAssignments    prometheus.Gauge
    inflationIntervalRate        prometheus.Gauge
    inflationIntervalBlocks      prometheus.Gauge
    inflationStartBlock          prometheus.Gauge
    minipoolAmounts              *prometheus.GaugeVec
    minipoolLaunchTimeout        prometheus.Gauge
    minipoolWithdrawDelay        prometheus.Gauge
    nodeConsensusThreshold       prometheus.Gauge
    submitBalancesFrequency      prometheus.Gauge
    submitPricesFrequency        prometheus.Gauge
    networkNodeFee               *prometheus.GaugeVec
    targetRethCollateralRate     prometheus.Gauge
    nodeMinimumPerMinipoolStake  prometheus.Gauge
    nodeMaximumPerMinipoolStake  prometheus.Gauge
    rewardsClaimerPerc           *prometheus.GaugeVec
    rewardsClaimerPercUpdate     *prometheus.GaugeVec
    membersQuorum                prometheus.Gauge
    membersRPLBond               prometheus.Gauge
    membersMinipoolUnbondedMax   prometheus.Gauge
    membersChallengeCooldown     prometheus.Gauge
    membersChallengeWindow       prometheus.Gauge
    membersChallengeCost         prometheus.Gauge
    proposalCooldown             prometheus.Gauge
    proposalVoteBlocks           prometheus.Gauge
    proposalVoteDelayBlocks      prometheus.Gauge
    proposalExecuteBlocks        prometheus.Gauge
    proposalActionBlocks         prometheus.Gauge
}


// network metrics process
type settingsMetricsProcess struct {
    rp *rocketpool.RocketPool
    bc beacon.Client
    metrics settingsGauges
    logger log.ColorLogger
}


// Start network metrics process
func startSettingsMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startSettingsMetricsProcess")
    timer := time.NewTicker(interval)
    var p *settingsMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newSettingsMetricsProcess(c, logger)
        if p != nil && err == nil {
            break;
        }
    }

    // Update metrics on interval
    for ; true; <- timer.C {
        err = p.updateMetrics()
        if err != nil {
            // print error here instead of exit
            logger.Printlnf("Error in updateMetrics: %w", err)
        }
    }
    logger.Printlnf("Exit startSettingsMetricsProcess")
}


// Create new settingsMetricsProcess object
func newSettingsMetricsProcess(c *cli.Context, logger log.ColorLogger) (*settingsMetricsProcess, error) {

    logger.Printlnf("Enter newSettingsMetricsProcess")
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := settingsGauges {
        flags:                        promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:            "rocketpool",
                Subsystem:            "settings",
                Name:                 "flags_bool",
                Help:                 "settings flags on rocketpool protocol",
            },
            []string{"flag"},
        ),
        lotMinimumEth:                promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "lot_minimum_eth",
            Help:                     "minimum lot size in ETH",
        }),
        lotMaximumEth:                promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "lot_maximum_eth",
            Help:                     "maximum lot size in ETH",
        }),
        lotDuration:                  promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "lot_duration_blocks",
            Help:                     "lot duration in blocks",
        }),
        lotStartingPrice:             promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "lot_starting_price_ratio",
            Help:                     "starting price relative to current ETH price, as a fraction",
        }),
        lotReservePrice:              promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "lot_reserve_price_ratio",
            Help:                     "reserve price relative to current ETH price, as a fraction",
        }),
        minimumDeposit:               promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "deposit_minimum_eth",
            Help:                     "minimum deposit size",
        }),
        maximumDepositPoolSize:       promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "deposit_maximum_pool_eth",
            Help:                     "maximum size of deposit pool",
        }),
        maximumDepositAssignments:    promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "deposit_maximum_assignments",
            Help:                     "maximum deposit assignments per transaction",
        }),
        inflationIntervalRate:        promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "inflation_interval_rate",
            Help:                     "RPL inflation rate per interval",
        }),
        inflationIntervalBlocks:      promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "inflation_interval_blocks",
            Help:                     "RPL inflation interval in blocks",
        }),
        inflationStartBlock:          promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "inflation_start_block",
            Help:                     "RPL inflation start block",
        }),
        minipoolAmounts:              promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:            "rocketpool",
                Subsystem:            "settings",
                Name:                 "minipool_amounts",
                Help:                 "amount settings for rocketpool minipool",
            },
            []string{"category"},
        ),
        minipoolLaunchTimeout:        promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "minipool_launch_timeout_blocks",
            Help:                     "Timeout period in blocks for prelaunch minipools to launch",
        }),
        minipoolWithdrawDelay:        promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "minipool_withdraw_delay_blocks",
            Help:                     "Withdrawal delay in blocks before withdrawable minipools can be closed",
        }),
        nodeConsensusThreshold:       promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "network_node_consensus_threshold",
            Help:                     "threshold of trusted nodes that must reach consensus on oracle data to commit it",
        }),
        submitBalancesFrequency:      promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "network_submit_balances_frequency_blocks",
            Help:                     "frequency in blocks at which network balances should be submitted by trusted nodes",
        }),
        submitPricesFrequency:        promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "network_submit_prices_frequency_blocks",
            Help:                     "frequency in blocks at which network prices should be submitted by trusted nodes",
        }),
        networkNodeFee:               promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:            "rocketpool",
                Subsystem:            "settings",
                Name:                 "network_node_fee_rates",
                Help:                 "node fee settings",
            },
            []string{"type"},
        ),
        targetRethCollateralRate:     promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "network_target_reth_collateral_rate",
            Help:                     "target collateralization rate for the rETH contract as a fraction",
        }),
        nodeMinimumPerMinipoolStake:  promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "node_minimum_per_minipool_stake",
            Help:                     "minimum RPL stake per minipool as a fraction of assigned user ETH",
        }),
        nodeMaximumPerMinipoolStake:  promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "node_maximum_per_minipool_stake",
            Help:                     "maximum RPL stake per minipool as a fraction of assigned user ETH",
        }),
        rewardsClaimerPerc:           promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:            "rocketpool",
                Subsystem:            "settings",
                Name:                 "rewards_claimer_perc",
                Help:                 "claim amount for a claimer as a fraction",
            },
            []string{"contract"},
        ),
        rewardsClaimerPercUpdate:     promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:            "rocketpool",
                Subsystem:            "settings",
                Name:                 "rewards_claimer_perc_update_block",
                Help:                 "block that a claimer's share was last updated at",
            },
            []string{"contract"},
        ),
        membersQuorum:                promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "members_quorum",
            Help:                     "Member proposal quorum threshold",
        }),
        membersRPLBond:               promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "members_bond_rpl",
            Help:                     "RPL bond required for a member",
        }),
        membersMinipoolUnbondedMax:   promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "members_max_unbonded_minipool_count",
            Help:                     "maximum number of unbonded minipools a member can run",
        }),
        membersChallengeCooldown:     promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "members_challenge_cooldown_blocks",
            Help:                     "period a member must wait for before submitting another challenge, in blocks",
        }),
        membersChallengeWindow:       promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "members_challenge_window_blocks",
            Help:                     "period during which a member can respond to a challenge, in blocks",
        }),
        membersChallengeCost:         promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "members_challenge_cost_eth",
            Help:                     "The fee for a non-member to challenge a member, in eth",
        }),
        proposalCooldown:             promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "proposal_cooldown_blocks",
            Help:                     "cooldown period a member must wait after making a proposal before making another in blocks",
        }),
        proposalVoteBlocks:           promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "proposal_vote_duration_blocks",
            Help:                     "period a proposal can be voted on for in blocks",
        }),
        proposalVoteDelayBlocks:      promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "proposal_vote_delay_blocks",
            Help:                     "delay after creation before a proposal can be voted on in blocks",
        }),
        proposalExecuteBlocks:        promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "proposal_execute_blocks",
            Help:                     "period during which a passed proposal can be executed in blocks",
        }),
        proposalActionBlocks:         promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:                "rocketpool",
            Subsystem:                "settings",
            Name:                     "proposal_action_blocks",
            Help:                     "period during which an action can be performed on an executed proposal in blocks",
        }),
    }

    p := &settingsMetricsProcess {
        rp: rp,
        bc: bc,
        metrics: metrics,
        logger: logger,
    }

    logger.Printlnf("Exit newSettingsMetricsProcess")
    return p, nil
}


// Update settings metrics
func (p *settingsMetricsProcess) updateMetrics() error {
    p.logger.Printlnf("Enter settings updateMetrics")

    var wg errgroup.Group
    wg.Go(func() error {
        err := p.updateAuctionSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateDepositSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateInflationSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateMinipoolSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateNetworkSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateNodeSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateRewardsSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateMembersSettings()
        return err
    })
    wg.Go(func() error {
        err := p.updateProposalsSettings()
        return err
    })

    // Wait
    err := wg.Wait()
    p.logger.Printlnf("Exit settings updateMetrics")

    return err
}


func (p *settingsMetricsProcess) updateAuctionSettings() error {
    var createLotEnabled, bidOnLotEnabled bool
    var lotMinimumEthValue, lotMaximumEthValue *big.Int
    var lotDuration uint64
    var lotStartingPrice, lotReservePrice float64

    var wg errgroup.Group

    // Auction settings
    wg.Go(func() error {
        response, err := protocol.GetCreateLotEnabled(p.rp, nil)
        if err == nil {
            createLotEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetBidOnLotEnabled(p.rp, nil)
        if err == nil {
            bidOnLotEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetLotMinimumEthValue(p.rp, nil)
        if err == nil {
            lotMinimumEthValue = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetLotMaximumEthValue(p.rp, nil)
        if err == nil {
            lotMaximumEthValue = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetLotDuration(p.rp, nil)
        if err == nil {
            lotDuration = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetLotStartingPriceRatio(p.rp, nil)
        if err == nil {
            lotStartingPrice = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetLotReservePriceRatio(p.rp, nil)
        if err == nil {
            lotReservePrice = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }
    
    p.metrics.flags.With(prometheus.Labels{"flag":"CreateLotEnabled"}).Set(float64(B2i(createLotEnabled)))
    p.metrics.flags.With(prometheus.Labels{"flag":"BidOnLotEnabled"}).Set(float64(B2i(bidOnLotEnabled)))
    p.metrics.lotMinimumEth.Set(eth.WeiToEth(lotMinimumEthValue))
    p.metrics.lotMaximumEth.Set(eth.WeiToEth(lotMaximumEthValue))
    p.metrics.lotDuration.Set(float64(lotDuration))
    p.metrics.lotStartingPrice.Set(lotStartingPrice)
    p.metrics.lotReservePrice.Set(lotReservePrice)

    return nil
}


func (p *settingsMetricsProcess) updateDepositSettings() error {
    var maximumDepositAssignments uint64
    var depositEnabled, assignDepositEnabled bool
    var minimumDeposit, maximumDepositPoolSize *big.Int

    var wg errgroup.Group

    // Deposit settings
    wg.Go(func() error {
        response, err := protocol.GetDepositEnabled(p.rp, nil)
        if err == nil {
            depositEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetAssignDepositsEnabled(p.rp, nil)
        if err == nil {
            assignDepositEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinimumDeposit(p.rp, nil)
        if err == nil {
            minimumDeposit = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMaximumDepositPoolSize(p.rp, nil)
        if err == nil {
            maximumDepositPoolSize = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMaximumDepositAssignments(p.rp, nil)
        if err == nil {
            maximumDepositAssignments = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.flags.With(prometheus.Labels{"flag":"DepositEnabled"}).Set(float64(B2i(depositEnabled)))
    p.metrics.flags.With(prometheus.Labels{"flag":"DepositAssignmentsEnabled"}).Set(float64(B2i(assignDepositEnabled)))
    p.metrics.minimumDeposit.Set(eth.WeiToEth(minimumDeposit))
    p.metrics.maximumDepositPoolSize.Set(eth.WeiToEth(maximumDepositPoolSize))
    p.metrics.maximumDepositAssignments.Set(float64(maximumDepositAssignments))

    return nil
}


func (p *settingsMetricsProcess) updateInflationSettings() error {
    var inflationIntervalRate float64
    var inflationIntervalBlocks, inflationStartBlock uint64

    var wg errgroup.Group

    // Inflation settings
    wg.Go(func() error {
        response, err := protocol.GetInflationIntervalRate(p.rp, nil)
        if err == nil {
            inflationIntervalRate = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetInflationIntervalBlocks(p.rp, nil)
        if err == nil {
            inflationIntervalBlocks = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetInflationStartBlock(p.rp, nil)
        if err == nil {
            inflationStartBlock = response
        }
        return err
    })
    
    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.inflationIntervalRate.Set(inflationIntervalRate)
    p.metrics.inflationIntervalBlocks.Set(float64(inflationIntervalBlocks))
    p.metrics.inflationStartBlock.Set(float64(inflationStartBlock))

    return nil
}


func (p *settingsMetricsProcess) updateMinipoolSettings() error {
    var minipoolSubmitWithdrawEnabled bool
    var minipoolLaunchBalance, minipoolFullDepositNodeAmount, minipoolHalfDepositNodeAmount, minipoolEmptyDepositNodeAmount *big.Int
    var minipoolFullDepositUserAmount, minipoolHalfDepositUserAmount, minipoolEmptyDepositUserAmount *big.Int
    var minipoolLaunchTimeout, minipoolWithdrawalDelay uint64

    var wg errgroup.Group

    // Minipool settings
    wg.Go(func() error {
        response, err := protocol.GetMinipoolLaunchBalance(p.rp, nil)
        if err == nil {
            minipoolLaunchBalance = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolFullDepositNodeAmount(p.rp, nil)
        if err == nil {
            minipoolFullDepositNodeAmount = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolHalfDepositNodeAmount(p.rp, nil)
        if err == nil {
            minipoolHalfDepositNodeAmount = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolEmptyDepositNodeAmount(p.rp, nil)
        if err == nil {
            minipoolEmptyDepositNodeAmount = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolFullDepositUserAmount(p.rp, nil)
        if err == nil {
            minipoolFullDepositUserAmount = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolHalfDepositUserAmount(p.rp, nil)
        if err == nil {
            minipoolHalfDepositUserAmount = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolEmptyDepositUserAmount(p.rp, nil)
        if err == nil {
            minipoolEmptyDepositUserAmount = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolSubmitWithdrawableEnabled(p.rp, nil)
        if err == nil {
            minipoolSubmitWithdrawEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolLaunchTimeout(p.rp, nil)
        if err == nil {
            minipoolLaunchTimeout = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinipoolWithdrawalDelay(p.rp, nil)
        if err == nil {
            minipoolWithdrawalDelay = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"LaunchBalance"}).Set(eth.WeiToEth(minipoolLaunchBalance))
    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"FullDepositNodeAmount"}).Set(eth.WeiToEth(minipoolFullDepositNodeAmount))
    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"HalfDepositNodeAmount"}).Set(eth.WeiToEth(minipoolHalfDepositNodeAmount))
    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"EmptyDepositNodeAmount"}).Set(eth.WeiToEth(minipoolEmptyDepositNodeAmount))
    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"FullDepositUserAmount"}).Set(eth.WeiToEth(minipoolFullDepositUserAmount))
    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"HalfDepositUserAmount"}).Set(eth.WeiToEth(minipoolHalfDepositUserAmount))
    p.metrics.minipoolAmounts.With(prometheus.Labels{"category":"EmptyDepositUserAmount"}).Set(eth.WeiToEth(minipoolEmptyDepositUserAmount))
    p.metrics.flags.With(prometheus.Labels{"flag":"MinipoolSubmitWithdrawEnabled"}).Set(float64(B2i(minipoolSubmitWithdrawEnabled)))
    p.metrics.minipoolLaunchTimeout.Set(float64(minipoolLaunchTimeout))
    p.metrics.minipoolWithdrawDelay.Set(float64(minipoolWithdrawalDelay))

    return nil
}


func (p *settingsMetricsProcess) updateNetworkSettings() error {
    var submitBalancesEnabled, submitPricesEnabled, processWithdrawalEnabled bool
    var nodeConsensusThreshold, targetRethCollateralRate float64
    var submitBalancesFrequency, submitPricesFrequency uint64
    var minimumNodeFee, targetNodeFee, maximumNodeFee float64
    var nodeFeeDemandRange *big.Int

    var wg errgroup.Group

    // Network
    wg.Go(func() error {
        response, err := protocol.GetNodeConsensusThreshold(p.rp, nil)
        if err == nil {
            nodeConsensusThreshold = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetSubmitBalancesEnabled(p.rp, nil)
        if err == nil {
            submitBalancesEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetSubmitBalancesFrequency(p.rp, nil)
        if err == nil {
            submitBalancesFrequency = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetSubmitPricesEnabled(p.rp, nil)
        if err == nil {
            submitPricesEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetSubmitPricesFrequency(p.rp, nil)
        if err == nil {
            submitPricesFrequency = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetProcessWithdrawalsEnabled(p.rp, nil)
        if err == nil {
            processWithdrawalEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinimumNodeFee(p.rp, nil)
        if err == nil {
            minimumNodeFee = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetTargetNodeFee(p.rp, nil)
        if err == nil {
            targetNodeFee = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMaximumNodeFee(p.rp, nil)
        if err == nil {
            maximumNodeFee = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetNodeFeeDemandRange(p.rp, nil)
        if err == nil {
            nodeFeeDemandRange = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetTargetRethCollateralRate(p.rp, nil)
        if err == nil {
            targetRethCollateralRate = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.nodeConsensusThreshold.Set(nodeConsensusThreshold)
    p.metrics.flags.With(prometheus.Labels{"flag":"SubmitBalancesEnabled"}).Set(float64(B2i(submitBalancesEnabled)))
    p.metrics.submitBalancesFrequency.Set(float64(submitBalancesFrequency))
    p.metrics.flags.With(prometheus.Labels{"flag":"SubmitPricesEnabled"}).Set(float64(B2i(submitPricesEnabled)))
    p.metrics.submitPricesFrequency.Set(float64(submitPricesFrequency))
    p.metrics.flags.With(prometheus.Labels{"flag":"ProcessWithdrawalEnabled"}).Set(float64(B2i(processWithdrawalEnabled)))
    p.metrics.networkNodeFee.With(prometheus.Labels{"type":"MinimumNodeFee"}).Set(minimumNodeFee)
    p.metrics.networkNodeFee.With(prometheus.Labels{"type":"TargetNodeFee"}).Set(targetNodeFee)
    p.metrics.networkNodeFee.With(prometheus.Labels{"type":"MaximumNodeFee"}).Set(maximumNodeFee)
    p.metrics.networkNodeFee.With(prometheus.Labels{"type":"DemandRange"}).Set(eth.WeiToEth(nodeFeeDemandRange))
    p.metrics.targetRethCollateralRate.Set(targetRethCollateralRate)

    return nil
}


func (p *settingsMetricsProcess) updateNodeSettings() error {
    var nodeRegistrationEnabled, nodeDepositEnabled bool
    var minimumPerMinipoolStake, maximumPerMinipoolStake float64

    var wg errgroup.Group

    // Node settings
    wg.Go(func() error {
        response, err := protocol.GetNodeRegistrationEnabled(p.rp, nil)
        if err == nil {
            nodeRegistrationEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetNodeDepositEnabled(p.rp, nil)
        if err == nil {
            nodeDepositEnabled = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMinimumPerMinipoolStake(p.rp, nil)
        if err == nil {
            minimumPerMinipoolStake = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetMaximumPerMinipoolStake(p.rp, nil)
        if err == nil {
            maximumPerMinipoolStake = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.flags.With(prometheus.Labels{"flag":"NodeRegistrationEnabled"}).Set(float64(B2i(nodeRegistrationEnabled)))
    p.metrics.flags.With(prometheus.Labels{"flag":"NodeDepositEnabled"}).Set(float64(B2i(nodeDepositEnabled)))
    p.metrics.nodeMinimumPerMinipoolStake.Set(minimumPerMinipoolStake)
    p.metrics.nodeMaximumPerMinipoolStake.Set(maximumPerMinipoolStake)

    return nil
}


func (p *settingsMetricsProcess) updateRewardsSettings() error {
    var rewardsClaimerPercDAO, rewardsClaimerPercNode, rewardsClaimerPercTrustedNode, rewardsClaimerPercTotal float64
    var rewardsClaimerPercUpdateDAO, rewardsClaimerPercUpdateNode, rewardsClaimerPercUpdateTrustedNode uint64

    var wg errgroup.Group

    // Rewards settings
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimerPerc(p.rp, "rocketClaimDAO", nil)
        if err == nil {
            rewardsClaimerPercDAO = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimerPerc(p.rp, "rocketClaimNode", nil)
        if err == nil {
            rewardsClaimerPercNode = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimerPerc(p.rp, "rocketClaimTrustedNode", nil)
        if err == nil {
            rewardsClaimerPercTrustedNode = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimerPercBlockUpdated(p.rp, "rocketClaimDAO", nil)
        if err == nil {
            rewardsClaimerPercUpdateDAO = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimerPercBlockUpdated(p.rp, "rocketClaimNode", nil)
        if err == nil {
            rewardsClaimerPercUpdateNode = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimerPercBlockUpdated(p.rp, "rocketClaimTrustedNode", nil)
        if err == nil {
            rewardsClaimerPercUpdateTrustedNode = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := protocol.GetRewardsClaimersPercTotal(p.rp, nil)
        if err == nil {
            rewardsClaimerPercTotal = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.rewardsClaimerPerc.With(prometheus.Labels{"contract":"rocketClaimDAO"}).Set(rewardsClaimerPercDAO)
    p.metrics.rewardsClaimerPerc.With(prometheus.Labels{"contract":"rocketClaimNode"}).Set(rewardsClaimerPercNode)
    p.metrics.rewardsClaimerPerc.With(prometheus.Labels{"contract":"rocketClaimTrustedNode"}).Set(rewardsClaimerPercTrustedNode)
    p.metrics.rewardsClaimerPercUpdate.With(prometheus.Labels{"contract":"rocketClaimDAO"}).Set(float64(rewardsClaimerPercUpdateDAO))
    p.metrics.rewardsClaimerPercUpdate.With(prometheus.Labels{"contract":"rocketClaimNode"}).Set(float64(rewardsClaimerPercUpdateNode))
    p.metrics.rewardsClaimerPercUpdate.With(prometheus.Labels{"contract":"rocketClaimTrustedNode"}).Set(float64(rewardsClaimerPercUpdateTrustedNode))
    p.metrics.rewardsClaimerPerc.With(prometheus.Labels{"contract":"Total"}).Set(rewardsClaimerPercTotal)

    return nil
}


func (p *settingsMetricsProcess) updateMembersSettings() error {
    var membersQuorum float64
    var membersRplBond, membersChallendgeCost *big.Int
    var membersMinipoolUnbondedMax, membersChallengeCooldown, membersChallengeWindow uint64

    var wg errgroup.Group

    // Members settings
    wg.Go(func() error {
        response, err := trustednode.GetQuorum(p.rp, nil)
        if err == nil {
            membersQuorum = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetRPLBond(p.rp, nil)
        if err == nil {
            membersRplBond = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetMinipoolUnbondedMax(p.rp, nil)
        if err == nil {
            membersMinipoolUnbondedMax = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetChallengeCooldown(p.rp, nil)
        if err == nil {
            membersChallengeCooldown = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetChallengeWindow(p.rp, nil)
        if err == nil {
            membersChallengeWindow = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetChallengeCost(p.rp, nil)
        if err == nil {
            membersChallendgeCost = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.membersQuorum.Set(membersQuorum)
    p.metrics.membersRPLBond.Set(eth.WeiToEth(membersRplBond))
    p.metrics.membersMinipoolUnbondedMax.Set(float64(membersMinipoolUnbondedMax))
    p.metrics.membersChallengeCooldown.Set(float64(membersChallengeCooldown))
    p.metrics.membersChallengeWindow.Set(float64(membersChallengeWindow))
    p.metrics.membersChallengeCost.Set(eth.WeiToEth(membersChallendgeCost))

    return nil
}


func (p *settingsMetricsProcess) updateProposalsSettings() error {
    var proposalsCooldown, proposalsVoteBlocks, proposalsVoteDelayBlocks, proposalsExecuteBlocks, proposalsActionBlocks uint64

    var wg errgroup.Group

    // Proposals settings
    wg.Go(func() error {
        response, err := trustednode.GetProposalCooldown(p.rp, nil)
        if err == nil {
            proposalsCooldown = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetProposalVoteBlocks(p.rp, nil)
        if err == nil {
            proposalsVoteBlocks = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetProposalVoteDelayBlocks(p.rp, nil)
        if err == nil {
            proposalsVoteDelayBlocks = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetProposalExecuteBlocks(p.rp, nil)
        if err == nil {
            proposalsExecuteBlocks = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := trustednode.GetProposalActionBlocks(p.rp, nil)
        if err == nil {
            proposalsActionBlocks = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.proposalCooldown.Set(float64(proposalsCooldown))
    p.metrics.proposalVoteBlocks.Set(float64(proposalsVoteBlocks))
    p.metrics.proposalVoteDelayBlocks.Set(float64(proposalsVoteDelayBlocks))
    p.metrics.proposalExecuteBlocks.Set(float64(proposalsExecuteBlocks))
    p.metrics.proposalActionBlocks.Set(float64(proposalsActionBlocks))

    return nil
}

