package node

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func getStatus(c *cli.Context) (*api.NodeStatusResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	s, err := services.GetSnapshotDelegation(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeStatusResponse{}
	response.PenalizedMinipools = map[common.Address]uint64{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	// Get merge update deployment status
	isMergeUpdateDeployed, err := rputils.IsMergeUpdateDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error determining if merge update contracts have been deployed: %w", err)
	}
	response.IsMergeUpdateDeployed = isMergeUpdateDeployed

	// Sync
	var wg errgroup.Group

	// Get node trusted status
	wg.Go(func() error {
		trusted, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
		if err == nil {
			response.Trusted = trusted
		}
		return err
	})

	// Get node details
	wg.Go(func() error {
		details, err := node.GetNodeDetails(rp, nodeAccount.Address, nil)
		if err == nil {
			response.Registered = details.Exists
			response.WithdrawalAddress = details.WithdrawalAddress
			response.PendingWithdrawalAddress = details.PendingWithdrawalAddress
			response.TimezoneLocation = details.TimezoneLocation
		}
		return err
	})

	// Get node account balances
	wg.Go(func() error {
		var err error
		response.AccountBalances, err = tokens.GetBalances(rp, nodeAccount.Address, nil)
		return err
	})

	// Get staking details
	wg.Go(func() error {
		var err error
		response.RplStake, err = node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		response.EffectiveRplStake, err = node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		response.MinimumRplStake, err = node.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		response.MaximumRplStake, err = node.GetNodeMaximumRPLStake(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		response.MinipoolLimit, err = node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		if cfg.Smartnode.GetSnapshotDelegationAddress() != "" {
			idHash := cfg.Smartnode.GetVotingSnapshotID()
			response.VotingDelegate, err = s.Delegation(nil, nodeAccount.Address, idHash)
			if err != nil {
				return err
			}
			votedProposals, err := getSnapshotVotedProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), nodeAccount.Address, response.VotingDelegate)
			if err != nil {
				return err
			}
			response.VotedOnProposals = votedProposals.Data.VotedProposals
		}
		return err
	})
	// Get snapshot active proposals
	wg.Go(func() error {
		snapshotResponse, err := getSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
		if err != nil {
			return err
		}
		response.ActiveSnapshotProposals = snapshotResponse.Data.Proposals
		return nil
	})

	// Get node minipool counts
	wg.Go(func() error {
		details, err := getNodeMinipoolCountDetails(rp, nodeAccount.Address)
		if err == nil {
			response.MinipoolCounts.Total = len(details)
			for _, mpDetails := range details {
				if mpDetails.Penalties > 0 {
					response.PenalizedMinipools[mpDetails.Address] = mpDetails.Penalties
				}
				if mpDetails.Finalised {
					response.MinipoolCounts.Finalised++
				} else {
					switch mpDetails.Status {
					case types.Initialized:
						response.MinipoolCounts.Initialized++
					case types.Prelaunch:
						response.MinipoolCounts.Prelaunch++
					case types.Staking:
						response.MinipoolCounts.Staking++
					case types.Withdrawable:
						response.MinipoolCounts.Withdrawable++
					case types.Dissolved:
						response.MinipoolCounts.Dissolved++
					}
					if mpDetails.RefundAvailable {
						response.MinipoolCounts.RefundAvailable++
					}
					if mpDetails.WithdrawalAvailable {
						response.MinipoolCounts.WithdrawalAvailable++
					}
					if mpDetails.CloseAvailable {
						response.MinipoolCounts.CloseAvailable++
					}
				}
			}
		}
		return err
	})

	if isMergeUpdateDeployed {
		wg.Go(func() error {
			var err error
			response.IsFeeDistributorInitialized, err = node.GetFeeDistributorInitialized(rp, nodeAccount.Address, nil)
			return err
		})
		wg.Go(func() error {
			var err error
			response.FeeDistributorAddress, err = node.GetDistributorAddress(rp, nodeAccount.Address, nil)
			if err != nil {
				return err
			}
			response.FeeDistributorBalance, err = rp.Client.BalanceAt(context.Background(), response.FeeDistributorAddress, nil)
			return err
		})
		// Get Smoothing Pool registration status
		wg.Go(func() error {
			var err error
			response.IsInSmoothingPool, err = node.GetSmoothingPoolRegistrationState(rp, nodeAccount.Address, nil)
			return err
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Get withdrawal address balances
	if !bytes.Equal(nodeAccount.Address.Bytes(), response.WithdrawalAddress.Bytes()) {
		withdrawalBalances, err := tokens.GetBalances(rp, response.WithdrawalAddress, nil)
		if err != nil {
			return nil, err
		}
		response.WithdrawalBalances = withdrawalBalances
	}

	// Get the collateral ratio
	rplPrice, err := network.GetRPLPrice(rp, nil)
	if err != nil {
		return nil, err
	}
	activeMinipools := response.MinipoolCounts.Total - response.MinipoolCounts.Finalised
	if activeMinipools > 0 {
		response.CollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / (float64(activeMinipools) * 16.0)
	} else {
		response.CollateralRatio = -1
	}

	// Return response
	return &response, nil

}
