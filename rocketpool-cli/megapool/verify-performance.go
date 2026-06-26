package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	verifyperf "github.com/rocket-pool/smartnode/shared/utils/cli/verify-performance"
)

func verifyMegapoolPerformance(megapoolAddress common.Address, validatorId uint32, startEpoch uint64, epochs uint64, yes bool) error {
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

	resp, err := rp.VerifyMegapoolValidatorPerformance(megapoolAddress, validatorId, startEpoch, endEpoch)
	if err != nil {
		return err
	}

	label := fmt.Sprintf("megapool validator %d", validatorId)
	if (megapoolAddress != common.Address{}) {
		label = fmt.Sprintf("megapool %s validator %d", megapoolAddress.Hex(), validatorId)
	}
	verifyperf.PrintResult(resp, label)
	return nil
}
