package node

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
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
		rplStake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.RplStake.Set(rplStake)
		return nil
	})
	wg.Go(func() error {
		effectiveRplStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.EffectiveRplStake.Set(effectiveRplStake)
		return nil
	})
	wg.Go(func() error {
		minimumRplStake, err := node.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.MinimumRplStake.Set(minimumRplStake)
		return nil
	})
	wg.Go(func() error {
		maximumRplStake, err := node.GetNodeMaximumRPLStake(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.MaximumRplStake.Set(maximumRplStake)
		return nil
	})
	wg.Go(func() error {
		ethMatched, ethMatchedLimit, pendingMatchAmount, err := rputils.CheckCollateral(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.EthMatched.Set(ethMatched)
		response.EthMatchedLimit.Set(ethMatchedLimit)
		response.PendingMatchAmount.Set(pendingMatchAmount)
		return nil
	})

	wg.Go(func() error {
		creditBalance, err := node.GetNodeDepositCredit(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.CreditBalance.Set(creditBalance)
		return nil
	})

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
		feeRecipientInfo, err := rputils.GetFeeRecipientInfo(rp, bc, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.FeeRecipientInfo = *feeRecipientInfo
		feeDistributorBalance, err := rp.Client.BalanceAt(context.Background(), feeRecipientInfo.FeeDistributorAddress, nil)
		if err != nil {
			return err
		}
		response.FeeDistributorBalance.Set(feeDistributorBalance)
		return nil
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
			maxStakeFraction, err = protocol.GetMaximumPerMinipoolStakeRaw(rp, nil)
			return err
		})

		// Wait for data
		if err := wg2.Wait(); err != nil {
			return nil, err
		}

		// Calculate the *real* minimum, including the pending bond reductions
		trueMinimumStake := big.NewInt(0).Add(&response.EthMatched, &response.PendingMatchAmount)
		trueMinimumStake.Mul(trueMinimumStake, minStakeFraction)
		trueMinimumStake.Div(trueMinimumStake, rplPrice)

		// Calculate the *real* maximum, including the pending bond reductions
		trueMaximumStake := eth.EthToWei(32)
		trueMaximumStake.Mul(trueMaximumStake, big.NewInt(int64(activeMinipools)))
		trueMaximumStake.Sub(trueMaximumStake, &response.EthMatched)
		trueMaximumStake.Sub(trueMaximumStake, &response.PendingMatchAmount) // (32 * activeMinipools - ethMatched - pendingMatch)
		trueMaximumStake.Mul(trueMaximumStake, maxStakeFraction)
		trueMaximumStake.Div(trueMaximumStake, rplPrice)

		response.MinimumRplStake.Set(trueMinimumStake)
		response.MaximumRplStake.Set(trueMaximumStake)

		if response.EffectiveRplStake.Cmp(trueMinimumStake) < 0 {
			response.EffectiveRplStake.SetUint64(0)
		} else if response.EffectiveRplStake.Cmp(trueMaximumStake) > 0 {
			response.EffectiveRplStake.Set(trueMaximumStake)
		}

		response.BondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(&response.RplStake) / (float64(activeMinipools)*32.0 - eth.WeiToEth(&response.EthMatched) - eth.WeiToEth(&response.PendingMatchAmount))
		response.BorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(&response.RplStake) / (eth.WeiToEth(&response.EthMatched) + eth.WeiToEth(&response.PendingMatchAmount))

		// Calculate the "eligible" info (ignoring pending bond reductions) based on the Beacon Chain
		_, _, pendingEligibleBorrowedEth, pendingEligibleBondedEth, err := getTrueBorrowAndBondAmounts(rp, bc, nodeAccount.Address)
		if err != nil {
			return nil, fmt.Errorf("error calculating eligible borrowed and bonded amounts: %w", err)
		}

		// Calculate the "eligible real" minimum based on the Beacon Chain, including pending bond reductions
		pendingTrueMinimumStake := big.NewInt(0).Mul(pendingEligibleBorrowedEth, minStakeFraction)
		pendingTrueMinimumStake.Div(pendingTrueMinimumStake, rplPrice)

		// Calculate the "eligible real" maximum based on the Beacon Chain, including the pending bond reductions
		pendingTrueMaximumStake := big.NewInt(0).Mul(pendingEligibleBondedEth, maxStakeFraction)
		pendingTrueMaximumStake.Div(pendingTrueMaximumStake, rplPrice)

		response.PendingMinimumRplStake.Set(pendingTrueMinimumStake)
		response.PendingMaximumRplStake.Set(pendingTrueMaximumStake)

		response.PendingEffectiveRplStake.Set(&response.RplStake)
		if response.PendingEffectiveRplStake.Cmp(pendingTrueMinimumStake) < 0 {
			response.PendingEffectiveRplStake.SetUint64(0)
		} else if response.PendingEffectiveRplStake.Cmp(pendingTrueMaximumStake) > 0 {
			response.PendingEffectiveRplStake.Set(pendingTrueMaximumStake)
		}

		response.PendingBondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(&response.RplStake) / eth.WeiToEth(pendingEligibleBondedEth)
		response.PendingBorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(&response.RplStake) / eth.WeiToEth(pendingEligibleBorrowedEth)
	} else {
		response.BorrowedCollateralRatio = -1
	}

	// Return response
	return &response, nil

}

