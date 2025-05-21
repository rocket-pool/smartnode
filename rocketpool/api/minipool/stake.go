package minipool

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/urfave/cli"

	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func canStakeMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanStakeMinipoolResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CanStakeMinipoolResponse{
		CanStake: false,
	}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
		return nil, err
	}

	// Check the minipool's status
	status, err := mp.GetStatusDetails(nil)
	if err != nil {
		return nil, err
	}

	if status.Status == rptypes.Prelaunch {

		// Get the scrub period
		scrubPeriodSeconds, err := trustednode.GetScrubPeriod(rp, nil)
		if err != nil {
			return nil, err
		}
		scrubPeriod := time.Duration(scrubPeriodSeconds) * time.Second

		// Get the time of the latest block
		latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return nil, fmt.Errorf("Can't get the latest block time: %w", err)
		}
		latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

		creationTime := status.StatusTime
		remainingTime := creationTime.Add(scrubPeriod).Sub(latestBlockTime)
		if remainingTime < 0 {
			response.CanStake = true
		}
	}

	if response.CanStake {
		// Get eth2 config
		eth2Config, err := bc.GetEth2Config()
		if err != nil {
			return nil, err
		}

		// Get minipool withdrawal credentials
		withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, mp.GetAddress(), nil)
		if err != nil {
			return nil, err
		}

		// Get the validator key for the minipool
		validatorPubkey, err := minipool.GetMinipoolPubkey(rp, mp.GetAddress(), nil)
		if err != nil {
			return nil, err
		}
		validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
		if err != nil {
			return nil, err
		}

		// Get the minipool type
		depositType, err := minipool.GetMinipoolDepositType(rp, mp.GetAddress(), nil)
		if err != nil {
			return nil, fmt.Errorf("error getting deposit type for minipool %s: %w", mp.GetAddress().Hex(), err)
		}

		var depositAmount uint64
		switch depositType {
		case rptypes.Full, rptypes.Half, rptypes.Empty:
			depositAmount = uint64(16e9) // 16 ETH in gwei
		case rptypes.Variable:
			depositAmount = uint64(31e9) // 31 ETH in gwei
		default:
			return nil, fmt.Errorf("error staking minipool %s: unknown deposit type %d", mp.GetAddress().Hex(), depositType)
		}

		// Get validator deposit data
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return nil, err
		}

		// Get transactor
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}

		// Get the gas limit
		signature := rptypes.BytesToValidatorSignature(depositData.Signature)
		gasInfo, err := mp.EstimateStakeGas(signature, depositDataRoot, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
	}

	// Return response
	return &response, nil

}

func stakeMinipool(c *cli.Context, minipoolAddress common.Address) (*api.StakeMinipoolResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.StakeMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
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

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Get minipool withdrawal credentials
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, mp.GetAddress(), nil)
	if err != nil {
		return nil, err
	}

	// Get the validator key for the minipool
	validatorPubkey, err := minipool.GetMinipoolPubkey(rp, mp.GetAddress(), nil)
	if err != nil {
		return nil, err
	}
	validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return nil, err
	}

	// Get the minipool type
	depositType, err := minipool.GetMinipoolDepositType(rp, mp.GetAddress(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting deposit type for minipool %s: %w", mp.GetAddress().Hex(), err)
	}

	var depositAmount uint64
	switch depositType {
	case rptypes.Full, rptypes.Half, rptypes.Empty:
		depositAmount = uint64(16e9) // 16 ETH in gwei
	case rptypes.Variable:
		depositAmount = uint64(31e9) // 31 ETH in gwei
	default:
		return nil, fmt.Errorf("error staking minipool %s: unknown deposit type %d", mp.GetAddress().Hex(), depositType)
	}

	// Get validator deposit data
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
	if err != nil {
		return nil, err
	}

	// Stake the minipool
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)
	hash, err := mp.Stake(signature, depositDataRoot, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
