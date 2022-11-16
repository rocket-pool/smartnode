package node

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	v110_node "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0/node"
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
	bc, err := services.GetBeaconClient(c)
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
	response.AccountAddressFormatted = formatResolvedAddress(c, response.AccountAddress)

	// Check if Atlas is deployed
	isAtlasDeployed, err := rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas is deployed: %w", err)
	}
	response.IsAtlasDeployed = isAtlasDeployed

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
			response.WithdrawalAddressFormatted = formatResolvedAddress(c, response.WithdrawalAddress)
			response.PendingWithdrawalAddress = details.PendingWithdrawalAddress
			response.PendingWithdrawalAddressFormatted = formatResolvedAddress(c, response.PendingWithdrawalAddress)
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
		if !isAtlasDeployed {
			rocketNodeStakingAddress := cfg.Smartnode.GetV110NodeStakingAddress()
			response.MinipoolLimit, err = v110_node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil, &rocketNodeStakingAddress)
			return err
		} else {
			response.EthMatched, err = node.GetNodeEthMatched(rp, nodeAccount.Address, nil)
			if err != nil {
				return err
			}
			response.EthMatchedLimit, err = node.GetNodeEthMatchedLimit(rp, nodeAccount.Address, nil)
			return err
		}
	})

	// Get active and past votes from Snapshot, but treat errors as non-Fatal
	wg.Go(func() error {
		var err error
		r := &response.SnapshotResponse
		if cfg.Smartnode.GetSnapshotDelegationAddress() != "" {
			idHash := cfg.Smartnode.GetVotingSnapshotID()
			response.VotingDelegate, err = s.Delegation(nil, nodeAccount.Address, idHash)
			if err != nil {
				r.Error = err.Error()
				return nil
			}
			blankAddress := common.Address{}
			if response.VotingDelegate != blankAddress {
				response.VotingDelegateFormatted = formatResolvedAddress(c, response.VotingDelegate)
			}

			votedProposals, err := GetSnapshotVotedProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), nodeAccount.Address, response.VotingDelegate)
			if err != nil {
				r.Error = err.Error()
				return nil
			}
			r.ProposalVotes = votedProposals.Data.Votes
		}
		snapshotResponse, err := GetSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
		if err != nil {
			r.Error = err.Error()
			return nil
		}
		r.ActiveSnapshotProposals = snapshotResponse.Data.Proposals
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

	wg.Go(func() error {
		var err error
		response.IsFeeDistributorInitialized, err = node.GetFeeDistributorInitialized(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		feeRecipientInfo, err := rputils.GetFeeRecipientInfo(rp, bc, nodeAccount.Address, nil)
		if err == nil {
			response.FeeRecipientInfo = *feeRecipientInfo
			response.FeeDistributorBalance, err = rp.Client.BalanceAt(context.Background(), feeRecipientInfo.FeeDistributorAddress, nil)
		}
		return err
	})

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
