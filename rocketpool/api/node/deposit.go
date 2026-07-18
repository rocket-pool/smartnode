package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

const (
	prestakeDepositAmount float64 = 1.0
	ValidatorEth          float64 = 32.0
)

// Returns withdrawal credentials pointing to the zero address, used to deliberately create an
// invalid beacon deposit for testing. Refuses to run on mainnet.
//
// Note: the Rocket Pool megapool contracts always recompute the deposit data root using the
// megapool's real withdrawal credentials, so invalid credentials cannot be submitted through
// depositMulti. Instead, --test-invalid-deposit creates a normal protocol deposit and then
// front-runs the 1 ETH prestake by depositing directly to the beacon deposit contract with
// these credentials before assignFunds runs.
func getTestInvalidWithdrawalCredentials(c *cli.Command) (common.Hash, error) {
	cfg, err := services.GetConfig(c)
	if err != nil {
		return common.Hash{}, err
	}
	if cfg.Smartnode.Network.Value.(cfgtypes.Network) == cfgtypes.Network_Mainnet {
		return common.Hash{}, fmt.Errorf("the test-invalid-deposit option cannot be used on mainnet")
	}
	return services.CalculateMegapoolWithdrawalCredentials(common.Address{}), nil
}

func canNodeDeposits(c *cli.Command, count uint64, amountWei *big.Int, minNodeFee float64, salt *big.Int, expressTicketsRequested int64, testInvalidDeposit bool) (*api.CanNodeDepositsResponse, error) {

	// Reject mainnet early when the test flag is set. Deposit data for the protocol path still
	// uses the real megapool withdrawal credentials (the contract requires it).
	if testInvalidDeposit {
		if _, err := getTestInvalidWithdrawalCredentials(c); err != nil {
			return nil, err
		}
	}

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

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeDepositsResponse{
		ValidatorPubkeys: make([]rptypes.ValidatorPubkey, count),
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg1 errgroup.Group
	var creditBalanceWei *big.Int
	var usableCreditBalanceWei *big.Int
	var depositPoolBalance *big.Int
	var expressTicketCount uint64
	var status api.MegapoolDetails
	// Check credit balance
	wg1.Go(func() error {
		creditBalanceWei, err = node.GetNodeCreditAndBalance(rp, nodeAccount.Address, nil)
		if err == nil {
			response.CreditBalance = creditBalanceWei
		}
		return err
	})

	// Get usable credit balance (capped by deposit pool balance)
	wg1.Go(func() error {
		var err error
		usableCreditBalanceWei, err = node.GetNodeUsableCreditAndBalance(rp, nodeAccount.Address, nil)
		if err == nil {
			response.UsableCreditBalance = usableCreditBalanceWei
		}
		return err
	})

	// Get deposit pool balance
	wg1.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(rp, nil)
		if err == nil {
			response.DepositBalance = depositPoolBalance
		}
		return err
	})

	wg1.Go(func() error {
		status, err = services.GetNodeMegapoolDetails(rp, bc, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		expressTicketCount = status.NodeExpressTicketCount
		return nil
	})

	// Check node balance
	wg1.Go(func() error {
		ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
		if err == nil {
			response.NodeBalance = ethBalanceWei
		}
		return err
	})

	// Check node deposits are enabled
	wg1.Go(func() error {
		depositEnabled, err := protocol.GetNodeDepositEnabled(rp, nil)
		if err == nil {
			response.DepositDisabled = !depositEnabled
		}
		return err
	})

	// Check whether the node has debt
	wg1.Go(func() error {
		// Load the megapool contract

		megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}

		// Check whether the megapool is deployed
		deployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		if !deployed {
			return nil
		}

		mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
		if err != nil {
			return err
		}
		hasDebt, err := mp.GetDebt(nil)
		if err == nil {
			response.NodeHasDebt = hasDebt.Cmp(big.NewInt(0)) > 0
		}
		return err
	})
	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	// Check for insufficient balance
	// Total balance = node wallet + usable credit
	// Usable credit includes node ETH balance stored in contract
	totalBalance := big.NewInt(0).Add(response.NodeBalance, usableCreditBalanceWei)
	response.InsufficientBalance = (amountWei.Cmp(totalBalance) > 0)

	// Check if credit can be used (either full or partial)
	// We can use credit if usable credit + node wallet balance >= amount needed
	response.CanUseCredit = usableCreditBalanceWei.Cmp(big.NewInt(0)) > 0 && totalBalance.Cmp(amountWei) >= 0

	// Check if we can't use credit AND don't have enough in wallet
	// This happens when usable credit is 0 (pool empty) and wallet balance is insufficient but user has credit
	if creditBalanceWei.Cmp(big.NewInt(0)) > 0 && !response.CanUseCredit && response.NodeBalance.Cmp(amountWei) < 0 {
		response.InsufficientBalanceWithoutCredit = true
	}

	// Update response && Break before the gas estimator if depositing won't work
	response.CanDeposit = !response.InsufficientBalance && !response.InsufficientBalanceWithoutCredit && !response.InvalidAmount && !response.DepositDisabled && !response.NodeHasDebt
	if !response.CanDeposit {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get how much credit to use
	if response.CanUseCredit {
		// Calculate how much ETH to send with the transaction
		// Use usable credit (capped by deposit pool balance) to determine the shortfall
		remainingAmount := big.NewInt(0).Sub(amountWei, usableCreditBalanceWei)
		if remainingAmount.Cmp(big.NewInt(0)) > 0 {
			// Send the remaining amount if the usable credit isn't enough to cover the whole deposit
			opts.Value = remainingAmount
		}
	} else {
		opts.Value = amountWei
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Get the withdrawal credentials
	withdrawalCredentials := services.CalculateMegapoolWithdrawalCredentials(megapoolAddress)

	// Create deposit data for all deposits (for gas estimation)
	// We need to create unique validator keys for each deposit to get accurate gas estimates
	depositAmount := uint64(1e9) // 1 ETH in gwei
	deposits := node.Deposits{}

	keyCount, err := w.GetValidatorKeyCount()
	if err != nil {
		return nil, err
	}

	// Get the next validator key for gas estimation
	validatorKeys, err := w.GetValidatorKeys(keyCount, uint(count))
	if err != nil {
		return nil, err
	}
	expressTicketsRequested = min(expressTicketsRequested, int64(expressTicketCount))
	lastBondAdded := big.NewInt(0)
	bondedEth := status.NodeBond
	if bondedEth == nil {
		bondedEth = big.NewInt(0)
	}
	queuedBondEth := status.NodeQueuedBond
	if queuedBondEth == nil {
		queuedBondEth = big.NewInt(0)
	}
	bondedEth = bondedEth.Add(bondedEth, queuedBondEth)
	for i := uint64(0); i < count; i++ {
		bondedEth = bondedEth.Add(bondedEth, lastBondAdded)
		// Get the bond requirement for each validator
		bondRequirement, err := node.GetBondRequirement(rp, big.NewInt(int64(uint64(status.ActiveValidatorCount)+i+1)), nil)
		if err != nil {
			return nil, err
		}
		lastBondAdded = bondRequirement
		// Find the bond requirement for the next validator
		nextBondRequirement := bondRequirement.Sub(bondRequirement, bondedEth)
		// Get validator deposit data and associated parameters
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKeys[i].PrivateKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return nil, err
		}
		pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
		signature := rptypes.BytesToValidatorSignature(depositData.Signature)

		// Add to deposits array
		deposits = append(deposits, node.NodeDeposit{
			BondAmount:         nextBondRequirement,
			UseExpressTicket:   expressTicketsRequested > 0,
			ValidatorPubkey:    pubKey[:],
			ValidatorSignature: signature[:],
			DepositDataRoot:    depositDataRoot,
		})

		// Store the pubkey in the response
		response.ValidatorPubkeys[i] = pubKey
		expressTicketsRequested--
	}

	// Ensure count is valid
	if count == 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	gasInfo, err := node.EstimateDepositMultiGas(rp, deposits, opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating gas for depositMulti: %w", err)
	}
	response.GasInfo = gasInfo

	return &response, nil

}

