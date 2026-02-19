package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	prestakeDepositAmount float64 = 1.0
	ValidatorEth          float64 = 32.0
)

func canNodeDeposits(c *cli.Context, count uint64, amountWei *big.Int, minNodeFee float64, salt *big.Int, expressTicketsRequested int64) (*api.CanNodeDepositsResponse, error) {

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

	saturnDeployed, err := state.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	if !saturnDeployed {
		return nil, fmt.Errorf("Multiple deposits are only supported after Saturn deployment")
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

	if saturnDeployed {
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
	}
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
	response.CanDeposit = !(response.InsufficientBalance || response.InsufficientBalanceWithoutCredit || response.InvalidAmount || response.DepositDisabled || response.NodeHasDebt)
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

func nodeDeposits(c *cli.Context, count uint64, amountWei *big.Int, minNodeFee float64, salt *big.Int, useCreditBalance bool, expressTicketsRequested int64, submit bool) (*api.NodeDepositsResponse, error) {

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

	saturnDeployed, err := state.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeDepositsResponse{}

	if !saturnDeployed {
		return nil, fmt.Errorf("Multiple deposits are only supported after Saturn deployment")
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

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
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

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
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
	}

	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

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
