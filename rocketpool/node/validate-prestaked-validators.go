package node

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/urfave/cli/v3"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

const (
	// prestakeAmountGwei is the 1 ETH deposit checked by challengeInvalidSignature.
	prestakeAmountGwei uint64 = 1_000_000_000
	// prestakeIndexPageSize is how many global indices one getPrestakedValidatorIndices call inspects.
	prestakeIndexPageSize int64 = 200
	// maxPrestakeChallengesPerCycle keeps one pass from holding the node loop on a large backlog.
	maxPrestakeChallengesPerCycle      = 8
	maxPendingDepositDissolvesPerCycle = 8
)

type prestakeValidation int

const (
	prestakeSkipped prestakeValidation = iota
	prestakeValid
	prestakeInvalidSignature
)

// Validate prestaked validators task
type validatePrestakedValidators struct {
	c              *cli.Command
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create validate prestaked validators task
func newValidatePrestakedValidators(c *cli.Command, logger log.ColorLogger) (*validatePrestakedValidators, error) {
	cfg, err := services.GetConfig(c)
	if err != nil {
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

	gas := loadAutoTxGas(cfg, &logger)
	return &validatePrestakedValidators{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		gasThreshold:   gas.thresholdGwei,
		maxFee:         gas.maxFee,
		maxPriorityFee: gas.maxPriorityFee,
		gasLimit:       0,
	}, nil
}

// Check prestaked validators, challenge invalid deposit signatures, and dissolve
// validators whose pending beacon deposit is signed to the wrong withdrawal credentials.
func (t *validatePrestakedValidators) run(state *state.NetworkStateIndex) error {
	if !state.Saturn2Deployed {
		return nil
	}

	t.log.Println("Checking prestaked megapool validators...")

	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}
	opts := &bind.CallOpts{
		BlockNumber: new(big.Int).SetUint64(state.ElBlockNumber),
	}
	// challengeInvalidSignature and dissolvePendingDeposit require the caller to have a deployed megapool.
	canChallenge, err := megapool.GetMegapoolDeployed(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return err
	}

	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, state.BeaconConfig.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return fmt.Errorf("error computing deposit domain: %w", err)
	}

	indices, failed, err := t.prestakedIndices(opts)
	if err != nil {
		return err
	}
	indices = append(indices, failed...)

	var prestaked []megapool.ValidatorPrestakeInfo
	var invalid []megapool.ValidatorPrestakeInfo
	for _, index := range indices {
		info, err := megapool.GetValidatorPrestakeInfo(t.rp, new(big.Int).SetUint64(index), opts)
		if err != nil {
			t.log.Printlnf("Error reading prestake info for global index %d: %v", index, err)
			continue
		}
		if !info.ValidatorInfo.InPrestake {
			continue
		}
		prestaked = append(prestaked, info)
		credentials := services.CalculateMegapoolWithdrawalCredentials(info.MegapoolAddress)
		valid := prestakeSignatureValid(info.Pubkey, info.Signature, credentials, depositDomain)
		if classifyPrestake(info, valid) == prestakeInvalidSignature {
			t.log.Printlnf("Invalid prestake signature for megapool %s validator %d (global index %d)", info.MegapoolAddress.Hex(), info.ValidatorId, index)
			invalid = append(invalid, info)
		}
	}
	if !canChallenge {
		if len(invalid) > 0 {
			t.log.Printlnf("%d prestaked validators have invalid signatures. This node cannot challenge until its megapool is deployed.", len(invalid))
		}
		return nil
	}
	if err := t.challengeInvalid(invalid); err != nil {
		return err
	}
	return t.dissolveWrongCredentialDeposits(prestaked, depositDomain)
}

func (t *validatePrestakedValidators) prestakedIndices(opts *bind.CallOpts) (indices []uint64, failed []uint64, err error) {
	offset := big.NewInt(0)
	limit := big.NewInt(prestakeIndexPageSize)
	for {
		page, err := megapool.GetPrestakedValidatorIndices(t.rp, offset, limit, opts)
		if err != nil {
			return nil, nil, err
		}
		for _, index := range page.Indices {
			value, err := uint256ToUint64(index)
			if err != nil {
				return nil, nil, fmt.Errorf("prestaked index: %w", err)
			}
			indices = append(indices, value)
		}
		for _, index := range page.FailedIndices {
			value, err := uint256ToUint64(index)
			if err != nil {
				return nil, nil, fmt.Errorf("failed prestaked index: %w", err)
			}
			failed = append(failed, value)
		}
		if !prestakePageAdvances(offset, page.NextIndex) {
			return indices, failed, nil
		}
		offset = page.NextIndex
	}
}

