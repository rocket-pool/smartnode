package megapool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func verifyValidator(rp *rocketpool.RocketPool, proof ValidatorProof, opts *bind.CallOpts) (bool, error) {
	beaconStateVerifier, err := getBeaconStateVerifier(rp, opts)
	if err != nil {
		return false, err
	}
	verifiedValidator := new(bool)
	if err := beaconStateVerifier.Call(opts, verifiedValidator, "verifyValidator"); err != nil {
		return false, fmt.Errorf("error verifying validatorindex %d at slot %d: %w", proof.ValidatorIndex, proof.Slot, err)
	}
	return *verifiedValidator, nil
}

func verifyExit(rp *rocketpool.RocketPool, validatorIndex *big.Int, withdrawableEpoch *big.Int, slot uint64, proof [][32]byte, opts *bind.CallOpts) (bool, error) {
	beaconStateVerifier, err := getBeaconStateVerifier(rp, opts)
	if err != nil {
		return false, err
	}
	verifiedExit := new(bool)
	if err := beaconStateVerifier.Call(opts, verifiedExit, "verifyExit", validatorIndex, withdrawableEpoch, slot, proof); err != nil {
		return false, fmt.Errorf("error verifying exit of validator index %d at slot %d: %w", validatorIndex.Int64(), slot, err)
	}
	return *verifiedExit, nil
}

func verifyWithdrawal(rp *rocketpool.RocketPool, validatorIndex *big.Int, withdrawalSlot uint64, withdrawalNum *big.Int, withdrawal withdrawal, slot uint64, proof [][32]byte, opts *bind.CallOpts) (bool, error) {
	beaconStateVerifier, err := getBeaconStateVerifier(rp, opts)
	if err != nil {
		return false, err
	}
	verifiedWithdrawal := new(bool)
	if err := beaconStateVerifier.Call(opts, verifiedWithdrawal, "verifyWithdrawal", validatorIndex, withdrawalSlot, withdrawalNum, withdrawal, slot, proof, opts); err != nil {
		return false, fmt.Errorf("error verifying withdrawal of validator index %d at withdrawalSlot %d: %w", validatorIndex.Int64(), withdrawalSlot, err)
	}
	return *verifiedWithdrawal, nil
}

var BeaconStateVerifierLock sync.Mutex

func getBeaconStateVerifier(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	BeaconStateVerifierLock.Lock()
	defer BeaconStateVerifierLock.Unlock()
	return rp.GetContract("beaconStateVerifierLock", opts)
}
