package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/urfave/cli/v3"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	rputils "github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Dissolve timed out minipools task
type dissolveInvalidCredentialsPendingDeposits struct {
	c   *cli.Command
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
	bc  *services.BeaconClientManager
	it  *searchData
}

type searchData struct {
	stateBlockTime   time.Time
	startBlock       *big.Int
	eventLogInterval *big.Int
	depositDomain    []byte
}

// Create dissolve invalid credentials pending deposits task
func newDissolveInvalidCredentialsPendingDeposits(c *cli.Command, logger log.ColorLogger) (*dissolveInvalidCredentialsPendingDeposits, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	// Get the beacon client
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &dissolveInvalidCredentialsPendingDeposits{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		rp:  rp,
		bc:  bc,
	}, nil

}

// Dissolve timed out megapool validators
func (t *dissolveInvalidCredentialsPendingDeposits) run(state *state.NetworkState) error {
	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	// Log
	t.log.Println("Checking for invalid info on pending validator deposits ...")

	// Dissolve validators
	err := t.dissolveInvalidCredentialPendingDeposits(state)
	if err != nil {
		return err
	}

	return nil
}

// Get megapool validators that can be dissolved due to using invalid credentials
func (t *dissolveInvalidCredentialsPendingDeposits) dissolveInvalidCredentialPendingDeposits(state *state.NetworkState) error {

	// Prepare the deposit-contract search parameters used by verifyDeposits
	t.it = &searchData{}
	if err := t.getEth1SearchArtifacts(state); err != nil {
		return fmt.Errorf("error preparing deposit search artifacts: %w", err)
	}

	validatorsToCheck := map[types.ValidatorPubkey]megapool.ValidatorInfoFromGlobalIndex{}
	for _, validator := range state.MegapoolValidatorGlobalIndex {
		if validator.ValidatorInfo.InPrestake {
			pubkey := types.ValidatorPubkey(validator.Pubkey)
			// Fetch the validator from the beacon state to compare credentials
			validatorFromState, err := t.bc.GetValidatorStatus(pubkey, nil)
			if err != nil {
				t.log.Printlnf("error getting the beacon state for validator 0x%s on megapool %s: %s", pubkey.Hex(), validator.MegapoolAddress.Hex(), err)
				continue
			}
			if !validatorFromState.Exists {
				t.log.Printlnf("Validator 0x%s does not exist on the beacon state. Searching for a pending deposit", pubkey.Hex())
			}

			// Track this validator so its deposit(s) get checked
			validatorsToCheck[pubkey] = validator

			// if !bytes.Equal(validatorFromState.WithdrawalCredentials.Bytes(), expectedWithdrawalAddress.Bytes()) {
			// 	t.log.Printlnf("Validator %s has an invalid credential %s while the expected is %s. Dissolving...", validatorFromState.Index, validatorFromState.WithdrawalCredentials, expectedWithdrawalAddress.Bytes())
			// 	t.dissolveMegapoolValidator(validator)
			// }
			// // Withdrawable epoch should be FAR_FUTURE_EPOCH
			// if validatorFromState.WithdrawableEpoch != FarFutureEpoch {
			// 	t.log.Printlnf("Validator %s has a withdrawable epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.WithdrawableEpoch, FarFutureEpoch)
			// 	t.dissolveMegapoolValidator(validator)
			// }
			// // Exit epoch should be FAR_FUTURE_EPOCH
			// if validatorFromState.ExitEpoch != FarFutureEpoch {
			// 	t.log.Printlnf("Validator %s has an exit epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.ExitEpoch, FarFutureEpoch)
			// 	t.dissolveMegapoolValidator(validator)
			// }
			// // Slashed should be false
			// if validatorFromState.Slashed {
			// 	t.log.Printlnf("Validator %s is slashed while the expected is false. Dissolving...", validatorFromState.Index)
			// 	t.dissolveMegapoolValidator(validator)
			// }
			// // Effective balance should be less than 32 ETH
			// if validatorFromState.EffectiveBalance >= 32000000000 {
			// 	t.log.Printlnf("Validator %s has an effective balance of %d while the expected is less than 32 ETH. Dissolving...", validatorFromState.Index, validatorFromState.EffectiveBalance)
			// 	t.dissolveMegapoolValidator(validator)
			// }
			// // Activation eligibility epoch should be FAR_FUTURE_EPOCH
			// if validatorFromState.ActivationEligibilityEpoch != FarFutureEpoch {
			// 	t.log.Printlnf("Validator %s has an activation eligibility epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.ActivationEligibilityEpoch, FarFutureEpoch)
			// 	t.dissolveMegapoolValidator(validator)
			// }
			// // Activation epoch should be FAR_FUTURE_EPOCH
			// if validatorFromState.ActivationEpoch != FarFutureEpoch {
			// 	t.log.Printlnf("Validator %s has an activation epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.ActivationEpoch, FarFutureEpoch)
			// 	t.dissolveMegapoolValidator(validator)
			// }
		}

	}

	// Verify the deposits
	err := t.verifyDeposits(validatorsToCheck)
	if err != nil {
		return err
	}

	return nil
}

