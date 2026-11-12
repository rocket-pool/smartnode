package node

import (
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
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

const (
	// prestakeAmountGwei is the 1 ETH deposit checked by challengeInvalidSignature.
	prestakeAmountGwei uint64 = 1_000_000_000
	// prestakeIndexPageSize is how many global indices one getPrestakedValidatorIndices call inspects.
	prestakeIndexPageSize int64 = 200
	// maxPrestakeChallengesPerCycle keeps one pass from holding the node loop on a large backlog.
	maxPrestakeChallengesPerCycle = 8
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

	gas := loadAutoTxGas(cfg, &logger)
	return &validatePrestakedValidators{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		gasThreshold:   gas.thresholdGwei,
		maxFee:         gas.maxFee,
		maxPriorityFee: gas.maxPriorityFee,
		gasLimit:       0,
	}, nil
}

// Check prestaked validators and challenge those with an invalid deposit signature
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
	// challengeInvalidSignature requires the caller to have a deployed megapool.
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

	var invalid []megapool.ValidatorPrestakeInfo
	for _, index := range indices {
		info, err := megapool.GetValidatorPrestakeInfo(t.rp, new(big.Int).SetUint64(index), opts)
		if err != nil {
			t.log.Printlnf("Error reading prestake info for global index %d: %v", index, err)
			continue
		}
		credentials := services.CalculateMegapoolWithdrawalCredentials(info.MegapoolAddress)
		valid := prestakeSignatureValid(info.Pubkey, info.Signature, credentials, depositDomain)
		switch classifyPrestake(info, valid) {
		case prestakeInvalidSignature:
			t.log.Printlnf("Invalid prestake signature for megapool %s validator %d (global index %d)", info.MegapoolAddress.Hex(), info.ValidatorId, index)
			invalid = append(invalid, info)
		case prestakeSkipped, prestakeValid:
		}
	}
	if len(invalid) == 0 {
		return nil
	}
	if !canChallenge {
		t.log.Printlnf("%d prestaked validators have invalid signatures. This node cannot challenge until its megapool is deployed.", len(invalid))
		return nil
	}
	return t.challengeInvalid(invalid)
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
