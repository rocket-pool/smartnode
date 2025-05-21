package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/security"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canProposeSetting(c *cli.Context, contractName string, settingName string, value string) (*api.SecurityCanProposeSettingResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.SecurityCanProposeSettingResponse{}

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
			response.GasInfo, err = security.EstimateProposeCreateLotEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing CreateLotEnabled: %w", err)
			}

		// BidOnLotEnabled
		case protocol.BidOnLotEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeBidOnLotEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing BidOnLotEnabled: %w", err)
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
			response.GasInfo, err = security.EstimateProposeDepositEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing DepositEnabled: %w", err)
			}

		// AssignDepositsEnabled
		case protocol.AssignDepositsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeAssignDepositsEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing AssignDepositsEnabled: %w", err)
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
			response.GasInfo, err = security.EstimateProposeMinipoolSubmitWithdrawableEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing MinipoolSubmitWithdrawableEnabled: %w", err)
			}

		// BondReductionEnabled
		case protocol.BondReductionEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeBondReductionEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing BondReductionEnabled: %w", err)
			}
		}

	case protocol.NetworkSettingsContractName:
		switch settingName {
		// SubmitBalancesEnabled
		case protocol.SubmitBalancesEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeSubmitBalancesEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitBalancesEnabled: %w", err)
			}

		// SubmitRewardsEnabled
		case protocol.SubmitRewardsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeSubmitRewardsEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SubmitRewardsEnabled: %w", err)
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
			response.GasInfo, err = security.EstimateProposeNodeRegistrationEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeRegistrationEnabled: %w", err)
			}

		// SmoothingPoolRegistrationEnabled
		case protocol.SmoothingPoolRegistrationEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeSmoothingPoolRegistrationEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing SmoothingPoolRegistrationEnabled: %w", err)
			}

		// NodeDepositEnabled
		case protocol.NodeDepositEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeNodeDepositEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing NodeDepositEnabled: %w", err)
			}

		// VacantMinipoolsEnabled
		case protocol.VacantMinipoolsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			response.GasInfo, err = security.EstimateProposeVacantMinipoolsEnabledGas(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error estimating gas for proposing VacantMinipoolsEnabled: %w", err)
			}
		}
	}

	// Make sure a setting was actually hit
	blankGasInfo := rocketpool.GasInfo{}
	if response.GasInfo == blankGasInfo {
		return nil, fmt.Errorf("[%s - %s] is not a valid PDAO contract and setting name combo", contractName, settingName)
	}

	// Update & return response
	response.CanPropose = true
	return &response, nil

}

func proposeSetting(c *cli.Context, contractName string, settingName string, value string) (*api.ProposePDAOSettingResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.ProposePDAOSettingResponse{}

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
			proposalID, hash, err = security.ProposeCreateLotEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing CreateLotEnabled: %w", err)
			}

		// BidOnLotEnabled
		case protocol.BidOnLotEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeBidOnLotEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing BidOnLotEnabled: %w", err)
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
			proposalID, hash, err = security.ProposeDepositEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing DepositEnabled: %w", err)
			}

		// AssignDepositsEnabled
		case protocol.AssignDepositsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeAssignDepositsEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing AssignDepositsEnabled: %w", err)
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
			proposalID, hash, err = security.ProposeMinipoolSubmitWithdrawableEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing MinipoolSubmitWithdrawableEnabled: %w", err)
			}

		// BondReductionEnabled
		case protocol.BondReductionEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeBondReductionEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing BondReductionEnabled: %w", err)
			}
		}

	case protocol.NetworkSettingsContractName:
		switch settingName {
		// SubmitBalancesEnabled
		case protocol.SubmitBalancesEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeSubmitBalancesEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitBalancesEnabled: %w", err)
			}

		// SubmitRewardsEnabled
		case protocol.SubmitRewardsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeSubmitRewardsEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SubmitRewardsEnabled: %w", err)
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
			proposalID, hash, err = security.ProposeNodeRegistrationEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeRegistrationEnabled: %w", err)
			}

		// SmoothingPoolRegistrationEnabled
		case protocol.SmoothingPoolRegistrationEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeSmoothingPoolRegistrationEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing SmoothingPoolRegistrationEnabled: %w", err)
			}

		// NodeDepositEnabled
		case protocol.NodeDepositEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeNodeDepositEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing NodeDepositEnabled: %w", err)
			}

		// VacantMinipoolsEnabled
		case protocol.VacantMinipoolsEnabledSettingPath:
			newValue, err := cliutils.ValidateBool(valueName, value)
			if err != nil {
				return nil, err
			}
			proposalID, hash, err = security.ProposeVacantMinipoolsEnabled(rp, newValue, opts)
			if err != nil {
				return nil, fmt.Errorf("error proposing VacantMinipoolsEnabled: %w", err)
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
