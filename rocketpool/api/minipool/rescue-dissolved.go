package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
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
	details.BeaconBalance = big.NewInt(0)

	// Ignore minipools that are too old
	if details.MinipoolVersion < 3 {
		details.CanRescue = false
		return details, nil
	}

	// Get the balance / share info and status details
	var pubkey rptypes.ValidatorPubkey
	var wg1 errgroup.Group
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
	if details.MinipoolStatus != rptypes.Dissolved {
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
	beaconBalanceGwei := big.NewInt(0).SetUint64(beaconStatus.Balance)
	details.BeaconBalance = big.NewInt(0).Mul(beaconBalanceGwei, big.NewInt(1e9))

	// Make sure it doesn't already have 32 ETH in it
	requiredBalance := eth.EthToWei(32)
	if details.BeaconBalance.Cmp(requiredBalance) >= 0 {
		details.CanRescue = false
		return details, nil
	}

	// Passed the checks!
	details.CanRescue = true

	// Get the simulated deposit TX
	one := eth.EthToWei(1)
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return api.MinipoolRescueDissolvedDetails{}, err
	}
	opts.Value = one
	opts.NoSend = true
	opts.GasLimit = 0

	// Get the gas info for depositing
	tx, err := getDepositTx(rp, w, bc, minipoolAddress, one, opts)
	if err != nil {
		return api.MinipoolRescueDissolvedDetails{}, fmt.Errorf("error estimating gas for rescue deposit on minipool %s: %w", minipoolAddress.Hex(), err)
	}
	gasLimit := tx.Gas()
	safeGasLimit := uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
	if gasLimit > rocketpool.MaxGasLimit {
		return api.MinipoolRescueDissolvedDetails{}, fmt.Errorf("estimated gas of %d is greater than the max gas limit of %d", gasLimit, rocketpool.MaxGasLimit)
	}
	if safeGasLimit > rocketpool.MaxGasLimit {
		safeGasLimit = rocketpool.MaxGasLimit
	}

	details.GasInfo = rocketpool.GasInfo{
		EstGasLimit:  gasLimit,
		SafeGasLimit: safeGasLimit,
	}

	return details, nil

}

// Create a transaction for submitting a rescue deposit, optionally simulating it only for gas estimation
func getDepositTx(rp *rocketpool.RocketPool, w *wallet.Wallet, bc beacon.Client, minipoolAddress common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {

	blankAddress := common.Address{}
	casperAddress, err := rp.GetAddress("casperDeposit", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon deposit contract address: %w", err)
	}
	if casperAddress == nil || *casperAddress == blankAddress {
		return nil, fmt.Errorf("Beacon deposit contract address was empty (0x0).")
	}

	depositContract, err := contracts.NewBeaconDeposit(*casperAddress, rp.Client)
	if err != nil {
		return nil, fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
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

	// Get the deposit amount in gwei
	amountGwei := big.NewInt(0).Div(amount, big.NewInt(1e9)).Uint64()

	// Get validator deposit data
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, amountGwei)
	if err != nil {
		return nil, err
	}
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)

	// Get the tx
	tx, err := depositContract.Deposit(opts, validatorPubkey[:], withdrawalCredentials[:], signature[:], depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error performing rescue deposit: %s", err.Error())
	}

	// Return
	return tx, nil

}

func rescueDissolvedMinipool(c *cli.Context, minipoolAddress common.Address, amount *big.Int, submit bool) (*api.RescueDissolvedMinipoolResponse, error) {

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
	response := api.RescueDissolvedMinipoolResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	opts.Value = amount

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	opts.NoSend = !submit

	// Submit the rescue deposit
	tx, err := getDepositTx(rp, w, bc, minipoolAddress, amount, opts)
	if err != nil {
		return nil, fmt.Errorf("error submitting rescue deposit: %w", err)
	}

	// Print transaction if requested
	if !submit {
		b, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		fmt.Printf("%x\n", b)
	}

	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