// verifyDeposits inspects the beacon deposit contract for every supplied
// megapool validator and dissolves any whose first valid deposit was made with
// withdrawal credentials that don't match its megapool's expected credentials.
func (t *dissolveInvalidCredentialsPendingDeposits) verifyDeposits(validators map[types.ValidatorPubkey]megapool.ValidatorInfoFromGlobalIndex) error {

	if len(validators) == 0 {
		return nil
	}

	// Build the pubkey set expected by GetDeposits
	pubkeysMap := make(map[types.ValidatorPubkey]bool, len(validators))
	for pubkey := range validators {
		pubkeysMap[pubkey] = true
	}

	// Pull all matching deposits from the beacon deposit contract
	depositMap, err := rputils.GetDeposits(t.rp, pubkeysMap, t.it.startBlock, t.it.eventLogInterval, nil)
	if err != nil {
		return fmt.Errorf("error retrieving beacon deposits: %w", err)
	}

	// Collect every validator whose first valid deposit has the wrong
	// withdrawal credentials so they can all be dissolved together at the end.
	validatorsToDissolve := []megapool.ValidatorInfoFromGlobalIndex{}

	// Check each megapool validator's deposit data
	for pubkey, validator := range validators {

		deposits, exists := depositMap[pubkey]
		if !exists || len(deposits) == 0 {
			// No deposit has landed in the deposit contract yet; nothing to validate.
			t.log.Printlnf("No deposit found in the deposit contract for megapool validator %d (pubkey 0x%s) on megapool %s", validator.ValidatorId, pubkey.Hex(), validator.MegapoolAddress.Hex())
			continue
		}

		expectedCreds := services.CalculateMegapoolWithdrawalCredentials(validator.MegapoolAddress)

		// Find the first deposit with a valid signature - that's the one the
		// beacon chain will accept and thus pin the validator's credentials.
		for depositIndex, deposit := range deposits {
			depositData := new(ethpb.Deposit_Data)
			depositData.Amount = deposit.Amount
			depositData.PublicKey = deposit.Pubkey.Bytes()
			depositData.WithdrawalCredentials = deposit.WithdrawalCredentials.Bytes()
			depositData.Signature = deposit.Signature.Bytes()

			if err := prdeposit.VerifyDepositSignature(depositData, t.it.depositDomain); err != nil {
				// Invalid signature, skip and try the next deposit for this pubkey
				t.log.Printlnf("Invalid deposit signature for megapool validator %d (pubkey 0x%s):", validator.ValidatorId, pubkey.Hex())
				t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
				t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
				t.log.Printlnf("\tError: %s", err.Error())
				continue
			}

			actualCreds := deposit.WithdrawalCredentials
			if actualCreds != expectedCreds {
				t.log.Println("=== INVALID WITHDRAWAL CREDENTIALS DETECTED ON DEPOSIT CONTRACT ===")
				t.log.Printlnf("\tMegapool:       %s", validator.MegapoolAddress.Hex())
				t.log.Printlnf("\tValidator ID:   %d", validator.ValidatorId)
				t.log.Printlnf("\tPubkey:         0x%s", pubkey.Hex())
				t.log.Printlnf("\tTX Hash:        %s", deposit.TxHash.Hex())
				t.log.Printlnf("\tBlock:          %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
				t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
				t.log.Printlnf("\tActual creds:   %s", actualCreds.Hex())
				t.log.Println("===================================================================")

				validatorsToDissolve = append(validatorsToDissolve, validator)
			}

			// Stop at the first valid deposit - subsequent ones can't change
			// the validator's effective withdrawal credentials.
			break
		}
	}

	t.dissolveMegapoolValidators(validatorsToDissolve)

	return nil
}

// dissolveMegapoolValidators dissolves every validator in the supplied slice
func (t *dissolveInvalidCredentialsPendingDeposits) dissolveMegapoolValidators(validators []megapool.ValidatorInfoFromGlobalIndex) {
	if len(validators) == 0 {
		return
	}
	t.log.Printlnf("Dissolving %d megapool validator(s) with invalid withdrawal credentials...", len(validators))
	for _, validator := range validators {
		t.dissolveMegapoolValidator(validator)
	}
}

func (t *dissolveInvalidCredentialsPendingDeposits) dissolveMegapoolValidator(validator megapool.ValidatorInfoFromGlobalIndex) {
	// Log
	t.log.Printlnf("Dissolving megapool validator ID: %d from megapool %s...", validator.ValidatorId, validator.MegapoolAddress)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		t.log.Printlnf("error getting the node account transactor: %v", err)
		return
	}

	eth2Config, err := t.bc.GetEth2Config()
	if err != nil {
		t.log.Printlnf("error getting the eth2 config: %v", err)
		return
	}

	// Build a pending-deposit proof: the validator's first deposit is still
	// sitting in the beacon state's pending_deposits queue (it has no beacon
	// index yet), so we can't build a ValidatorProof for it.
	pendingDepositProof, err := services.GetPendingDepositProof(t.c, types.ValidatorPubkey(validator.Pubkey), nil)
	if err != nil {
		t.log.Printlnf("error getting pending deposit proof: %v", err)
		return
	}
	slotTimestamp := uint64(eth2Config.GetSlotTime(pendingDepositProof.Slot).Unix())

	// Get the gas limit
	gasInfo, err := megapool.EstimateDissolveWithPendingDepositProofGas(t.rp, validator.MegapoolAddress, validator.ValidatorId, slotTimestamp, pendingDepositProof, opts)
	if err != nil {
		t.log.Printlnf("error estimating the gas required to dissolve the validator: %v", err)
		return
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
		return
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = gasInfo.SafeGasLimit

	// Dissolve
	tx, err := megapool.DissolveWithPendingDepositProof(t.rp, validator.MegapoolAddress, validator.ValidatorId, slotTimestamp, pendingDepositProof, opts)
	if err != nil {
		t.log.Printlnf("error dissolving the validator: %v", err)
		return
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		t.log.Printlnf("error waiting for the transaction to be included in a block: %v", err)
		return
	}

	// Log
	t.log.Printlnf("Successfully dissolved megapool validator ID: %s from megapool %s. (Invalid credentials)", validator.ValidatorId, validator.MegapoolAddress)
}

// Get various elements needed to do eth1 prestake and deposit contract searches
func (t *dissolveInvalidCredentialsPendingDeposits) getEth1SearchArtifacts(state *state.NetworkState) error {

	// Get the time of the state's EL block
	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
	t.it.stateBlockTime = genesisTime.Add(secondsSinceGenesis)

	// Get the block to start searching the deposit contract from
	stateBlockNumber := big.NewInt(0).SetUint64(state.ElBlockNumber)
	offset := big.NewInt(BlockStartOffset)
	if stateBlockNumber.Cmp(offset) < 0 {
		offset = stateBlockNumber // Deal with chains that are younger than the look-behind interval
	}
	targetBlockNumber := big.NewInt(0).Sub(stateBlockNumber, offset)
	targetBlock, err := t.ec.HeaderByNumber(context.Background(), targetBlockNumber)
	if err != nil {
		return fmt.Errorf("error getting header for EL block %d: %w", targetBlockNumber, err)
	}
	t.it.startBlock = targetBlock.Number

	// Check the prestake event from the minipool and validate its signature
	eventLogInterval, err := t.cfg.GetEventLogInterval()
	if err != nil {
		return fmt.Errorf("error getting event log interval %w", err)
	}
	t.it.eventLogInterval = big.NewInt(int64(eventLogInterval))

	// Put together the signature validation data
	eth2Config := state.BeaconConfig
	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return fmt.Errorf("error computing deposit domain: %w", err)
	}
	t.it.depositDomain = depositDomain

	return nil

}
