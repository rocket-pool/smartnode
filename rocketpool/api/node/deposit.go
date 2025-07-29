package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	nodev131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/node"
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
	houstonBondAmount     float64 = 8.0
	ValidatorEth          float64 = 32.0
)

func canNodeDeposit(c *cli.Context, minNodeFee float64, salt *big.Int, numValidators uint64, numExpressTickets uint32) (*api.CanNodeDepositResponse, error) {

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
	response := api.CanNodeDepositResponse{}

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
		// Adjust the salt
		if salt.Cmp(big.NewInt(0)) == 0 {
			nonce, err := ec.NonceAt(context.Background(), nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			salt.SetUint64(nonce)
		}
	}

	// Data
	var wg1 errgroup.Group
	var minipoolAddress common.Address
	var depositPoolBalance *big.Int
	var reducedBond float64
	// Check credit balance
	wg1.Go(func() error {
		ethBalanceWei, err := node.GetNodeCreditAndBalance(rp, nodeAccount.Address, nil)
		if err == nil {
			response.CreditBalance = ethBalanceWei
		}
		return err
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

	// Get deposit pool balance
	wg1.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(rp, nil)
		return err
	})

	if saturnDeployed {
		wg1.Go(func() error {
			reducedBond, err = protocol.GetReducedBond(rp, nil)
			if err != nil {
				return err
			}
			return nil
		})
	} else {
		reducedBond = houstonBondAmount
	}

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	// Check for insufficient balance
	totalBalance := big.NewInt(0).Add(response.NodeBalance, response.CreditBalance)
	reducedBondWei := eth.EthToWei(reducedBond)
	totalAmountNeeded := big.NewInt(0).Mul(reducedBondWei, big.NewInt(int64(numValidators)))
	response.InsufficientBalance = (totalAmountNeeded.Cmp(totalBalance) > 0)

	// Check if the credit balance can be used
	response.DepositBalance = depositPoolBalance
	response.CanUseCredit = (depositPoolBalance.Cmp(eth.EthToWei(1)) >= 0) && totalBalance.Cmp(totalAmountNeeded) >= 0
	var usableBalance *big.Int
	if response.CreditBalance.Cmp(response.DepositBalance) > 0 {
		usableBalance = response.DepositBalance
	} else {
		usableBalance = response.CreditBalance
	}
	if usableBalance.Cmp(totalAmountNeeded) > 0 {
		usableBalance = totalAmountNeeded
	}
	totalAmountSupplied := big.NewInt(0).Sub(totalAmountNeeded, usableBalance)
	// Update response
	response.CanDeposit = !(response.InsufficientBalance || response.InvalidAmount || response.DepositDisabled)
	if !response.CanDeposit {
		return &response, nil
	}

	if response.CanDeposit && !response.CanUseCredit && response.NodeBalance.Cmp(totalAmountSupplied) < 0 {
		// Can't use credit and there's not enough ETH in the node wallet to deposit so error out
		response.InsufficientBalanceWithoutCredit = true
		response.CanDeposit = false
	}

	// Break before the gas estimator if depositing won't work
	if !response.CanDeposit {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	opts.Value = totalAmountSupplied

	deposits := make([]rptypes.DepositData, numValidators)
	var usedExpressTickets uint32

	for i := uint64(0); i < numValidators; i++ {
		// Get the next validator key
		validatorKey, err := w.GetNextValidatorKey(uint(i))
		if err != nil {
			return nil, err
		}

		var withdrawalCredentials common.Hash

		if !saturnDeployed {
			// Get the next minipool address and withdrawal credentials
			minipoolAddress, err = minipool.GetExpectedAddress(rp, nodeAccount.Address, salt, nil)
			if err != nil {
				return nil, err
			}
			response.MinipoolAddress = minipoolAddress
			withdrawalCredentials, err = minipool.GetMinipoolWithdrawalCredentials(rp, minipoolAddress, nil)
			if err != nil {
				return nil, err
			}
		} else {
			// In case Saturn is deployed, the withdrawal credential will always be the Megapool

			// Get the megapool address
			megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}

			// calculte the withdrawal credentials (in case megapool is not deployed)
			withdrawalCredentials = services.CalculateMegapoolWithdrawalCredentials(megapoolAddress)

		}

		// Get validator deposit data and associated parameters
		depositAmount := uint64(1e9) // 1 ETH in gwei
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return nil, err
		}
		deposits[i].BondAmount = reducedBondWei
		deposits[i].DepositDataRoot = depositDataRoot
		deposits[i].ValidatorPubkey = depositData.PublicKey
		deposits[i].ValidatorSignature = depositData.Signature
		if usedExpressTickets < numExpressTickets {
			deposits[i].UseExpressTicket = true
			usedExpressTickets += 1
		}
		// Do a final sanity check
		err = validateDepositInfo(eth2Config, uint64(depositAmount), rptypes.BytesToValidatorPubkey(deposits[i].ValidatorPubkey), withdrawalCredentials, rptypes.BytesToValidatorSignature(deposits[i].ValidatorSignature))
		if err != nil {
			return nil, fmt.Errorf("Your deposit failed the validation safety check: %w\n"+
				"For your safety, this deposit will not be submitted and your ETH will not be staked.\n"+
				"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS and include the following information:\n"+
				"\tDomain Type: 0x%s\n"+
				"\tGenesis Fork Version: 0x%s\n"+
				"\tGenesis Validator Root: 0x%s\n"+
				"\tDeposit Amount: %d gwei\n"+
				"\tValidator Pubkey: %s\n"+
				"\tWithdrawal Credentials: %s\n"+
				"\tSignature: %s\n",
				err,
				hex.EncodeToString(eth2types.DomainDeposit[:]),
				hex.EncodeToString(eth2Config.GenesisForkVersion),
				hex.EncodeToString(eth2types.ZeroGenesisValidatorsRoot),
				depositAmount,
				deposits[i].ValidatorPubkey,
				withdrawalCredentials.Hex(),
				deposits[i].ValidatorSignature,
			)
		}
	}

	if !saturnDeployed {
		// Run the deposit gas estimator
		gasInfo, err := nodev131.EstimateDepositWithCreditGas(rp, deposits[0].BondAmount, minNodeFee, rptypes.BytesToValidatorPubkey(deposits[0].ValidatorPubkey), rptypes.BytesToValidatorSignature(deposits[0].ValidatorSignature), deposits[0].DepositDataRoot, salt, minipoolAddress, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	} else {
		// Run the deposit gas estimator
		gasInfo, err := node.EstimateDepositMultiGas(rp, deposits, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	}

	return &response, nil

}

func nodeDeposit(c *cli.Context, numValidators uint64, numExpressTickets uint32, minNodeFee float64, salt *big.Int, submit bool) (*api.NodeDepositResponse, error) {

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
	response := api.NodeDepositResponse{}

	if !saturnDeployed {
		// Adjust the salt
		if salt.Cmp(big.NewInt(0)) == 0 {
			nonce, err := ec.NonceAt(context.Background(), nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			salt.SetUint64(nonce)
		}
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

	// Data
	var wg1 errgroup.Group
	var depositPoolBalance *big.Int
	var nodeCredit *big.Int
	var reducedBond float64
	// Check credit balance
	wg1.Go(func() error {
		ethBalanceWei, err := node.GetNodeCreditAndBalance(rp, nodeAccount.Address, nil)
		if err == nil {
			nodeCredit = ethBalanceWei
		}
		return err
	})

	// Get deposit pool balance
	wg1.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(rp, nil)
		return err
	})

	if saturnDeployed {
		wg1.Go(func() error {
			reducedBond, err = protocol.GetReducedBond(rp, nil)
			if err != nil {
				return err
			}
			return nil
		})
	} else {
		reducedBond = houstonBondAmount
	}

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	reducedBondWei := eth.EthToWei(reducedBond)
	totalAmountNeeded := big.NewInt(0).Mul(reducedBondWei, big.NewInt(int64(numValidators)))

	var usableBalance *big.Int
	if nodeCredit.Cmp(depositPoolBalance) > 0 {
		usableBalance = depositPoolBalance
	} else {
		usableBalance = nodeCredit
	}
	if usableBalance.Cmp(totalAmountNeeded) > 0 {
		usableBalance = totalAmountNeeded
	}

	totalAmountSupplied := big.NewInt(0).Sub(totalAmountNeeded, usableBalance)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	opts.Value = totalAmountSupplied

	var minipoolAddress common.Address
	var usedExpressTickets uint32
	deposits := make([]rptypes.DepositData, numValidators)
	for i := uint64(0); i < numValidators; i++ {

		// Create and save a new validator key
		validatorKey, err := w.CreateValidatorKey()
		if err != nil {
			return nil, err
		}

		var withdrawalCredentials common.Hash

		if !saturnDeployed {
			// Get the next minipool address and withdrawal credentials
			minipoolAddress, err = minipool.GetExpectedAddress(rp, nodeAccount.Address, salt, nil)
			if err != nil {
				return nil, err
			}
			withdrawalCredentials, err = minipool.GetMinipoolWithdrawalCredentials(rp, minipoolAddress, nil)
			if err != nil {
				return nil, err
			}
		} else {
			// In case Saturn is deployed, the withdrawal credential will always be the Megapool

			// Get the megapool address
			megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}

			// Get the withdrawal credentials
			withdrawalCredentials = services.CalculateMegapoolWithdrawalCredentials(megapoolAddress)
		}

		// Get validator deposit data and associated parameters
		depositAmount := uint64(1e9) // 1 ETH in gwei
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return nil, err
		}
		deposits[i].BondAmount = reducedBondWei
		pubKey := depositData.PublicKey
		signature := depositData.Signature
		deposits[i].DepositDataRoot = depositDataRoot
		deposits[i].ValidatorPubkey = pubKey
		deposits[i].ValidatorSignature = signature
		if usedExpressTickets < numExpressTickets {
			deposits[i].UseExpressTicket = true
			usedExpressTickets += 1
		}

		validatorPubkey := rptypes.BytesToValidatorPubkey(pubKey)
		// Make sure a validator with this pubkey doesn't already exist
		status, err := bc.GetValidatorStatus(validatorPubkey, nil)
		if err != nil {
			return nil, fmt.Errorf("Error checking for existing validator status: %w\nYour funds have not been deposited for your own safety.", err)
		}
		if status.Exists {
			return nil, fmt.Errorf("**** ALERT ****\n"+
				"The following validator pubkey is already in use on the Beacon chain:\n\t%s\n"+
				"Rocket Pool will not allow you to deposit this validator for your own safety so you do not get slashed.\n"+
				"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.\n"+
				"***************\n", validatorPubkey.Hex())
		}

		validatorSignature := rptypes.BytesToValidatorSignature(signature)
		// Do a final sanity check
		err = validateDepositInfo(eth2Config, depositAmount, validatorPubkey, withdrawalCredentials, validatorSignature)
		if err != nil {
			return nil, fmt.Errorf("Your deposit failed the validation safety check: %w\n"+
				"For your safety, this deposit will not be submitted and your ETH will not be staked.\n"+
				"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS and include the following information:\n"+
				"\tDomain Type: 0x%s\n"+
				"\tGenesis Fork Version: 0x%s\n"+
				"\tGenesis Validator Root: 0x%s\n"+
				"\tDeposit Amount: %d gwei\n"+
				"\tValidator Pubkey: %s\n"+
				"\tWithdrawal Credentials: %s\n"+
				"\tSignature: %s\n",
				err,
				hex.EncodeToString(eth2types.DomainDeposit[:]),
				hex.EncodeToString(eth2Config.GenesisForkVersion),
				hex.EncodeToString(eth2types.ZeroGenesisValidatorsRoot),
				depositAmount,
				validatorPubkey.Hex(),
				withdrawalCredentials.Hex(),
				validatorSignature.Hex(),
			)
		}
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Do not send transaction unless requested
	opts.NoSend = !submit

	var tx *types.Transaction
	if !saturnDeployed {
		// Legacy Deposit
		pubkey := rptypes.BytesToValidatorPubkey(deposits[0].ValidatorPubkey)
		tx, err = nodev131.DepositWithCredit(rp, deposits[0].BondAmount, minNodeFee, pubkey, rptypes.BytesToValidatorSignature(deposits[0].ValidatorSignature), deposits[0].DepositDataRoot, salt, minipoolAddress, opts)
		if err != nil {
			return nil, err
		}
	} else {
		// Deposit
		tx, err = node.DepositMulti(rp, deposits, opts)
		if err != nil {
			return nil, err
		}
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
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
	response.MinipoolAddress = minipoolAddress
	response.DepositData = deposits

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
