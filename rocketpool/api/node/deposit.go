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
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	prestakeDepositAmount float64 = 1.0
	ValidatorEth          float64 = 32.0
)

func canNodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64, salt *big.Int) (*api.CanNodeDepositResponse, error) {

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

	// Check for Houston
	isHoustonDeployed, err := state.IsHoustonDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}

	// Response
	response := api.CanNodeDepositResponse{}

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
	var pendingMatchAmount *big.Int
	var minipoolAddress common.Address
	var depositPoolBalance *big.Int

	// Check credit balance
	if isHoustonDeployed {
		wg1.Go(func() error {
			ethBalanceWei, err := node.GetNodeUsableCreditAndBalance(rp, nodeAccount.Address, nil)
			if err == nil {
				response.CreditBalance = ethBalanceWei
			}
			return err
		})
	} else {
		wg1.Go(func() error {
			ethBalanceWei, err := node.GetNodeDepositCredit(rp, nodeAccount.Address, nil)
			if err == nil {
				response.CreditBalance = ethBalanceWei
			}
			return err
		})
	}

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

	// Get node staking information
	wg1.Go(func() error {
		ethMatched, ethMatchedLimit, pendingMatchAmount, err = rputils.CheckCollateral(rp, nodeAccount.Address, nil)
		if err != nil {
			return fmt.Errorf("error checking collateral for node %s: %w", nodeAccount.Address.Hex(), err)
		}
		return nil
	})

	// Get deposit pool balance
	wg1.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(rp, nil)
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	// Check for insufficient balance
	totalBalance := big.NewInt(0).Add(response.NodeBalance, response.CreditBalance)
	response.InsufficientBalance = (amountWei.Cmp(totalBalance) > 0)

	// Check if the credit balance can be used
	response.DepositBalance = depositPoolBalance
	response.CanUseCredit = (depositPoolBalance.Cmp(eth.EthToWei(1)) >= 0)

	// Check data
	validatorEthWei := eth.EthToWei(ValidatorEth)
	matchRequest := big.NewInt(0).Sub(validatorEthWei, amountWei)
	availableToMatch := big.NewInt(0).Sub(ethMatchedLimit, ethMatched)
	availableToMatch.Sub(availableToMatch, pendingMatchAmount)
	response.InsufficientRplStake = (availableToMatch.Cmp(matchRequest) == -1)

	// Update response
	response.CanDeposit = !(response.InsufficientBalance || response.InsufficientRplStake || response.InvalidAmount || response.DepositDisabled)
	if !response.CanDeposit {
		return &response, nil
	}

	if response.CanDeposit && !response.CanUseCredit && response.NodeBalance.Cmp(amountWei) < 0 {
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

	// Get how much credit to use
	if response.CanUseCredit {
		remainingAmount := big.NewInt(0).Sub(amountWei, response.CreditBalance)
		if remainingAmount.Cmp(big.NewInt(0)) > 0 {
			// Send the remaining amount if the credit isn't enough to cover the whole deposit
			opts.Value = remainingAmount
		}
	} else {
		opts.Value = amountWei
	}

	// Get the next validator key
	validatorKey, err := w.GetNextValidatorKey()
	if err != nil {
		return nil, err
	}

	// Get the next minipool address and withdrawal credentials
	minipoolAddress, err = minipool.GetExpectedAddress(rp, nodeAccount.Address, salt, nil)
	if err != nil {
		return nil, err
	}
	response.MinipoolAddress = minipoolAddress
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get validator deposit data and associated parameters
	depositAmount := uint64(1e9) // 1 ETH in gwei
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
	if err != nil {
		return nil, err
	}
	pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)

	// Do a final sanity check
	err = validateDepositInfo(eth2Config, uint64(depositAmount), pubKey, withdrawalCredentials, signature)
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
			pubKey.Hex(),
			withdrawalCredentials.Hex(),
			signature.Hex(),
		)
	}

	// Run the deposit gas estimator
	if response.CanUseCredit {
		gasInfo, err := node.EstimateDepositWithCreditGas(rp, amountWei, minNodeFee, pubKey, signature, depositDataRoot, salt, minipoolAddress, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	} else {
		gasInfo, err := node.EstimateDepositGas(rp, amountWei, minNodeFee, pubKey, signature, depositDataRoot, salt, minipoolAddress, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	}

	return &response, nil

}

func nodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64, salt *big.Int, useCreditBalance bool, submit bool) (*api.NodeDepositResponse, error) {

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

	// Check for Houston
	isHoustonDeployed, err := state.IsHoustonDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}

	// Response
	response := api.NodeDepositResponse{}

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
	scrubPeriodUnix, err := trustednode.GetScrubPeriod(rp, nil)
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
	var creditBalanceWei *big.Int
	if isHoustonDeployed {
		// Get the node's credit and ETH staked on behalf balance
		creditBalanceWei, err = node.GetNodeUsableCreditAndBalance(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}
	} else {
		creditBalanceWei, err = node.GetNodeDepositCredit(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}
	}

	// Get how much credit to use
	if useCreditBalance {
		remainingAmount := big.NewInt(0).Sub(amountWei, creditBalanceWei)
		if remainingAmount.Cmp(big.NewInt(0)) > 0 {
			// Send the remaining amount if the credit isn't enough to cover the whole deposit
			opts.Value = remainingAmount
		}
	} else {
		opts.Value = amountWei
	}

	// Create and save a new validator key
	validatorKey, err := w.CreateValidatorKey()
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

	// Get validator deposit data and associated parameters
	depositAmount := uint64(1e9) // 1 ETH in gwei
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
	if err != nil {
		return nil, err
	}
	pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)

	// Make sure a validator with this pubkey doesn't already exist
	status, err := bc.GetValidatorStatus(pubKey, nil)
	if err != nil {
		return nil, fmt.Errorf("Error checking for existing validator status: %w\nYour funds have not been deposited for your own safety.", err)
	}
	if status.Exists {
		return nil, fmt.Errorf("**** ALERT ****\n"+
			"Your minipool %s has the following as a validator pubkey:\n\t%s\n"+
			"This key is already in use by validator %d on the Beacon chain!\n"+
			"Rocket Pool will not allow you to deposit this validator for your own safety so you do not get slashed.\n"+
			"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.\n"+
			"***************\n", minipoolAddress.Hex(), pubKey.Hex(), status.Index)
	}

	// Do a final sanity check
	err = validateDepositInfo(eth2Config, depositAmount, pubKey, withdrawalCredentials, signature)
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
			pubKey.Hex(),
			withdrawalCredentials.Hex(),
			signature.Hex(),
		)
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Do not send transaction unless requested
	opts.NoSend = !submit

	// Deposit
	var tx *types.Transaction
	if useCreditBalance {
		tx, err = node.DepositWithCredit(rp, amountWei, minNodeFee, pubKey, signature, depositDataRoot, salt, minipoolAddress, opts)
	} else {
		tx, err = node.Deposit(rp, amountWei, minNodeFee, pubKey, signature, depositDataRoot, salt, minipoolAddress, opts)
	}
	if err != nil {
		return nil, err
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
	response.ValidatorPubkey = pubKey

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
