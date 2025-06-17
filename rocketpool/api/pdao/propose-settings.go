package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	node131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/node"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func canProposeSetting(c *cli.Context, contractName string, settingName string, value string) (*api.CanProposePDAOSettingResponse, error) {

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

	// Check if Saturn is already deployed
	saturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanProposePDAOSettingResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Sync
	var stakedRpl *big.Int
	var lockedRpl *big.Int
	var proposalBond *big.Int
	var isRplLockingAllowed bool
	var wg errgroup.Group

	if saturnDeployed {
		// Get the node's RPL stake
		wg.Go(func() error {
			var err error
			stakedRpl, err = node.GetNodeStakedRPL(rp, nodeAccount.Address, nil)
			return err
		})

		// Get the node's locked RPL
		wg.Go(func() error {
			var err error
			lockedRpl, err = node.GetNodeLockedRPL(rp, nodeAccount.Address, nil)
			return err
		})

	} else {
		// Get the node's RPL stake
		wg.Go(func() error {
			var err error
			stakedRpl, err = node131.GetNodeRPLStake(rp, nodeAccount.Address, nil)
			return err
		})

		// Get the node's locked RPL
		wg.Go(func() error {
			var err error
			lockedRpl, err = node131.GetNodeRPLLocked(rp, nodeAccount.Address, nil)
			return err
		})
	}

	// Get the node's RPL stake
	wg.Go(func() error {
		var err error
		proposalBond, err = protocol.GetProposalBond(rp, nil)
		return err
	})

	// Get is RPL locking allowed
	wg.Go(func() error {
		var err error
		isRplLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	response.StakedRpl = stakedRpl
	response.LockedRpl = lockedRpl
	response.ProposalBond = proposalBond
	response.IsRplLockingDisallowed = !isRplLockingAllowed

	freeRpl := big.NewInt(0).Sub(stakedRpl, lockedRpl)
	response.InsufficientRpl = (freeRpl.Cmp(proposalBond) < 0)

	// return if proposing is not possible
	response.CanPropose = !(response.InsufficientRpl || response.IsRplLockingDisallowed)
	if !response.CanPropose {
		return &response, nil
	}

	// Get the latest finalized block number and corresponding pollard
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, fmt.Errorf("error creating pollard: %w", err)
	}
	response.BlockNumber = blockNumber

	// Get the account transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Estimate the gas
	valueName := "value"
	switch contractName {
	case protocol.AuctionSettingsContractName:
		switch settingName {
		// CreateLotEnabled
		case protocol.CreateLotEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeCreateLotEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing CreateLotEnabled: %w", err)
			}

		// BidOnLotEnabled
		case protocol.BidOnLotEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeBidOnLotEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing BidOnLotEnabled: %w", err)
			}

		// LotMinimumEthValue
		case protocol.LotMinimumEthValueSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeLotMinimumEthValueGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing LotMinimumEthValue: %w", err)
			}

		// LotMaximumEthValue
		case protocol.LotMaximumEthValueSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeLotMaximumEthValueGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing LotMaximumEthValue: %w", err)
			}

		// LotDuration
		case protocol.LotDurationSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeLotDurationGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing LotDuration: %w", err)
			}

		// LotStartingPriceRatio
		case protocol.LotStartingPriceRatioSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeLotStartingPriceRatioGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing LotStartingPriceRatio: %w", err)
			}

		// LotReservePriceRatio
		case protocol.LotReservePriceRatioSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeLotReservePriceRatioGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing LotReservePriceRatio: %w", err)
			}
		}

	case protocol.DepositSettingsContractName:
		switch settingName {
		// DepositEnabled
		case protocol.DepositEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeDepositEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing DepositEnabled: %w", err)
			}

		// AssignDepositsEnabled
		case protocol.AssignDepositsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeAssignDepositsEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing AssignDepositsEnabled: %w", err)
			}

		// MinimumDeposit
		case protocol.MinimumDepositSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinimumDepositGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinimumDeposit: %w", err)
			}

		// MaximumDepositPoolSize
		case protocol.MaximumDepositPoolSizeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMaximumDepositPoolSizeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaximumDepositPoolSize: %w", err)
			}

		// MaximumDepositAssignments
		case protocol.MaximumDepositAssignmentsSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMaximumDepositAssignmentsGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaximumDepositAssignments: %w", err)
			}

		// MaximumSocializedDepositAssignments
		case protocol.MaximumSocializedDepositAssignmentsSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMaximumSocializedDepositAssignmentsGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaximumSocializedDepositAssignments: %w", err)
			}

		// DepositFee
		case protocol.DepositFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeDepositFeeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing DepositFee: %w", err)
			}
		}

	case protocol.MinipoolSettingsContractName:
		switch settingName {
		// MinipoolSubmitWithdrawableEnabled
		case protocol.MinipoolSubmitWithdrawableEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinipoolSubmitWithdrawableEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinipoolSubmitWithdrawableEnabled: %w", err)
			}

		// MinipoolLaunchTimeout
		case protocol.MinipoolLaunchTimeoutSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinipoolLaunchTimeoutGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinipoolLaunchTimeout: %w", err)
			}

		// BondReductionEnabled
		case protocol.BondReductionEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeBondReductionEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing BondReductionEnabled: %w", err)
			}

		// MaximumMinipoolCount
		case protocol.MaximumMinipoolCountSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMaximumMinipoolCountGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaximumMinipoolCount: %w", err)
			}

		// MinipoolUserDistributeWindowStart
		case protocol.MinipoolUserDistributeWindowStartSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinipoolUserDistributeWindowStartGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinipoolUserDistributeWindowStart: %w", err)
			}

		// MinipoolUserDistributeWindowLength
		case protocol.MinipoolUserDistributeWindowLengthSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinipoolUserDistributeWindowLengthGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinipoolUserDistributeWindowLength: %w", err)
			}
		}

	case protocol.NetworkSettingsContractName:
		switch settingName {
		// NodeConsensusThreshold
		case protocol.NodeConsensusThresholdSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeConsensusThresholdGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeConsensusThreshold: %w", err)
			}

		// SubmitBalancesEnabled
		case protocol.SubmitBalancesEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSubmitBalancesEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitBalancesEnabled: %w", err)
			}

		// SubmitBalancesFrequency
		case protocol.SubmitBalancesFrequencySettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSubmitBalancesFrequencyGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitBalancesFrequency: %w", err)
			}

		// SubmitPricesEnabled
		case protocol.SubmitPricesEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSubmitPricesEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitPricesEnabled: %w", err)
			}

		// SubmitPricesFrequency
		case protocol.SubmitPricesFrequencySettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSubmitPricesFrequencyGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitPricesFrequency: %w", err)
			}

		// MinimumNodeFee
		case protocol.MinimumNodeFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinimumNodeFeeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinimumNodeFee: %w", err)
			}

		// TargetNodeFee
		case protocol.TargetNodeFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeTargetNodeFeeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing TargetNodeFee: %w", err)
			}

		// MaximumNodeFee
		case protocol.MaximumNodeFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMaximumNodeFeeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaximumNodeFee: %w", err)
			}

		// NodeFeeDemandRange
		case protocol.NodeFeeDemandRangeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeFeeDemandRangeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeFeeDemandRange: %w", err)
			}

		// TargetRethCollateralRate
		case protocol.TargetRethCollateralRateSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeTargetRethCollateralRateGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing TargetRethCollateralRate: %w", err)
			}

		// NetworkPenaltyThreshold
		case protocol.NetworkPenaltyThresholdSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNetworkPenaltyThresholdGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NetworkPenaltyThreshold: %w", err)
			}

		// NetworkPenaltyPerRate
		case protocol.NetworkPenaltyPerRateSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNetworkPenaltyPerRateGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NetworkPenaltyPerRate: %w", err)
			}

		// SubmitRewardsEnabled
		case protocol.SubmitRewardsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSubmitRewardsEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitRewardsEnabled: %w", err)
			}

		// NodeShare
		case protocol.NetworkNodeCommissionSharePath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeShareGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeShare: %w", err)
			}
		// NodeShareSecurityCouncilAdder
		case protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeShareSecurityCouncilAdderGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeShareSecurityCouncilAdder: %w", err)
			}
		// VoterShare
		case protocol.NetworkVoterSharePath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeVoterShareGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing VoterShare: %w", err)
			}
		//MaxNodeShareSecurityCouncilAdder
		case protocol.NetworkMaxNodeShareSecurityCouncilAdderPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateMaxNodeShareSecurityCouncilAdder(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaxNodeShareSecurityCouncilAdder: %w", err)
			}

		}

	case protocol.NodeSettingsContractName:
		switch settingName {
		// NodeRegistrationEnabled
		case protocol.NodeRegistrationEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeRegistrationEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeRegistrationEnabled: %w", err)
			}

		// SmoothingPoolRegistrationEnabled
		case protocol.SmoothingPoolRegistrationEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSmoothingPoolRegistrationEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SmoothingPoolRegistrationEnabled: %w", err)
			}

		// NodeDepositEnabled
		case protocol.NodeDepositEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeDepositEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeDepositEnabled: %w", err)
			}

		// VacantMinipoolsEnabled
		case protocol.VacantMinipoolsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeVacantMinipoolsEnabledGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing VacantMinipoolsEnabled: %w", err)
			}

		// MinimumPerMinipoolStake
		case protocol.MinimumPerMinipoolStakeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMinimumPerMinipoolStakeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinimumPerMinipoolStake: %w", err)
			}

		// MaximumPerMinipoolStake
		case protocol.MaximumPerMinipoolStakeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMaximumPerMinipoolStakeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MaximumPerMinipoolStake: %w", err)
			}
		// ReducedBond
		case protocol.ReducedBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeReducedBond(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ReducedBond: %w", err)
			}
		// NodeUnstakingPeriod
		case protocol.NodeUnstakingPeriodSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeNodeUnstakingPeriod(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeUnstakingPeriod: %w", err)
			}

		}

	case protocol.ProposalsSettingsContractName:
		switch settingName {
		// VotePhase1Time
		case protocol.VotePhase1TimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeVotePhase1TimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing VoteTime: %w", err)
			}
		// VoteTimePhase2
		case protocol.VotePhase2TimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeVotePhase2TimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing VoteTime: %w", err)
			}
		// VoteDelayTime
		case protocol.VoteDelayTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeVoteDelayTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing VoteDelayTime: %w", err)
			}

		// ExecuteTime
		case protocol.ExecuteTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeExecuteTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ExecuteTime: %w", err)
			}

		// ProposalBond
		case protocol.ProposalBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeProposalBondGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ProposalBond: %w", err)
			}

		// ChallengeBond
		case protocol.ChallengeBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeChallengeBondGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ChallengeBond: %w", err)
			}

		// ChallengePeriod
		case protocol.ChallengePeriodSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeChallengePeriodGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ChallengePeriod: %w", err)
			}

		// ProposalQuorum
		case protocol.ProposalQuorumSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeProposalQuorumGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ProposalQuorum: %w", err)
			}

		// ProposalVetoQuorum
		case protocol.ProposalVetoQuorumSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeProposalVetoQuorumGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ProposalVetoQuorum: %w", err)
			}

		// ProposalMaxBlockAge
		case protocol.ProposalMaxBlockAgeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeProposalMaxBlockAgeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing ProposalMaxBlockAge: %w", err)
			}
		}

	case protocol.RewardsSettingsContractName:
		switch settingName {
		// RewardsClaimIntervalTime
		case protocol.RewardsClaimIntervalPeriodsSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeRewardsClaimIntervalTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing RewardsClaimIntervalTime: %w", err)
			}
		}

	case protocol.SecuritySettingsContractName:
		switch settingName {
		// SecurityMembersQuorum
		case protocol.SecurityMembersQuorumSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSecurityMembersQuorumGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SecurityMembersQuorum: %w", err)
			}

		// SecurityMembersLeaveTime
		case protocol.SecurityMembersLeaveTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSecurityMembersLeaveTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SecurityMembersLeaveTime: %w", err)
			}

		// SecurityProposalVoteTime
		case protocol.SecurityProposalVoteTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSecurityProposalVoteTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SecurityProposalVoteTime: %w", err)
			}

		// SecurityProposalExecuteTime
		case protocol.SecurityProposalExecuteTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSecurityProposalExecuteTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SecurityProposalExecuteTime: %w", err)
			}

		// SecurityProposalActionTime
		case protocol.SecurityProposalActionTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeSecurityProposalActionTimeGas(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SecurityProposalActionTime: %w", err)
			}
		}

	case protocol.MegapoolSettingsContractName:
		switch settingName {
		// TimeBeforeDissolve
		case protocol.MegapoolTimeBeforeDissolveSettingsPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = protocol.EstimateProposeMegapoolTimeBeforeDissolve(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing TimeBeforeDissolve: %w", err)
			}
		}
	// MaximumEthPenalty
	case protocol.ReducedBondSettingPath:
		newValue, err := cliutils.ValidateBigInt(valueName, value)
		if err != nil {
			return nil, err
		}
		response.GasInfo, err = protocol.EstimateProposeMaximumEthPenalty(rp, newValue, blockNumber, pollard, opts)
		if err != nil {
			return nil, fmt.Errorf("error estimating gas for proposing ReducedBond: %w", err)
		}

	}

	// Make sure a setting was actually hit
	blankGasInfo := rocketpool.GasInfo{}
	if response.GasInfo == blankGasInfo {
		return nil, fmt.Errorf("[%s - %s] is not a valid PDAO contract and setting name combo", contractName, settingName)
	}

	// Update & return response
	return &response, nil

}