// Calculate the true borrowed and bonded ETH amounts for a node based on the Beacon status of the minipools
func getTrueBorrowAndBondAmounts(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address) (*big.Int, *big.Int, *big.Int, *big.Int, error) {

	mpDetails, err := minipool.GetNodeMinipools(rp, nodeAddress, nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error loading minipool details: %w", err)
	}

	beaconHead, err := bc.GetBeaconHead()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting beacon head: %w", err)
	}

	pubkeys := make([]types.ValidatorPubkey, len(mpDetails))
	nodeDeposits := make([]*big.Int, len(mpDetails))
	userDeposits := make([]*big.Int, len(mpDetails))
	pendingNodeDeposits := make([]*big.Int, len(mpDetails))
	pendingUserDeposits := make([]*big.Int, len(mpDetails))

	// Data
	var wg errgroup.Group

	for i, mpd := range mpDetails {
		if !mpd.Exists {
			nodeDeposits[i] = big.NewInt(0)
			userDeposits[i] = big.NewInt(0)
			pendingNodeDeposits[i] = big.NewInt(0)
			pendingUserDeposits[i] = big.NewInt(0)
			continue
		}

		i := i
		address := mpd.Address
		pubkeys[i] = mpd.Pubkey

		wg.Go(func() error {
			mp, err := minipool.NewMinipool(rp, address, nil)
			if err != nil {
				return fmt.Errorf("error making binding for minipool %s: %w", address.Hex(), err)
			}

			nodeDeposit, err := mp.GetNodeDepositBalance(nil)
			if err != nil {
				return fmt.Errorf("error getting node deposit for minipool %s: %w", address.Hex(), err)
			}
			nodeDeposits[i] = nodeDeposit

			userDeposit, err := mp.GetUserDepositBalance(nil)
			if err != nil {
				return fmt.Errorf("error getting user deposit for minipool %s: %w", address.Hex(), err)
			}
			userDeposits[i] = userDeposit

			reduceBondTime, err := minipool.GetReduceBondTime(rp, address, nil)
			if err != nil {
				return fmt.Errorf("error getting bond reduction time for minipool %s: %w", address.Hex(), err)
			}

			// Ignore minipools that don't have a bond reduction pending
			if reduceBondTime == time.Unix(0, 0) {
				pendingNodeDeposits[i] = nodeDeposit
				pendingUserDeposits[i] = userDeposit
				return nil
			}

			// Get the new (pending) bond
			newBond, err := minipool.GetReduceBondValue(rp, address, nil)
			if err != nil {
				return fmt.Errorf("error getting pending bond reduced balance for minipool %s: %w", address.Hex(), err)
			}
			pendingNodeDeposits[i] = newBond

			// New user deposit = old + delta
			pendingUserDeposits[i] = big.NewInt(0).Sub(nodeDeposit, newBond)
			pendingUserDeposits[i].Add(pendingUserDeposits[i], userDeposit)
			return nil
		})
	}

	// Wait for data
	if err = wg.Wait(); err != nil {
		return nil, nil, nil, nil, err
	}

	statuses, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error loading validator statuses: %w", err)
	}

	eligibleBorrowedEth := big.NewInt(0)
	eligibleBondedEth := big.NewInt(0)
	pendingEligibleBorrowedEth := big.NewInt(0)
	pendingEligibleBondedEth := big.NewInt(0)
	for i, pubkey := range pubkeys {
		status, exists := statuses[pubkey]
		if !exists {
			// Validator doesn't exist on Beacon yet
			continue
		}
		if status.ActivationEpoch > beaconHead.Epoch {
			// Validator hasn't activated yet
			continue
		}
		if status.ExitEpoch <= beaconHead.Epoch {
			// Validator exited
			continue
		}
		// It's eligible, so add up the borrowed and bonded amounts
		eligibleBorrowedEth.Add(eligibleBorrowedEth, userDeposits[i])
		eligibleBondedEth.Add(eligibleBondedEth, nodeDeposits[i])
		pendingEligibleBorrowedEth.Add(pendingEligibleBorrowedEth, pendingUserDeposits[i])
		pendingEligibleBondedEth.Add(pendingEligibleBondedEth, pendingNodeDeposits[i])
	}

	return eligibleBorrowedEth, eligibleBondedEth, pendingEligibleBorrowedEth, pendingEligibleBondedEth, nil

}
