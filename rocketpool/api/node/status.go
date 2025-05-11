package node

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	node131 "github.com/rocket-pool/rocketpool-go/legacy/v1.3.1/node"
	mp "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/alerting/alertmanager/models"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
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
	if err := services.RequireBeaconClientSynced(c); err != nil {
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
	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("Error getting the signer registry on network [%v].", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	saturnDeployed, err := state.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeStatusResponse{}
	response.IsSaturnDeployed = saturnDeployed
	response.PenalizedMinipools = map[common.Address]uint64{}
	response.NodeRPLLocked = big.NewInt(0)

	// Get the legacy MinipoolQueue contract address
	legacyMinipoolQueueAddress := cfg.Smartnode.GetV110MinipoolQueueAddress()

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address
	response.AccountAddressFormatted = formatResolvedAddress(c, response.AccountAddress)

	// Sync
	var wg errgroup.Group

	if saturnDeployed {
		wg.Go(func() error {
			deployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
			if err == nil {
				response.MegapoolDeployed = deployed
			}
			megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
			if err == nil {
				response.MegapoolAddress = megapoolAddress
			}

			// Load the megapool contract
			mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
			if err == nil {
				debt, err := mp.GetDebt(nil)
				if err == nil {
					response.MegapoolNodeDebt = debt
				}
				refund, err := mp.GetRefundValue(nil)
				if err == nil {
					response.MegapoolRefundValue = refund
				}
				validatorCount, err := mp.GetActiveValidatorCount(nil)
				if err == nil {
					response.MegapoolActiveValidatorCount = uint16(validatorCount)
				}
			}
			return err
		})

		wg.Go(func() error {
			expressTicketCount, err := node.GetExpressTicketCount(rp, nodeAccount.Address, nil)
			if err == nil {
				response.ExpressTicketCount = expressTicketCount
			}
			return err
		})

	}

	wg.Go(func() error {
		mpDetails, err := mp.GetNodeMinipoolDetails(rp, bc, nodeAccount.Address, &legacyMinipoolQueueAddress)
		if err == nil {
			response.Minipools = mpDetails
		}
		return err
	})

	wg.Go(func() error {
		delegate, err := rp.GetContract("rocketMinipoolDelegate", nil)
		if err != nil {
			return fmt.Errorf("Error getting latest minipool delegate contract: %w", err)
		}
		response.LatestDelegate = *delegate.Address
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

	// Get node details
	wg.Go(func() error {
		details, err := node.GetNodeDetails(rp, nodeAccount.Address, true, nil)
		if err == nil {
			response.Registered = details.Exists
			response.PrimaryWithdrawalAddress = details.PrimaryWithdrawalAddress
			response.PrimaryWithdrawalAddressFormatted = formatResolvedAddress(c, response.PrimaryWithdrawalAddress)
			response.PendingPrimaryWithdrawalAddress = details.PendingPrimaryWithdrawalAddress
			response.PendingPrimaryWithdrawalAddressFormatted = formatResolvedAddress(c, response.PendingPrimaryWithdrawalAddress)
			response.IsRPLWithdrawalAddressSet = details.IsRPLWithdrawalAddressSet
			response.RPLWithdrawalAddress = details.RPLWithdrawalAddress
			response.RPLWithdrawalAddressFormatted = formatResolvedAddress(c, response.RPLWithdrawalAddress)
			response.PendingRPLWithdrawalAddress = details.PendingRPLWithdrawalAddress
			response.PendingRPLWithdrawalAddressFormatted = formatResolvedAddress(c, response.PendingRPLWithdrawalAddress)
			response.TimezoneLocation = details.TimezoneLocation
		}
		return err
	})

	// Check whether RPL locking is allowed for the node
	wg.Go(func() error {
		var err error
		response.IsRPLLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
		return err
	})

	if saturnDeployed {
		// Get the node's locked RPL
		wg.Go(func() error {
			var err error
			response.NodeRPLLocked, err = node.GetNodeLockedRPL(rp, nodeAccount.Address, nil)
			return err
		})
		// Get staking details
		wg.Go(func() error {
			var err error
			response.RplStake, err = node.GetNodeStakedRPL(rp, nodeAccount.Address, nil)
			return err
		})
		wg.Go(func() error {
			var err error
			response.RplStakeMegapool, err = node.GetNodeMegapoolStakedRPL(rp, nodeAccount.Address, nil)
			return err
		})
		wg.Go(func() error {
			var err error
			response.RplStakeLegacy, err = node.GetNodeLegacyStakedRPL(rp, nodeAccount.Address, nil)
			return err
		})
		wg.Go(func() error {
			var err error
			response.UnstakingRPL, err = node.GetNodeUnstakingRPL(rp, nodeAccount.Address, nil)
			return err
		})
		wg.Go(func() error {
			var err error
			unstakingPeriod, err := protocol.GetNodeUnstakingPeriod(rp, nil)
			if err != nil {
				response.UnstakingPeriodDuration = time.Duration(unstakingPeriod.Int64()) * time.Second
			}
			return err
		})
		wg.Go(func() error {
			var err error
			lastUnstakeTimestamp, err := node.GetNodeLastUnstakeTime(rp, nodeAccount.Address, nil)
			if err != nil {
				// Convert the lastUnstakeTimestamp to a time.Time object
				response.LastRPLUnstakeTime = time.Unix(int64(lastUnstakeTimestamp), 0)
			}
			return err
		})

		wg.Go(func() error {
			var err error
			response.MaximumRplStake, err = node.GetNodeMaximumRPLStakeForMinipools(rp, nodeAccount.Address, nil)
			return err
		})
	} else {
		// Get the node's locked RPL
		wg.Go(func() error {
			var err error
			response.NodeRPLLocked, err = node131.GetNodeRPLLocked(rp, nodeAccount.Address, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.MaximumRplStake, err = node131.GetNodeMaximumRPLStake(rp, nodeAccount.Address, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.MinimumRplStake, err = node131.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil)
			return err
		})
	}

	// Check if Voting is Initialized
	wg.Go(func() error {
		var err error
		response.IsVotingInitialized, err = network.GetVotingInitialized(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the node onchain voting delegate
	wg.Go(func() error {
		var err error
		response.OnchainVotingDelegate, err = network.GetCurrentVotingDelegate(rp, nodeAccount.Address, nil)
		if err == nil {
			response.OnchainVotingDelegateFormatted = formatResolvedAddress(c, response.OnchainVotingDelegate)
		}
		return err
	})

	// Get node account balances
	wg.Go(func() error {
		var err error
		response.AccountBalances, err = tokens.GetBalances(rp, nodeAccount.Address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.MaximumStakeFraction, err = protocol.GetMaximumPerMinipoolStake(rp, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		response.EthBorrowed, response.EthBorrowedLimit, response.PendingBorrowAmount, err = rputils.CheckCollateral(saturnDeployed, rp, nodeAccount.Address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.CreditBalance, err = node.GetNodeDepositCredit(rp, nodeAccount.Address, nil)
		return err
	})

	// Get active and past votes from Snapshot, but treat errors as non-Fatal
	if reg != nil {
		wg.Go(func() error {
			var err error
			r := &response.SnapshotResponse
			if cfg.Smartnode.GetRocketSignerRegistryAddress() != "" {
				response.SignallingAddress, err = reg.NodeToSigner(&bind.CallOpts{}, nodeAccount.Address)
				if err != nil {
					r.Error = err.Error()
					return nil
				}
				blankAddress := common.Address{}
				if response.SignallingAddress != blankAddress {
					response.SignallingAddressFormatted = formatResolvedAddress(c, response.SignallingAddress)
				}
				votedProposals, err := pdao.GetSnapshotVotedProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), nodeAccount.Address, response.SignallingAddress)
				if err != nil {
					r.Error = err.Error()
					return nil
				}
				r.ProposalVotes = votedProposals.Data.Votes
			}
			snapshotResponse, err := pdao.GetSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
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
		feeRecipientInfo, err := rputils.GetFeeRecipientInfoWithoutState(rp, bc, nodeAccount.Address, nil)
		if err == nil {
			response.FeeRecipientInfo = *feeRecipientInfo
			response.FeeDistributorBalance, err = rp.Client.BalanceAt(context.Background(), feeRecipientInfo.FeeDistributorAddress, nil)
		}
		return err
	})

	// Get alerts from Alertmanager
	wg.Go(func() error {
		alerts, err := alerting.FetchAlerts(cfg)
		if err != nil {
			// no reason to make `rocketpool node status` fail if we can't get alerts
			// (this is more likely to happen in native mode than docker where
			// alertmanager is more complex to set up)
			// Do save a warning though to print to the user
			response.Warning = fmt.Sprintf("Error fetching alerts from Alertmanager: %s", err)
			alerts = make([]*models.GettableAlert, 0)
		}
		response.Alerts = make([]api.NodeAlert, len(alerts))

		for i, a := range alerts {
			response.Alerts[i] = api.NodeAlert{
				State:       *a.Status.State,
				Labels:      a.Labels,
				Annotations: a.Annotations,
			}
		}
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Get withdrawal address balances
	if !bytes.Equal(nodeAccount.Address.Bytes(), response.PrimaryWithdrawalAddress.Bytes()) {
		withdrawalBalances, err := tokens.GetBalances(rp, response.PrimaryWithdrawalAddress, nil)
		if err != nil {
			return nil, err
		}
		response.PrimaryWithdrawalBalances = withdrawalBalances
	}
	if !bytes.Equal(nodeAccount.Address.Bytes(), response.RPLWithdrawalAddress.Bytes()) &&
		!bytes.Equal(response.PrimaryWithdrawalAddress.Bytes(), response.RPLWithdrawalAddress.Bytes()) {
		withdrawalBalances, err := tokens.GetBalances(rp, response.RPLWithdrawalAddress, nil)
		if err != nil {
			return nil, err
		}
		response.RPLWithdrawalBalances = withdrawalBalances
	}

	creditAndBalance, err := node.GetNodeCreditAndBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.CreditAndEthOnBehalfBalance = creditAndBalance
	usableCreditAndBalance, err := node.GetNodeUsableCreditAndBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.UsableCreditAndEthOnBehalfBalance = usableCreditAndBalance
	ethBalance, err := node.GetNodeEthBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.EthOnBehalfBalance = ethBalance

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
		trueMinimumStake := big.NewInt(0).Add(response.EthBorrowed, response.PendingBorrowAmount)
		trueMinimumStake.Mul(trueMinimumStake, minStakeFraction)
		trueMinimumStake.Div(trueMinimumStake, rplPrice)

		// Calculate the *real* maximum, including the pending bond reductions
		trueMaximumStake := eth.EthToWei(32)
		trueMaximumStake.Mul(trueMaximumStake, big.NewInt(int64(activeMinipools)))
		trueMaximumStake.Sub(trueMaximumStake, response.EthBorrowed)
		trueMaximumStake.Sub(trueMaximumStake, response.PendingBorrowAmount) // (32 * activeMinipools - ethBorrowed - pendingBorrow)
		trueMaximumStake.Mul(trueMaximumStake, maxStakeFraction)
		trueMaximumStake.Div(trueMaximumStake, rplPrice)

		response.MinimumRplStake = trueMinimumStake
		response.MaximumRplStake = trueMaximumStake

		if response.EffectiveRplStake.Cmp(trueMinimumStake) < 0 {
			response.EffectiveRplStake.SetUint64(0)
		} else if response.EffectiveRplStake.Cmp(trueMaximumStake) > 0 {
			response.EffectiveRplStake.Set(trueMaximumStake)
		}

		response.BondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / (float64(activeMinipools)*32.0 - eth.WeiToEth(response.EthBorrowed) - eth.WeiToEth(response.PendingBorrowAmount))
		response.BorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / (eth.WeiToEth(response.EthBorrowed) + eth.WeiToEth(response.PendingBorrowAmount))

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

		response.PendingMinimumRplStake = pendingTrueMinimumStake
		response.PendingMaximumRplStake = pendingTrueMaximumStake

		response.PendingEffectiveRplStake = big.NewInt(0).Set(response.RplStake)
		if response.PendingEffectiveRplStake.Cmp(pendingTrueMinimumStake) < 0 {
			response.PendingEffectiveRplStake.SetUint64(0)
		} else if response.PendingEffectiveRplStake.Cmp(pendingTrueMaximumStake) > 0 {
			response.PendingEffectiveRplStake.Set(pendingTrueMaximumStake)
		}

		pendingEligibleBondedEthFloat := eth.WeiToEth(pendingEligibleBondedEth)
		if pendingEligibleBondedEthFloat == 0 {
			response.PendingBondedCollateralRatio = 0
		} else {
			response.PendingBondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / pendingEligibleBondedEthFloat
		}

		pendingEligibleBorrowedEthFloat := eth.WeiToEth(pendingEligibleBorrowedEth)
		if pendingEligibleBorrowedEthFloat == 0 {
			response.PendingBorrowedCollateralRatio = 0
		} else {
			response.PendingBorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(response.RplStake) / pendingEligibleBorrowedEthFloat
		}
	} else {
		response.BorrowedCollateralRatio = -1
		response.BondedCollateralRatio = -1
		response.PendingEffectiveRplStake = big.NewInt(0)
		response.PendingMinimumRplStake = big.NewInt(0)
		response.PendingMaximumRplStake = big.NewInt(0)
		response.PendingBondedCollateralRatio = -1
		response.PendingBorrowedCollateralRatio = -1
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

	latestBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting latest block header: %w", err)
	}
	blockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	var reductionWindowStart uint64
	var reductionWindowLength uint64

	// Data
	var wg1 errgroup.Group

	wg1.Go(func() error {
		var err error
		reductionWindowStart, err = tnsettings.GetBondReductionWindowStart(rp, nil)
		return err
	})
	wg1.Go(func() error {
		var err error
		reductionWindowLength, err = tnsettings.GetBondReductionWindowLength(rp, nil)
		return err
	})

	// Wait for data
	if err = wg1.Wait(); err != nil {
		return nil, nil, nil, nil, err
	}

	reductionWindowEnd := time.Duration(reductionWindowStart+reductionWindowLength) * time.Second

	// Data
	var wg errgroup.Group
	zeroTime := time.Unix(0, 0)

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

			reduceBondCancelled, err := minipool.GetReduceBondCancelled(rp, address, nil)
			if err != nil {
				return fmt.Errorf("error getting bond reduction cancel status for minipool %s: %w", address.Hex(), err)
			}

			// Ignore minipools that don't have a bond reduction pending
			timeSinceReductionStart := blockTime.Sub(reduceBondTime)
			if reduceBondTime == zeroTime ||
				reduceBondCancelled ||
				timeSinceReductionStart > reductionWindowEnd {
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
