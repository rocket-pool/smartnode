package node

import (
	"math"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)


func getRewards(c *cli.Context) (*api.NodeRewardsResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireEthClientSynced(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    cfg, err := services.GetConfig(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeRewardsResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Get the event log interval
    eventLogInterval, err := apiutils.GetEventLogInterval(cfg)
    if err != nil {
        return nil, err
    }

    var totalEffectiveStake *big.Int
    var totalRplSupply *big.Int
    var inflationInterval *big.Int
    var odaoSize uint64
    var nodeOperatorRewardsPercent float64
    var trustedNodeOperatorRewardsPercent float64

    // Sync
    var wg errgroup.Group

    // Check if the node is registered or not
    wg.Go(func() error {
        exists, err := node.GetNodeExists(rp, nodeAccount.Address, nil)
        if err == nil {
            response.Registered = exists
        }
        return err
    })

    // Get the node registration time
    wg.Go(func() error {
        time, err := rewards.GetNodeRegistrationTime(rp, nodeAccount.Address, nil)
        if err == nil {
            response.NodeRegistrationTime = time
        }
        return err
    })

    // Get node trusted status
    wg.Go(func() error {
        trusted, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
        if err == nil {
            response.Trusted = trusted
        }
        return err
    })

    // Get cumulative rewards
    wg.Go(func() error {
        rewards, err := rewards.CalculateLifetimeNodeRewards(rp, nodeAccount.Address, eventLogInterval)
        if err == nil {
            response.CumulativeRewards = eth.WeiToEth(rewards)
        }
        return err
    })

    // Get the start of the rewards checkpoint
    wg.Go(func() error {
        lastCheckpoint, err := rewards.GetClaimIntervalTimeStart(rp, nil)
        if err == nil {
            response.LastCheckpoint = lastCheckpoint
        }
        return err
    })

    // Get the rewards checkpoint interval
    wg.Go(func() error {
        rewardsInterval, err := rewards.GetClaimIntervalTime(rp, nil)
        if err == nil {
            response.RewardsInterval = rewardsInterval
        }
        return err
    })

    // Get the node's effective stake
    wg.Go(func() error {
        effectiveStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil)
        if err == nil {
            response.EffectiveRplStake = eth.WeiToEth(effectiveStake)
        }
        return err
    })

    // Get the node's total stake
    wg.Go(func() error {
        stake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
        if err == nil {
            response.TotalRplStake = eth.WeiToEth(stake)
        }
        return err
    })

    // Get the total network effective stake
    wg.Go(func() error {
        totalEffectiveStake, err = node.GetTotalEffectiveRPLStake(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the total RPL supply
    wg.Go(func() error {
        totalRplSupply, err = tokens.GetRPLTotalSupply(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the RPL inflation interval
    wg.Go(func() error {
        inflationInterval, err = tokens.GetRPLInflationIntervalRate(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the node operator rewards percent
    wg.Go(func() error {
        nodeOperatorRewardsPercent, err = rewards.GetNodeOperatorRewardsPercent(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Check if rewards are currently available from the previous checkpoint
    wg.Go(func() error {
        unclaimedRewardsWei, err := rewards.GetNodeClaimRewardsAmount(rp, nodeAccount.Address, nil)
        if err == nil {
            response.UnclaimedRewards = eth.WeiToEth(unclaimedRewardsWei)
        }
        return err
    })


    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }
    
    // Calculate the estimated rewards
    rewardsIntervalDays := response.RewardsInterval.Seconds() / (60*60*24)
    inflationPerDay := eth.WeiToEth(inflationInterval)
    totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
    if totalRplAtNextCheckpoint < 0 {
        totalRplAtNextCheckpoint = 0
    }

    if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
        response.EstimatedRewards = response.EffectiveRplStake / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
    }

    if response.Trusted {
        
        var wg2 errgroup.Group

        // Get the node registration time
        wg2.Go(func() error {
            time, err := rewards.GetTrustedNodeRegistrationTime(rp, nodeAccount.Address, nil)
            if err == nil {
                response.TrustedNodeRegistrationTime = time
            }
            return err
        })

        // Get cumulative ODAO rewards
        wg2.Go(func() error {
            rewards, err := rewards.CalculateLifetimeTrustedNodeRewards(rp, nodeAccount.Address, eventLogInterval)
            if err == nil {
                response.CumulativeTrustedRewards = eth.WeiToEth(rewards)
            }
            return err
        })

        // Get the ODAO member count
        wg2.Go(func() error {
            odaoSize, err = trustednode.GetMemberCount(rp, nil)
            if err != nil {
                return err
            }
            return nil
        })

        // Get the trusted node operator rewards percent
        wg2.Go(func() error {
            trustedNodeOperatorRewardsPercent, err = rewards.GetTrustedNodeOperatorRewardsPercent(rp, nil)
            if err != nil {
                return err
            }
            return nil
        })

        // Get the node's oDAO RPL stake
        wg2.Go(func() error {
            bond, err := trustednode.GetMemberRPLBondAmount(rp, nodeAccount.Address, nil)
            if err == nil {
                response.TrustedRplBond = eth.WeiToEth(bond)
            }
            return err
        })

        // Check if rewards are currently available from the previous checkpoint for the ODAO
        wg2.Go(func() error {
            unclaimedRewardsWei, err := rewards.GetTrustedNodeClaimRewardsAmount(rp, nodeAccount.Address, nil)
            if err == nil {
                response.UnclaimedTrustedRewards = eth.WeiToEth(unclaimedRewardsWei)
            }
            return err
        })

        // Wait for data
        if err := wg2.Wait(); err != nil {
            return nil, err
        }

        response.EstimatedTrustedRewards = totalRplAtNextCheckpoint * trustedNodeOperatorRewardsPercent / float64(odaoSize)
    
    }

    // Return response
    return &response, nil

}

