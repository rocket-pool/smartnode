package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/signing"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/settings"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	prdeposit "github.com/prysmaticlabs/prysm/v3/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

type nodeDepositHandler struct {
	amountWei   *big.Int
	minNodeFee  float64
	salt        *big.Int
	depositPool *deposit.DepositPool
	pSettings   *settings.ProtocolDaoSettings
	oSettings   *settings.OracleDaoSettings
	mpMgr       *minipool.IMinipoolManager
}

func (h *nodeDepositHandler) CreateBindings(ctx *callContext) error {
	rp := ctx.rp
	var err error

	h.depositPool, err = deposit.NewDepositPool(rp)
	if err != nil {
		return fmt.Errorf("error getting deposit pool binding: %w", err)
	}
	h.pSettings, err = settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error getting pDAO settings binding: %w", err)
	}
	h.oSettings, err = settings.NewOracleDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error getting oDAO settings binding: %w", err)
	}
	h.mpMgr, err = minipool.NewMinipoolManager(rp)
	if err != nil {
		return fmt.Errorf("error getting minipool manager binding: %w", err)
	}
	return nil
}

func (h *nodeDepositHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	node := ctx.node
	node.GetDepositCredit(mc)
	h.depositPool.GetBalance(mc)
	h.pSettings.GetNodeDepositEnabled(mc)
	h.oSettings.GetScrubPeriod(mc)
}