func proposeSetting(c *cli.Context, contractName string, settingName string, value string, blockNumber uint32) (*api.ProposePDAOSettingResponse, error) {

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

	// Response
	response := api.ProposePDAOSettingResponse{}

	// Decode the pollard
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error regenerating pollard: %w", err)
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit the proposal
	var proposalID uint64
	var hash common.Hash
	valueName := "value"
	switch contractName {
	case protocol.AuctionSettingsContractName:
		switch settingName {
		// CreateLotEnabled
		case protocol.CreateLotEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeCreateLotEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing CreateLotEnabled: %w", err)
			}

		// BidOnLotEnabled
		case protocol.BidOnLotEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeBidOnLotEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing BidOnLotEnabled: %w", err)
			}

		// LotMinimumEthValue
		case protocol.LotMinimumEthValueSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeLotMinimumEthValue(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing LotMinimumEthValue: %w", err)
			}

		// LotMaximumEthValue
		case protocol.LotMaximumEthValueSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeLotMaximumEthValue(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing LotMaximumEthValue: %w", err)
			}

		// LotDuration
		case protocol.LotDurationSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeLotDuration(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing LotDuration: %w", err)
			}

		// LotStartingPriceRatio
		case protocol.LotStartingPriceRatioSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeLotStartingPriceRatio(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing LotStartingPriceRatio: %w", err)
			}

		// LotReservePriceRatio
		case protocol.LotReservePriceRatioSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeLotReservePriceRatio(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing LotReservePriceRatio: %w", err)
			}
		}

	case protocol.DepositSettingsContractName:
		switch settingName {
		// DepositEnabled
		case protocol.DepositEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeDepositEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing DepositEnabled: %w", err)
			}

		// AssignDepositsEnabled
		case protocol.AssignDepositsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeAssignDepositsEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing AssignDepositsEnabled: %w", err)
			}

		// MinimumDeposit
		case protocol.MinimumDepositSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinimumDeposit(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinimumDeposit: %w", err)
			}

		// MaximumDepositPoolSize
		case protocol.MaximumDepositPoolSizeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumDepositPoolSize(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumDepositPoolSize: %w", err)
			}

		// MaximumDepositAssignments
		case protocol.MaximumDepositAssignmentsSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumDepositAssignments(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumDepositAssignments: %w", err)
			}

		// MaximumSocializedDepositAssignments
		case protocol.MaximumSocializedDepositAssignmentsSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumSocializedDepositAssignments(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumSocializedDepositAssignments: %w", err)
			}

		// DepositFee
		case protocol.DepositFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeDepositFee(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing DepositFee: %w", err)
			}
		}

	case protocol.MinipoolSettingsContractName:
		switch settingName {
		// MinipoolSubmitWithdrawableEnabled
		case protocol.MinipoolSubmitWithdrawableEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinipoolSubmitWithdrawableEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinipoolSubmitWithdrawableEnabled: %w", err)
			}

		// MinipoolLaunchTimeout
		case protocol.MinipoolLaunchTimeoutSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinipoolLaunchTimeout(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinipoolLaunchTimeout: %w", err)
			}

		// BondReductionEnabled
		case protocol.BondReductionEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeBondReductionEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing BondReductionEnabled: %w", err)
			}

		// MaximumMinipoolCount
		case protocol.MaximumMinipoolCountSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumMinipoolCount(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumMinipoolCount: %w", err)
			}

		// MinipoolUserDistributeWindowStart
		case protocol.MinipoolUserDistributeWindowStartSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinipoolUserDistributeWindowStart(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinipoolUserDistributeWindowStart: %w", err)
			}

		// MinipoolUserDistributeWindowLength
		case protocol.MinipoolUserDistributeWindowLengthSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinipoolUserDistributeWindowLength(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinipoolUserDistributeWindowLength: %w", err)
			}
		}

	case protocol.NetworkSettingsContractName:
		switch settingName {
		// NodeConsensusThreshold
		case protocol.NodeConsensusThresholdSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeConsensusThreshold(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeConsensusThreshold: %w", err)
			}

		// SubmitBalancesEnabled
		case protocol.SubmitBalancesEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSubmitBalancesEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitBalancesEnabled: %w", err)
			}

		// SubmitBalancesFrequency
		case protocol.SubmitBalancesFrequencySettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSubmitBalancesFrequency(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitBalancesFrequency: %w", err)
			}

		// SubmitPricesEnabled
		case protocol.SubmitPricesEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSubmitPricesEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitPricesEnabled: %w", err)
			}

		// SubmitPricesFrequency
		case protocol.SubmitPricesFrequencySettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSubmitPricesFrequency(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitPricesFrequency: %w", err)
			}

		// MinimumNodeFee
		case protocol.MinimumNodeFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinimumNodeFee(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinimumNodeFee: %w", err)
			}

		// TargetNodeFee
		case protocol.TargetNodeFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeTargetNodeFee(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing TargetNodeFee: %w", err)
			}

		// MaximumNodeFee
		case protocol.MaximumNodeFeeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumNodeFee(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumNodeFee: %w", err)
			}

		// NodeFeeDemandRange
		case protocol.NodeFeeDemandRangeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeFeeDemandRange(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeFeeDemandRange: %w", err)
			}

		// TargetRethCollateralRate
		case protocol.TargetRethCollateralRateSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeTargetRethCollateralRate(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing TargetRethCollateralRate: %w", err)
			}

		// NetworkPenaltyThreshold
		case protocol.NetworkPenaltyThresholdSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNetworkPenaltyThreshold(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NetworkPenaltyThreshold: %w", err)
			}

		// NetworkPenaltyPerRate
		case protocol.NetworkPenaltyPerRateSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNetworkPenaltyPerRate(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NetworkPenaltyPerRate: %w", err)
			}

		// SubmitRewardsEnabled
		case protocol.SubmitRewardsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSubmitRewardsEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitRewardsEnabled: %w", err)
			}
		// NodeShare
		case protocol.NetworkNodeCommissionSharePath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeShare(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeShare: %w", err)
			}
		// NodeShareSecurityCouncilAdder
		case protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeShareSecurityCouncilAdder(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeShareSecurityCouncilAdder: %w", err)
			}
		// VoterShare
		case protocol.NetworkVoterSharePath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeVoterShare(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing VoterShare: %w", err)
			}
		// MaxNodeShareSecurityCouncilAdder
		case protocol.NetworkMaxNodeShareSecurityCouncilAdderPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaxNodeShareSecurityCouncilAdder(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaxNodeShareSecurityCouncilAdder: %w", err)
			}

		}

	case protocol.NodeSettingsContractName:
		switch settingName {
		// NodeRegistrationEnabled
		case protocol.NodeRegistrationEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeRegistrationEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeRegistrationEnabled: %w", err)
			}

		// SmoothingPoolRegistrationEnabled
		case protocol.SmoothingPoolRegistrationEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSmoothingPoolRegistrationEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SmoothingPoolRegistrationEnabled: %w", err)
			}

		// NodeDepositEnabled
		case protocol.NodeDepositEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeDepositEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeDepositEnabled: %w", err)
			}

		// VacantMinipoolsEnabled
		case protocol.VacantMinipoolsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeVacantMinipoolsEnabled(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing VacantMinipoolsEnabled: %w", err)
			}

		// MinimumPerMinipoolStake
		case protocol.MinimumPerMinipoolStakeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMinimumPerMinipoolStake(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinimumPerMinipoolStake: %w", err)
			}

		// MaximumPerMinipoolStake
		case protocol.MaximumPerMinipoolStakeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumPerMinipoolStake(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumPerMinipoolStake: %w", err)
			}
		// ReducedBond
		case protocol.ReducedBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeReducedBond(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ReduceBond: %w", err)
			}
		// NodeUnstakingPeriod
		case protocol.NodeUnstakingPeriodSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeNodeUnstakingPeriod(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeUnstakingPeriod: %w", err)
			}

		}

	case protocol.ProposalsSettingsContractName:
		switch settingName {
		// VotePhase1Time
		case protocol.VotePhase1TimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeVotePhase1Time(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing VoteTime: %w", err)
			}

		// VotePhase2Time
		case protocol.VotePhase2TimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeVotePhase2Time(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing VoteTime: %w", err)
			}

		// VoteDelayTime
		case protocol.VoteDelayTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeVoteDelayTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing VoteDelayTime: %w", err)
			}

		// ExecuteTime
		case protocol.ExecuteTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeExecuteTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ExecuteTime: %w", err)
			}

		// ProposalBond
		case protocol.ProposalBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeProposalBond(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ProposalBond: %w", err)
			}

		// ChallengeBond
		case protocol.ChallengeBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeChallengeBond(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ChallengeBond: %w", err)
			}

		// ChallengePeriod
		case protocol.ChallengePeriodSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeChallengePeriod(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ChallengePeriod: %w", err)
			}

		// ProposalQuorum
		case protocol.ProposalQuorumSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeProposalQuorum(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ProposalQuorum: %w", err)
			}

		// ProposalVetoQuorum
		case protocol.ProposalVetoQuorumSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeProposalVetoQuorum(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ProposalVetoQuorum: %w", err)
			}

		// ProposalMaxBlockAge
		case protocol.ProposalMaxBlockAgeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeProposalMaxBlockAge(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing ProposalMaxBlockAge: %w", err)
			}
		}

	case protocol.RewardsSettingsContractName:
		switch settingName {
		// RewardsClaimIntervalTime
		case protocol.RewardsClaimIntervalPeriodsSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeRewardsClaimIntervalTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing RewardsClaimIntervalTime: %w", err)
			}
		}

	case protocol.SecuritySettingsContractName:
		switch settingName {
		// SecurityMembersQuorum
		case protocol.SecurityMembersQuorumSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSecurityMembersQuorum(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SecurityMembersQuorum: %w", err)
			}

		// SecurityMembersLeaveTime
		case protocol.SecurityMembersLeaveTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSecurityMembersLeaveTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SecurityMembersLeaveTime: %w", err)
			}

		// SecurityProposalVoteTime
		case protocol.SecurityProposalVoteTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSecurityProposalVoteTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SecurityProposalVoteTime: %w", err)
			}

		// SecurityProposalExecuteTime
		case protocol.SecurityProposalExecuteTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSecurityProposalExecuteTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SecurityProposalExecuteTime: %w", err)
			}

		// SecurityProposalActionTime
		case protocol.SecurityProposalActionTimeSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeSecurityProposalActionTime(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SecurityProposalActionTime: %w", err)
			}
		}

	case protocol.MegapoolSettingsContractName:
		switch settingName {
		// TimeBeforeDissolve
		case protocol.MegapoolTimeBeforeDissolveSettingsPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMegapoolTimeBeforeDissolve(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing TimeBeforeDissolve: %w", err)
			}
		// MaximumEthPenalty
		case protocol.ReducedBondSettingPath:
			newValue, err := cliutils.ValidateBigInt(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = protocol.ProposeMaximumEthPenalty(rp, newValue, blockNumber, pollard, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MaximumEthPenalty: %w", err)
			}
		}
	}

	// Make sure a setting was actually hit
	blankHash := common.Hash{}
	if hash == blankHash {
		return nil, fmt.Errorf("[%s - %s] is not a valid PDAO contract and setting name combo", contractName, settingName)
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}
