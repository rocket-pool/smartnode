package services

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v3"

	megapool140 "github.com/rocket-pool/smartnode/bindings/legacy/v1.4.0/megapool"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

func encodeValidatorBundleIfNeeded(rp *rocketpool.RocketPool, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof) (bool, []byte, error) {
	useBundles, err := megapool.UsesProofBundles(rp, nil)
	if err != nil {
		return false, nil, fmt.Errorf("error checking beacon state verifier version: %w", err)
	}
	if !useBundles {
		return false, nil, nil
	}
	proofData, err := megapool.EncodeValidatorProofBundleV1(validatorProof, slotProof)
	if err != nil {
		return false, nil, err
	}
	return true, proofData, nil
}

func EstimateMegapoolStakeGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	if useBundles {
		return megapool.EstimateStakeGas(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.EstimateStakeGas(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func StakeMegapool(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return nil, err
	}
	if useBundles {
		return megapool.Stake(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.Stake(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func EstimateMegapoolNotifyExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	if useBundles {
		return megapool.EstimateNotifyExitGas(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.EstimateNotifyExitGas(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func NotifyMegapoolExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return nil, err
	}
	if useBundles {
		return megapool.NotifyExit(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.NotifyExit(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func EstimateMegapoolNotifyNotExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	if useBundles {
		return megapool.EstimateNotifyNotExitGas(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.EstimateNotifyNotExitGas(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func NotifyMegapoolNotExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return nil, err
	}
	if useBundles {
		return megapool.NotifyNotExit(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.NotifyNotExit(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func EstimateMegapoolDissolveWithProofGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	if useBundles {
		return megapool.EstimateDissolveWithProof(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.EstimateDissolveWithProof(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

func DissolveMegapoolWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
	useBundles, proofData, err := encodeValidatorBundleIfNeeded(rp, validatorProof, slotProof)
	if err != nil {
		return nil, err
	}
	if useBundles {
		return megapool.DissolveWithProof(rp, megapoolAddress, validatorId, slotTimestamp, megapool.ValidatorProofVersion1, proofData, opts)
	}
	return megapool140.DissolveWithProof(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
}

// MegapoolFinalBalanceProof is a version-selected payload for notifyFinalBalance.
// Build it once with BuildMegapoolFinalBalanceProof and reuse for gas estimate and submit.
type MegapoolFinalBalanceProof struct {
	useBundles      bool
	slotTimestamp   uint64
	proofVersion    *big.Int
	proofData       []byte
	withdrawalProof megapool.WithdrawalProof
	validatorProof  megapool.ValidatorProof
	slotProof       megapool.SlotProof
}

func BuildMegapoolFinalBalanceProof(c *cli.Command, rp *rocketpool.RocketPool, megapoolAddress common.Address, slotHint uint64, validatorIndex uint64, validatorPubkey rptypes.ValidatorPubkey, w wallet.Wallet) (*MegapoolFinalBalanceProof, error) {
	useBundles, err := megapool.UsesProofBundles(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking beacon state verifier version: %w", err)
	}
	if useBundles {
		proofVersion, proofData, slotTimestamp, err := GetFinalBalanceProofBundle(c, slotHint, validatorIndex, validatorPubkey, megapoolAddress, w)
		if err != nil {
			return nil, err
		}
		return &MegapoolFinalBalanceProof{
			useBundles:    true,
			slotTimestamp: slotTimestamp,
			proofVersion:  proofVersion,
			proofData:     proofData,
		}, nil
	}
	withdrawalProof, validatorProof, slotProof, slotTimestamp, err := GetFinalBalanceProofs(c, slotHint, validatorIndex, validatorPubkey, megapoolAddress, w)
	if err != nil {
		return nil, err
	}
	return &MegapoolFinalBalanceProof{
		slotTimestamp:   slotTimestamp,
		withdrawalProof: withdrawalProof,
		validatorProof:  validatorProof,
		slotProof:       slotProof,
	}, nil
}

func EstimateMegapoolNotifyFinalBalanceGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, proof *MegapoolFinalBalanceProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	if proof.useBundles {
		return megapool.EstimateNotifyFinalBalance(rp, megapoolAddress, validatorId, proof.slotTimestamp, proof.proofVersion, proof.proofData, opts)
	}
	return megapool140.EstimateNotifyFinalBalance(rp, megapoolAddress, validatorId, proof.slotTimestamp, proof.withdrawalProof, proof.validatorProof, proof.slotProof, opts)
}

func NotifyMegapoolFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, proof *MegapoolFinalBalanceProof, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
	if proof.useBundles {
		return megapool.NotifyFinalBalance(rp, megapoolAddress, validatorId, proof.slotTimestamp, proof.proofVersion, proof.proofData, opts)
	}
	return megapool140.NotifyFinalBalance(rp, megapoolAddress, validatorId, proof.slotTimestamp, proof.withdrawalProof, proof.validatorProof, proof.slotProof, opts)
}