func (h *nodeDepositHandler) PrepareResponse(ctx *callContext, response *api.NodeDepositResponse) error {
	rp := ctx.rp
	bc := ctx.bc
	node := ctx.node
	cfg := ctx.cfg
	opts := ctx.opts
	w := ctx.w

	// Initial population
	response.CreditBalance = node.Credit
	response.DepositDisabled = !h.pSettings.Node.IsDepositingEnabled
	response.DepositBalance = h.depositPool.Balance
	response.ScrubPeriod = h.oSettings.Minipools.ScrubPeriod.Formatted()

	// Get Beacon config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return fmt.Errorf("error getting Beacon config: %w", err)
	}

	// Adjust the salt
	if h.salt.Cmp(big.NewInt(0)) == 0 {
		nonce, err := rp.Client.NonceAt(context.Background(), node.Address, nil)
		if err != nil {
			return fmt.Errorf("error getting node's latest nonce: %w", err)
		}
		h.salt.SetUint64(nonce)
	}

	// Check node balance
	response.NodeBalance, err = rp.Client.BalanceAt(context.Background(), node.Address, nil)
	if err != nil {
		return fmt.Errorf("error getting node's ETH balance: %w", err)
	}

	// Check the node's collateral
	collateral, err := rputils.CheckCollateral(rp, node.Address, nil)
	if err != nil {
		return fmt.Errorf("error checking node collateral: %w", err)
	}
	ethMatched := collateral.EthMatched
	ethMatchedLimit := collateral.EthMatchedLimit
	pendingMatchAmount := collateral.PendingMatchAmount

	// Check for insufficient balance
	totalBalance := big.NewInt(0).Add(response.NodeBalance, response.CreditBalance)
	response.InsufficientBalance = (h.amountWei.Cmp(totalBalance) > 0)

	// Check if the credit balance can be used
	response.CanUseCredit = (response.DepositBalance.Cmp(eth.EthToWei(1)) >= 0)

	// Check data
	validatorEthWei := eth.EthToWei(ValidatorEth)
	matchRequest := big.NewInt(0).Sub(validatorEthWei, h.amountWei)
	availableToMatch := big.NewInt(0).Sub(ethMatchedLimit, ethMatched)
	availableToMatch.Sub(availableToMatch, pendingMatchAmount)
	response.InsufficientRplStake = (availableToMatch.Cmp(matchRequest) == -1)

	// Update response
	response.CanDeposit = !(response.InsufficientBalance || response.InsufficientRplStake || response.InvalidAmount || response.DepositDisabled)
	if response.CanDeposit && !response.CanUseCredit && response.NodeBalance.Cmp(h.amountWei) < 0 {
		// Can't use credit and there's not enough ETH in the node wallet to deposit so error out
		response.InsufficientBalanceWithoutCredit = true
		response.CanDeposit = false
	}

	// Return if depositing won't work
	if !response.CanDeposit {
		return nil
	}

	// Make sure ETH2 is on the correct chain
	depositContractInfo, err := rputils.GetDepositContractInfo(rp, cfg, bc)
	if err != nil {
		return fmt.Errorf("error verifying the EL and BC are on the same chain: %w", err)
	}
	if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
		depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
		return fmt.Errorf("FATAL: Beacon network mismatch! Expected %s on chain %d, but beacon is using %s on chain %d.",
			depositContractInfo.RPDepositContract.Hex(),
			depositContractInfo.RPNetwork,
			depositContractInfo.BeaconDepositContract.Hex(),
			depositContractInfo.BeaconNetwork)
	}

	// Get how much credit to use
	if response.CanUseCredit {
		remainingAmount := big.NewInt(0).Sub(h.amountWei, response.CreditBalance)
		if remainingAmount.Cmp(big.NewInt(0)) > 0 {
			// Send the remaining amount if the credit isn't enough to cover the whole deposit
			opts.Value = remainingAmount
		}
	} else {
		opts.Value = h.amountWei
	}

	// Get the next available validator key without saving it
	validatorKey, err := w.GetNextValidatorKey()
	if err != nil {
		return fmt.Errorf("error getting next available validator key: %w", err)
	}

	// Get the next minipool address
	var minipoolAddress common.Address
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetExpectedMinipoolAddress(mc, &minipoolAddress, h.salt)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting expected minipool address: %w", err)
	}
	response.MinipoolAddress = minipoolAddress

	// Get the withdrawal credentials
	var withdrawalCredentials common.Hash
	err = rp.Query(func(mc *batch.MultiCaller) error {
		h.mpMgr.GetMinipoolWithdrawalCredentials(mc, &withdrawalCredentials, minipoolAddress)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool withdrawal credentials: %w", err)
	}

	// Get validator deposit data and associated parameters
	depositAmount := uint64(1e9) // 1 ETH in gwei
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
	if err != nil {
		return fmt.Errorf("error getting deposit data: %w", err)
	}
	pubkey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)
	response.ValidatorPubkey = pubkey

	// Make sure a validator with this pubkey doesn't already exist
	status, err := bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return fmt.Errorf("Error checking for existing validator status: %w\nYour funds have not been deposited for your own safety.", err)
	}
	if status.Exists {
		return fmt.Errorf("**** ALERT ****\n"+
			"Your minipool %s has the following as a validator pubkey:\n\t%s\n"+
			"This key is already in use by validator %d on the Beacon chain!\n"+
			"Rocket Pool will not allow you to deposit this validator for your own safety so you do not get slashed.\n"+
			"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.\n"+
			"***************\n", minipoolAddress.Hex(), pubkey.Hex(), status.Index)
	}

	// Do a final sanity check
	err = validateDepositInfo(eth2Config, uint64(depositAmount), pubkey, withdrawalCredentials, signature)
	if err != nil {
		return fmt.Errorf("FATAL: Your deposit failed the validation safety check: %w\n"+
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
			pubkey.Hex(),
			withdrawalCredentials.Hex(),
			signature.Hex(),
		)
	}

	// Get tx info
	var txInfo *core.TransactionInfo
	var funcName string
	if response.CanUseCredit {
		txInfo, err = node.DepositWithCredit(h.amountWei, h.minNodeFee, pubkey, signature, depositDataRoot, h.salt, minipoolAddress, opts)
		funcName = "DepositWithCredit"
	} else {
		txInfo, err = node.Deposit(h.amountWei, h.minNodeFee, pubkey, signature, depositDataRoot, h.salt, minipoolAddress, opts)
		funcName = "Deposit"
	}
	if err != nil {
		return fmt.Errorf("error getting TX info for %s: %w", funcName, err)
	}
	response.TxInfo = txInfo

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