func (t *validatePrestakedValidators) challengeInvalid(invalid []megapool.ValidatorPrestakeInfo) error {
	submitted := 0
	for _, info := range invalid {
		if submitted >= maxPrestakeChallengesPerCycle {
			t.log.Printlnf("%d invalid prestake signatures remain for the next cycle", len(invalid)-submitted)
			return nil
		}
		sent, err := t.challenge(info)
		if err != nil {
			t.log.Printlnf("Error challenging megapool %s validator %d: %v", info.MegapoolAddress.Hex(), info.ValidatorId, err)
			continue
		}
		if !sent {
			return nil
		}
		submitted++
	}
	return nil
}

func (t *validatePrestakedValidators) challenge(info megapool.ValidatorPrestakeInfo) (bool, error) {
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}
	gasLimits, err := megapool.EstimateChallengeInvalidSignatureGas(t.rp, info.MegapoolAddress, info.ValidatorId, opts)
	if err != nil {
		return false, err
	}
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Sign() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return false, err
		}
	}
	if !gasLimits.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return false, nil
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gasLimits.Safe

	tx, err := megapool.ChallengeInvalidSignature(t.rp, info.MegapoolAddress, info.ValidatorId, opts)
	if err != nil {
		return false, err
	}
	if err := transactions.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log); err != nil {
		return false, err
	}
	t.log.Printlnf("Challenged invalid prestake signature for megapool %s validator %d", info.MegapoolAddress.Hex(), info.ValidatorId)
	return true, nil
}

type pendingDepositDissolve struct {
	info         megapool.ValidatorPrestakeInfo
	depositIndex uint64
}

func (t *validatePrestakedValidators) dissolveWrongCredentialDeposits(prestaked []megapool.ValidatorPrestakeInfo, depositDomain []byte) error {
	if len(prestaked) == 0 {
		return nil
	}
	beaconState, slotTimestamp, err := services.GetHeadBeaconState(t.bc, t.rp.Client)
	if err != nil {
		return err
	}
	matches := wrongCredentialPendingDeposits(beaconState.GetPendingDeposits(), prestaked, depositDomain)
	if len(matches) == 0 {
		return nil
	}
	limit := maxPendingDepositDissolvesPerCycle
	if len(matches) < limit {
		limit = len(matches)
	} else {
		t.log.Printlnf("%d pending deposits with invalid withdrawal credentials remain for the next cycle", len(matches)-limit)
	}
	batch := matches[:limit]
	indices := make([]uint64, len(batch))
	for i, match := range batch {
		indices[i] = match.depositIndex
		t.log.Printlnf("Pending deposit %d for megapool %s validator %d uses invalid withdrawal credentials", match.depositIndex, match.info.MegapoolAddress.Hex(), match.info.ValidatorId)
	}
	proofs, slotProof, err := services.PendingDepositProofsFromState(beaconState, indices)
	if err != nil {
		if errors.Is(err, generic.ErrPendingDepositProofUnsupported) {
			t.log.Printlnf("%d prestaked validators have a valid pending deposit to the wrong withdrawal credentials. Dissolving them requires a Gloas beacon state.", len(matches))
			return nil
		}
		return err
	}
	for i, match := range batch {
		sent, err := t.dissolvePendingDeposit(match.info, proofs[i], slotProof, slotTimestamp)
		if err != nil {
			t.log.Printlnf("Error dissolving megapool %s validator %d: %v", match.info.MegapoolAddress.Hex(), match.info.ValidatorId, err)
			continue
		}
		if !sent {
			return nil
		}
	}
	return nil
}

func (t *validatePrestakedValidators) dissolvePendingDeposit(info megapool.ValidatorPrestakeInfo, depositProof megapool.PendingDepositProof, slotProof megapool.SlotProof, slotTimestamp uint64) (bool, error) {
	proofData, err := megapool.EncodePendingDepositProofBundleV1(depositProof, slotProof)
	if err != nil {
		return false, err
	}
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}
	gasLimits, err := megapool.EstimateDissolvePendingDepositGas(t.rp, info.MegapoolAddress, info.ValidatorId, slotTimestamp, megapool.PendingDepositProofVersion1, proofData, opts)
	if err != nil {
		return false, err
	}
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Sign() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return false, err
		}
	}
	if !gasLimits.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return false, nil
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gasLimits.Safe

	tx, err := megapool.DissolvePendingDeposit(t.rp, info.MegapoolAddress, info.ValidatorId, slotTimestamp, megapool.PendingDepositProofVersion1, proofData, opts)
	if err != nil {
		return false, err
	}
	if err := transactions.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log); err != nil {
		return false, err
	}
	t.log.Printlnf("Dissolved megapool %s validator %d for a pending deposit with invalid withdrawal credentials", info.MegapoolAddress.Hex(), info.ValidatorId)
	return true, nil
}

