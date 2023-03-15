package node

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	v110_node "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0/node"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
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
	isAtlasDeployed, err := state.IsAtlasDeployed(rp, nil)
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
			response.EthMatched, response.EthMatchedLimit, response.PendingMatchAmount, err = rputils.CheckCollateral(rp, nodeAccount.Address, nil)
			return err
		}
	})

	if isAtlasDeployed {
		wg.Go(func() error {
			var err error
			response.CreditBalance, err = node.GetNodeDepositCredit(rp, nodeAccount.Address, nil)
			return err
		})
	}

	// Get active and past votes from Snapshot, but treat errors as non-Fatal
	if s != nil {
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
	}

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
		feeRecipientInfo, err := rputils.GetFeeRecipientInfo_Legacy(rp, bc, nodeAccount.Address, nil)
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
		if isAtlasDeployed {
			var wg2 errgroup.Group
			var minStakeFraction *big.Int
			var maxStakeFraction *big.Int
			wg2.Go(func() error {
				var err error
				minStakeFraction, err = protocol.GetMinimumPerMinipoolStakeRaw(rp, nil)
				return err
			})
			wg2.Go(func() error {
				var err error
				minStakeFraction, err = protocol.GetMaximumPerMinipoolStakeRaw(rp, nil)
				return err
			})

			// Wait for data
			if err := wg2.Wait(); err != nil {
				return nil, err
			}

			// Calculate the *real* minimum, including the pending bond reductions
			trueMinimumStake := big.NewInt(0).Add(response.EthMatched, response.PendingMatchAmount)
			trueMinimumStake.Mul(trueMinimumStake, minStakeFraction)
			trueMinimumStake.Div(trueMinimumStake, rplPrice)

			// Calculate the *real* maximum, including the pending bond reductions
			trueMaximumStake := eth.EthToWei(32)
			trueMaximumStake.Mul(trueMaximumStake, big.NewInt(int64(activeMinipools)))
			trueMaximumStake.Sub(trueMaximumStake, response.EthMatched)
			trueMaximumStake.Sub(trueMaximumStake, response.PendingMatchAmount) // (32 * activeMinipools - ethMatched - pendingMatch)
			trueMaximumStake.Mul(trueMaximumStake, maxStakeFraction)
			trueMaximumStake.Div(trueMaximumStake, rplPrice)

			response.MinimumRplStake = trueMinimumStake
			response.MaximumRplStake = trueMaximumStake

			if response.EffectiveRplStake.Cmp(trueMinimumStake) < 0 {
				response.EffectiveRplStake.SetUint64(0)
			} else if response.EffectiveRplStake.Cmp(trueMaximumStake) > 0 {
				response.EffectiveRplStake.Set(trueMaximumStake)
			}

			response.BondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / (float64(activeMinipools)*32.0 - eth.WeiToEth(response.EthMatched) - eth.WeiToEth(response.PendingMatchAmount))
			response.BorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / (eth.WeiToEth(response.EthMatched) + eth.WeiToEth(response.PendingMatchAmount))
		} else {
			// Legacy behavior
			response.BorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / (float64(activeMinipools) * 16.0)
		}
	} else {
		response.BorrowedCollateralRatio = -1
	}

	// Return response
	return &response, nil

}