func nodeDeposits(c *cli.Command, count uint64, amountWei *big.Int, minNodeFee float64, salt *big.Int, useCreditBalance bool, expressTicketsRequested int64, submit bool, testInvalidDeposit bool, opts *bind.TransactOpts) (*api.NodeDepositsResponse, error) {

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

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeDepositsResponse{}

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
	scrubPeriodUnix, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		return nil, err
	}
	scrubPeriod := time.Duration(scrubPeriodUnix) * time.Second
	response.ScrubPeriod = scrubPeriod

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	status, err := services.GetNodeMegapoolDetails(rp, bc, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	expressTicketCount := status.NodeExpressTicketCount

	// Get the withdrawal credentials
	withdrawalCredentials := services.CalculateMegapoolWithdrawalCredentials(megapoolAddress)

	// Get the node's credit and ETH staked on behalf balance
	creditBalanceWei, err := node.GetNodeCreditAndBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Set the value to the total amount needed// Get how much credit to use
	if useCreditBalance {
		remainingAmount := big.NewInt(0).Sub(amountWei, creditBalanceWei)
		if remainingAmount.Cmp(big.NewInt(0)) > 0 {
			// Send the remaining amount if the credit isn't enough to cover the whole deposit
			opts.Value = remainingAmount
		}
	} else {
		opts.Value = amountWei
	}

	// Create validator keys and deposit data for all deposits
	depositAmount := uint64(1e9) // 1 ETH in gwei
	deposits := make([]node.NodeDeposit, count)
	response.ValidatorPubkeys = make([]rptypes.ValidatorPubkey, count)
	validatorKeys := make([]*eth2types.BLSPrivateKey, count)

	expressTicketsRequested = min(expressTicketsRequested, int64(expressTicketCount))
	lastBondAdded := big.NewInt(0)
	bondedEth := status.NodeBond
	if bondedEth == nil {
		bondedEth = big.NewInt(0)
	}
	queuedBondEth := status.NodeQueuedBond
	if queuedBondEth == nil {
		queuedBondEth = big.NewInt(0)
	}
	bondedEth = bondedEth.Add(bondedEth, queuedBondEth)
	for i := uint64(0); i < count; i++ {
		bondedEth = bondedEth.Add(bondedEth, lastBondAdded)
		// Get the bond requirement for each validator
		bondRequirement, err := node.GetBondRequirement(rp, big.NewInt(int64(uint64(status.ActiveValidatorCount)+i+1)), nil)
		if err != nil {
			return nil, err
		}
		lastBondAdded = bondRequirement
		// Find the bond requirement for the next validator
		nextBondRequirement := bondRequirement.Sub(bondRequirement, bondedEth)

		validatorKey, err := w.CreateValidatorKey()
		if err != nil {
			return nil, err
		}
		validatorKeys[i] = validatorKey
		// Get validator deposit data and associated parameters
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return nil, err
		}
		pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
		signature := rptypes.BytesToValidatorSignature(depositData.Signature)

		// Make sure a validator with this pubkey doesn't already exist
		status, err := bc.GetValidatorStatus(pubKey, nil)
		if err != nil {
			return nil, fmt.Errorf("Error checking for existing validator status for deposit %d/%d: %w\nYour funds have not been deposited for your own safety.", i+1, count, err)
		}
		if status.Exists {
			return nil, fmt.Errorf("**** ALERT ****\n"+
				"The following validator pubkey is already in use on the Beacon chain:\n\t%s\n"+
				"Rocket Pool will not allow you to deposit this validator for your own safety so you do not get slashed.\n"+
				"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.\n"+
				"***************\n", pubKey.Hex())
		}

		// Do a final sanity check
		err = validateDepositInfo(eth2Config, depositAmount, pubKey, withdrawalCredentials, signature)
		if err != nil {
			return nil, fmt.Errorf("Your deposit %d/%d failed the validation safety check: %w\n"+
				"For your safety, this deposit will not be submitted and your ETH will not be staked.\n"+
				"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS and include the following information:\n"+
				"\tDomain Type: 0x%s\n"+
				"\tGenesis Fork Version: 0x%s\n"+
				"\tGenesis Validator Root: 0x%s\n"+
				"\tDeposit Amount: %d gwei\n"+
				"\tValidator Pubkey: %s\n"+
				"\tWithdrawal Credentials: %s\n"+
				"\tSignature: %s\n",
				i+1, count, err,
				hex.EncodeToString(eth2types.DomainDeposit[:]),
				hex.EncodeToString(eth2Config.GenesisForkVersion),
				hex.EncodeToString(eth2types.ZeroGenesisValidatorsRoot),
				depositAmount,
				pubKey.Hex(),
				withdrawalCredentials.Hex(),
				signature.Hex(),
			)
		}

		deposits[i] = node.NodeDeposit{
			BondAmount:         nextBondRequirement,
			UseExpressTicket:   expressTicketsRequested > 0,
			ValidatorPubkey:    pubKey[:],
			ValidatorSignature: signature[:],
			DepositDataRoot:    depositDataRoot,
		}

		response.ValidatorPubkeys[i] = pubKey
		expressTicketsRequested--
	}

	// Do not send transaction unless requested
	opts.NoSend = !submit

	// Make multiple deposits in a single transaction
	tx, err := node.DepositMulti(rp, deposits, opts)
	if err != nil {
		return nil, err
	}

	// Print transaction if requested
	if !submit {
		b, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		fmt.Printf("%x\n", b)
	} else {
		// Save wallet
		if err := w.Save(); err != nil {
			return nil, err
		}

		// Front-run the beacon prestake with invalid withdrawal credentials so the
		// validator appears on the beacon chain with credentials that do not match the
		// megapool. This exercises dissolve-invalid-credentials handling.
		if testInvalidDeposit {
			receipt, err := bind.WaitMined(context.Background(), rp.Client, tx)
			if err != nil {
				return nil, fmt.Errorf("error waiting for depositMulti tx %s: %w", tx.Hash().Hex(), err)
			}
			if receipt.Status == 0 {
				return nil, fmt.Errorf("depositMulti transaction %s reverted", tx.Hash().Hex())
			}

			if err := submitTestInvalidBeaconDeposits(c, eth2Config, validatorKeys, response.ValidatorPubkeys, opts); err != nil {
				return nil, fmt.Errorf("protocol deposit succeeded (tx %s) but failed to submit test-invalid beacon deposit(s): %w", tx.Hash().Hex(), err)
			}
		}
	}

	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}

