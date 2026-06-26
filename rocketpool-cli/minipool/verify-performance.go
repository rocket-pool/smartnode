package minipool

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	verifyperf "github.com/rocket-pool/smartnode/shared/utils/cli/verify-performance"
)

func verifyMinipoolPerformance(address common.Address, startEpoch uint64, epochs uint64, yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	startEpoch, endEpoch, err := verifyperf.ResolveEpochRange(rp, startEpoch, epochs)
	if err != nil {
		return err
	}
	if !yes && !verifyperf.ConfirmLargeRange(endEpoch-startEpoch+1) {
		return verifyperf.PrintCancelled()
	}

	resp, err := rp.VerifyMinipoolPerformance(address, startEpoch, endEpoch)
	if err != nil {
		return err
	}
	verifyperf.PrintResult(resp, "minipool "+address.Hex())
	return nil
}
