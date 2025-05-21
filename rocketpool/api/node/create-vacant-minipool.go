package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func canCreateVacantMinipool(c *cli.Context, amountWei *big.Int, minNodeFee float64, salt *big.Int, pubkey rptypes.ValidatorPubkey) (*api.CanCreateVacantMinipoolResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
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
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanCreateVacantMinipoolResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Adjust the salt
	if salt.Cmp(big.NewInt(0)) == 0 {
		nonce, err := ec.NonceAt(context.Background(), nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}
		salt.SetUint64(nonce)
	}

	// Data
	var wg1 errgroup.Group
	var ethMatched *big.Int
	var ethMatchedLimit *big.Int

	// Check node deposits are enabled
	wg1.Go(func() error {
		depositEnabled, err := protocol.GetVacantMinipoolsEnabled(rp, nil)
		if err == nil {
			response.DepositDisabled = !depositEnabled
		}
		return err
	})

	// Get node staking information
	wg1.Go(func() error {
		var err error
		ethMatched, err = node.GetNodeEthMatched(rp, nodeAccount.Address, nil)
		return err
	})
	wg1.Go(func() error {
		var err error
		ethMatchedLimit, err = node.GetNodeEthMatchedLimit(rp, nodeAccount.Address, nil)
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	// Get the next minipool address and withdrawal credentials
	minipoolAddress, err := minipool.GetExpectedAddress(rp, nodeAccount.Address, salt, nil)
	if err != nil {
		return nil, err
	}

	// Check data
	validatorEthWei := eth.EthToWei(ValidatorEth)
	matchRequest := big.NewInt(0).Sub(validatorEthWei, amountWei)
	availableToMatch := big.NewInt(0).Sub(ethMatchedLimit, ethMatched)

	response.InsufficientRplStake = (availableToMatch.Cmp(matchRequest) == -1)
	response.MinipoolAddress = minipoolAddress

	// Update response
	response.CanDeposit = !(response.InsufficientRplStake || response.InvalidAmount || response.DepositDisabled)
	if !response.CanDeposit {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Check if the pubkey is for an existing active_ongoing validator
	validatorStatus, err := bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking status of existing validator: %w", err)
	}
	if !validatorStatus.Exists {
		return nil, fmt.Errorf("validator %s does not exist on the Beacon chain. If you recently created it, please wait until the Consensus layer has processed your deposits.", pubkey.Hex())
	}
	if validatorStatus.Status != beacon.ValidatorState_ActiveOngoing {
		return nil, fmt.Errorf("validator %s must be in the active_ongoing state to be migrated, but it is currently in %s.", pubkey.Hex(), string(validatorStatus.Status))
	}
	if cfg.Smartnode.Network.Value.(cfgtypes.Network) != cfgtypes.Network_Devnet && validatorStatus.WithdrawalCredentials[0] != 0x00 {
		return nil, fmt.Errorf("validator %s already has withdrawal credentials [%s], which are not BLS credentials.", pubkey.Hex(), validatorStatus.WithdrawalCredentials.Hex())
	}

	// Convert the existing balance from gwei to wei
	balanceWei := big.NewInt(0).SetUint64(validatorStatus.Balance)
	balanceWei.Mul(balanceWei, big.NewInt(1e9))

	// Run the deposit gas estimator
	gasInfo, err := node.EstimateCreateVacantMinipoolGas(rp, amountWei, minNodeFee, pubkey, salt, minipoolAddress, balanceWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	return &response, nil

}

func createVacantMinipool(c *cli.Context, amountWei *big.Int, minNodeFee float64, salt *big.Int, pubkey rptypes.ValidatorPubkey) (*api.CreateVacantMinipoolResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
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
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CreateVacantMinipoolResponse{}

	// Adjust the salt
	if salt.Cmp(big.NewInt(0)) == 0 {
		nonce, err := ec.NonceAt(context.Background(), nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}
		salt.SetUint64(nonce)
	}

	// Make sure ETH2 is on the correct chain
	depositContractInfo, err := getDepositContractInfo(c)
	if err != nil {
		return nil, err
	}
	if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
		depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
		return nil, fmt.Errorf("Beacon network mismatch! Expected %s on chain %d, but beacon is using %s on chain %d.",
			depositContractInfo.RPDepositContract.Hex(),
			depositContractInfo.RPNetwork,
			depositContractInfo.BeaconDepositContract.Hex(),
			depositContractInfo.BeaconNetwork)
	}

	// Get the scrub period
	scrubPeriodUnix, err := trustednode.GetPromotionScrubPeriod(rp, nil)
	if err != nil {
		return nil, err
	}
	scrubPeriod := time.Duration(scrubPeriodUnix) * time.Second
	response.ScrubPeriod = scrubPeriod

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the next minipool address and withdrawal credentials
	minipoolAddress, err := minipool.GetExpectedAddress(rp, nodeAccount.Address, salt, nil)
	if err != nil {
		return nil, err
	}
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}
	response.WithdrawalCredentials = withdrawalCredentials

	// Check if the pubkey is for an existing active_ongoing validator
	validatorStatus, err := bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking status of existing validator: %w", err)
	}
	if !validatorStatus.Exists {
		return nil, fmt.Errorf("validator %s does not exist.", pubkey.Hex())
	}
	if validatorStatus.Status != beacon.ValidatorState_ActiveOngoing {
		return nil, fmt.Errorf("validator %s must be in the active_ongoing state to be migrated, but it is currently in %s.", pubkey.Hex(), string(validatorStatus.Status))
	}
	if cfg.Smartnode.Network.Value.(cfgtypes.Network) != cfgtypes.Network_Devnet && validatorStatus.WithdrawalCredentials[0] != 0x00 {
		return nil, fmt.Errorf("validator %s already has withdrawal credentials [%s], which are not BLS credentials.", pubkey.Hex(), validatorStatus.WithdrawalCredentials.Hex())
	}

	// Convert the existing balance from gwei to wei
	balanceWei := big.NewInt(0).SetUint64(validatorStatus.Balance)
	balanceWei.Mul(balanceWei, big.NewInt(1e9))

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Create the minipool
	tx, err := node.CreateVacantMinipool(rp, amountWei, minNodeFee, pubkey, salt, minipoolAddress, balanceWei, opts)
	if err != nil {
		return nil, err
	}

	response.TxHash = tx.Hash()
	response.MinipoolAddress = minipoolAddress

	// Return response
	return &response, nil

}