// submitTestInvalidBeaconDeposits deposits 1 ETH per validator directly to the beacon deposit
// contract with zero-address withdrawal credentials. This is used to test the dissolve-invalid-credentials handling.
func submitTestInvalidBeaconDeposits(c *cli.Command, eth2Config beacon.Eth2Config, validatorKeys []*eth2types.BLSPrivateKey, pubkeys []rptypes.ValidatorPubkey, opts *bind.TransactOpts) error {

	invalidCredentials, err := getTestInvalidWithdrawalCredentials(c)
	if err != nil {
		return err
	}

	rocketPool, err := services.GetRocketPool(c)
	if err != nil {
		return err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return err
	}
	// Fresh transactor so Value/gas from depositMulti do not leak into the beacon deposits
	beaconOpts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}
	// Preserve fee settings from the caller's opts when present
	if opts != nil {
		beaconOpts.GasFeeCap = opts.GasFeeCap
		beaconOpts.GasTipCap = opts.GasTipCap
		beaconOpts.GasPrice = opts.GasPrice
	}

	blankAddress := common.Address{}
	casperAddress, err := rocketPool.GetAddress("casperDeposit", nil)
	if err != nil {
		return fmt.Errorf("error getting Beacon deposit contract address: %w", err)
	}
	if casperAddress == nil || *casperAddress == blankAddress {
		return fmt.Errorf("Beacon deposit contract address was empty (0x0)")
	}

	depositContract, err := contracts.NewBeaconDeposit(*casperAddress, rocketPool.Client)
	if err != nil {
		return fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	depositAmountGwei := uint64(1e9) // 1 ETH prestake amount
	beaconOpts.Value = eth.EthToWei(prestakeDepositAmount)

	for i, key := range validatorKeys {
		// Clear nonce/gas so each deposit gets a fresh pending nonce after the previous mine
		beaconOpts.Nonce = nil
		beaconOpts.GasLimit = 0

		depositData, depositDataRoot, err := validator.GetDepositData(key, invalidCredentials, eth2Config, depositAmountGwei)
		if err != nil {
			return fmt.Errorf("error creating invalid deposit data for validator %d: %w", i+1, err)
		}
		signature := rptypes.BytesToValidatorSignature(depositData.Signature)

		tx, err := depositContract.Deposit(beaconOpts, pubkeys[i][:], invalidCredentials[:], signature[:], depositDataRoot)
		if err != nil {
			return fmt.Errorf("error submitting invalid beacon deposit for validator %s: %w", pubkeys[i].Hex(), err)
		}
		fmt.Printf("Submitted test-invalid beacon deposit for validator %s (tx %s) with withdrawal credentials %s\n",
			pubkeys[i].Hex(), tx.Hash().Hex(), invalidCredentials.Hex())

		receipt, err := bind.WaitMined(context.Background(), rocketPool.Client, tx)
		if err != nil {
			return fmt.Errorf("error waiting for invalid beacon deposit tx %s: %w", tx.Hash().Hex(), err)
		}
		if receipt.Status == 0 {
			return fmt.Errorf("invalid beacon deposit transaction %s for validator %s reverted", tx.Hash().Hex(), pubkeys[i].Hex())
		}
	}

	return nil
}

func validateDepositInfo(eth2Config beacon.Eth2Config, depositAmount uint64, pubkey rptypes.ValidatorPubkey, withdrawalCredentials common.Hash, signature rptypes.ValidatorSignature) error {

	// Get the deposit domain based on the eth2 config
	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return err
	}

	// Create the deposit struct
	depositData := new(ethpb.Deposit_Data)
	depositData.Amount = depositAmount
	depositData.PublicKey = pubkey.Bytes()
	depositData.WithdrawalCredentials = withdrawalCredentials.Bytes()
	depositData.Signature = signature.Bytes()

	// Validate the signature
	err = prdeposit.VerifyDepositSignature(depositData, depositDomain)
	return err

}