// wrongCredentialPendingDeposits returns prestaked validators whose pubkey appears
// in a pending deposit that is validly signed to withdrawal credentials other than
// that validator's megapool. One deposit dissolves a validator, so later deposits
// for the same validator are skipped.
func wrongCredentialPendingDeposits(deposits []*generic.PendingDeposit, prestaked []megapool.ValidatorPrestakeInfo, depositDomain []byte) []pendingDepositDissolve {
	byPubkey := make(map[string][]megapool.ValidatorPrestakeInfo, len(prestaked))
	for _, info := range prestaked {
		if !info.ValidatorInfo.InPrestake || len(info.Pubkey) == 0 {
			continue
		}
		byPubkey[string(info.Pubkey)] = append(byPubkey[string(info.Pubkey)], info)
	}
	seen := make(map[string]struct{})
	var matches []pendingDepositDissolve
	for index, deposit := range deposits {
		if deposit == nil || len(deposit.Pubkey) == 0 || !pendingDepositSignatureValid(deposit, depositDomain) {
			continue
		}
		for _, info := range byPubkey[string(deposit.Pubkey)] {
			key := fmt.Sprintf("%s:%d", info.MegapoolAddress.Hex(), info.ValidatorId)
			if _, ok := seen[key]; ok {
				continue
			}
			credentials := services.CalculateMegapoolWithdrawalCredentials(info.MegapoolAddress)
			if bytes.Equal(deposit.WithdrawalCredentials, credentials.Bytes()) {
				continue
			}
			seen[key] = struct{}{}
			matches = append(matches, pendingDepositDissolve{
				info:         info,
				depositIndex: uint64(index),
			})
		}
	}
	return matches
}

// pendingDepositSignatureValid reports whether the deposit signature verifies
// against that deposit's own withdrawal credentials and amount.
func pendingDepositSignatureValid(deposit *generic.PendingDeposit, depositDomain []byte) bool {
	if deposit == nil || len(deposit.Pubkey) == 0 || len(deposit.Signature) == 0 || len(deposit.WithdrawalCredentials) == 0 {
		return false
	}
	depositData := &ethpb.Deposit_Data{
		Amount:                deposit.Amount,
		PublicKey:             deposit.Pubkey,
		WithdrawalCredentials: deposit.WithdrawalCredentials,
		Signature:             deposit.Signature,
	}
	return prdeposit.VerifyDepositSignature(depositData, depositDomain) == nil
}

// classifyPrestake reports whether a prestake can be challenged.
// An empty signature is a legacy validator and cannot be challenged.
func classifyPrestake(info megapool.ValidatorPrestakeInfo, signatureValid bool) prestakeValidation {
	if !info.ValidatorInfo.InPrestake || len(info.Signature) == 0 {
		return prestakeSkipped
	}
	if signatureValid {
		return prestakeValid
	}
	return prestakeInvalidSignature
}

// prestakeSignatureValid reports whether the retained signature is a valid 1 ETH deposit into the megapool.
func prestakeSignatureValid(pubkey, signature []byte, withdrawalCredentials common.Hash, depositDomain []byte) bool {
	if len(pubkey) == 0 || len(signature) == 0 {
		return false
	}
	depositData := &ethpb.Deposit_Data{
		Amount:                prestakeAmountGwei,
		PublicKey:             pubkey,
		WithdrawalCredentials: withdrawalCredentials.Bytes(),
		Signature:             signature,
	}
	return prdeposit.VerifyDepositSignature(depositData, depositDomain) == nil
}

func prestakePageAdvances(offset, next *big.Int) bool {
	return next != nil && next.Cmp(offset) > 0
}

func uint256ToUint64(value *big.Int) (uint64, error) {
	if value == nil || value.Sign() < 0 || !value.IsUint64() {
		return 0, fmt.Errorf("value %v does not fit in uint64", value)
	}
	return value.Uint64(), nil
}
