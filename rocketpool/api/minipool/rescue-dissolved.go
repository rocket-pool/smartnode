package minipool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func getMinipoolRescueDissolvedDetailsForNode(c *cli.Context) (*api.GetMinipoolRescueDissolvedDetailsForNodeResponse, error) {

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
	response := api.GetMinipoolRescueDissolvedDetailsForNodeResponse{}

	// Check if Atlas has been deployed
	isAtlasDeployed, err := state.IsAtlasDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	response.IsAtlasDeployed = isAtlasDeployed
	if !isAtlasDeployed {
		return &response, nil
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the minipool addresses for this node
	addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Iterate over each minipool to get its close details
	details := make([]api.MinipoolRescueDissolvedDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolDetailsBatchSize
		if mei > len(addresses) {
			mei = len(addresses)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				mpDetails, err := getMinipoolRescueDissolvedDetails(rp, w, bc, address, nodeAccount.Address)
				if err == nil {
					details[mi] = mpDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return nil, err
		}

	}

	response.Details = details
	return &response, nil

}

func getMinipoolRescueDissolvedDetails(rp *rocketpool.RocketPool, w *wallet.Wallet, bc beacon.Client, minipoolAddress common.Address, nodeAddress common.Address) (api.MinipoolRescueDissolvedDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return api.MinipoolRescueDissolvedDetails{}, err
	}

	// Validate minipool owner
	if err := validateMinipoolOwner(mp, nodeAddress); err != nil {
		return api.MinipoolRescueDissolvedDetails{}, err
	}

	var details api.MinipoolRescueDissolvedDetails
	details.Address = mp.GetAddress()
	details.MinipoolVersion = mp.GetVersion()
	details.Balance = big.NewInt(0)

	// Ignore minipools that are too old
	if details.MinipoolVersion < 3 {
		details.CanRescue = false
		return details, nil
	}

	// Get the balance / share info and status details
	var pubkey types.ValidatorPubkey
	var wg1 errgroup.Group
	wg1.Go(func() error {
		var err error
		details.Balance, err = rp.Client.BalanceAt(context.Background(), minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting balance of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		details.IsFinalized, err = mp.GetFinalised(nil)
		if err != nil {
			return fmt.Errorf("error getting finalized status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		details.MinipoolStatus, err = mp.GetStatus(nil)
		if err != nil {
			return fmt.Errorf("error getting status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		pubkey, err = minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting pubkey for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})

	if err := wg1.Wait(); err != nil {
		return api.MinipoolRescueDissolvedDetails{}, err
	}

	// Can't rescue a minipool that's already finalized
	if details.IsFinalized {
		details.CanRescue = false
		return details, nil
	}

	// Make sure it's dissolved
	if details.MinipoolStatus != types.Dissolved {
		details.CanRescue = false
		return details, nil
	}

	// Make sure it doesn't already have 32 ETH in it
	requiredBalance := eth.EthToWei(32)
	if details.Balance.Cmp(requiredBalance) >= 0 {
		details.CanRescue = false
		return details, nil
	}

	// Check the Beacon status
	beaconStatus, err := bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return api.MinipoolRescueDissolvedDetails{}, fmt.Errorf("error getting validator status for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), pubkey.Hex(), err)
	}
	details.BeaconState = beaconStatus.Status
	if details.BeaconState != beacon.ValidatorState_PendingInitialized {
		details.CanRescue = false
		return details, nil
	}

	// Passed the checks!
	details.CanRescue = true

	// Get the gas info for depositing
	remainingAmount := big.NewInt(0).Sub(requiredBalance, details.Balance)
	details.GasInfo, err = getDepositGasInfo(rp, w, bc, minipoolAddress, remainingAmount)
	if err != nil {
		return api.MinipoolRescueDissolvedDetails{}, fmt.Errorf("error estimating gas for rescue deposit on minipool %s: %w", minipoolAddress.Hex(), err)
	}

	return details, nil

}

// Estimate the gas required to do a rescue deposit
func getDepositGasInfo(rp *rocketpool.RocketPool, w *wallet.Wallet, bc beacon.Client, minipoolAddress common.Address, amount *big.Int) (rocketpool.GasInfo, error) {

	blankAddress := common.Address{}
	casperAddress, err := rp.GetAddress("casperDeposit", nil)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error getting Beacon deposit contract address: %w", err)
	}
	if casperAddress == nil || *casperAddress == blankAddress {
		return rocketpool.GasInfo{}, fmt.Errorf("Beacon deposit contract address was empty (0x0).")
	}

	depositContract, err := contracts.NewBeaconDeposit(*casperAddress, rp.Client)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return rocketpool.GasInfo{}, err
	}

	// Get minipool withdrawal credentials
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, mp.GetAddress(), nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}

	// Get the validator key for the minipool
	validatorPubkey, err := minipool.GetMinipoolPubkey(rp, mp.GetAddress(), nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}

	// Get the deposit amount in gwei
	amountGwei := big.NewInt(0).Div(amount, big.NewInt(1e9)).Uint64()

	// Get validator deposit data
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, amountGwei)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	opts.Value = amount
	opts.NoSend = true // Simulation only
	opts.GasLimit = 0  // Estimate gas limit

	// Get the gas info
	tx, err := depositContract.Deposit(opts, validatorPubkey[:], withdrawalCredentials[:], signature[:], depositDataRoot)
	gasLimit := tx.Gas()
	safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
	if gasLimit > rocketpool.MaxGasLimit {
		return rocketpool.GasInfo{}, fmt.Errorf("estimated gas of %d is greater than the max gas limit of %d", gasLimit, rocketpool.MaxGasLimit)
	}
	if safeGasLimit > rocketpool.MaxGasLimit {
		safeGasLimit = rocketpool.MaxGasLimit
	}

	gasInfo := rocketpool.GasInfo{
		EstGasLimit:  gasLimit,
		SafeGasLimit: safeGasLimit,
	}

	return gasInfo, nil

}

func rescueDissolvedMinipool(c *cli.Context, minipoolAddress common.Address) (*api.RescueDissolvedMinipoolResponse, error) {

	// TODO!

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

	// Response
	response := api.RescueDissolvedMinipoolResponse{}

	// Check if Atlas has been deployed
	isAtlasDeployed, err := state.IsAtlasDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	if !isAtlasDeployed {
		return nil, fmt.Errorf("Atlas has not been deployed yet.")
	}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Check if it's an upgraded Atlas-era minipool
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", minipoolAddress.Hex(), mp.GetVersion())
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

	// Get some details
	var status types.MinipoolStatus
	var distributed bool
	var wg errgroup.Group
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		if err != nil {
			return fmt.Errorf("error getting status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg.Go(func() error {
		var err error
		distributed, err = mpv3.GetUserDistributed(nil)
		if err != nil {
			return fmt.Errorf("error checking distributed flag of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	if status == types.Dissolved {
		// If it's dissolved, just close it
		hash, err := mp.Close(opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else if distributed {
		// It's already been distributed so just finalize it
		hash, err := mpv3.Finalise(opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else {
		// Do a distribution, which will finalize it
		hash, err := mpv3.DistributeBalance(false, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	}

	// Return response
	return &response, nil

}
