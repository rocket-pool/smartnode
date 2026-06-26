package megapool

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// verifyPerformance computes the RPIP-73 target-vote performance over the
// inclusive epoch range [startEpoch, endEpoch] for one or more validators of a
// megapool. If megapoolAddress is the zero address, the node's own megapool
// address is used
func verifyPerformance(
	c *cli.Command,
	megapoolAddress common.Address,
	targets string,
	startEpoch uint64,
	endEpoch uint64,
) (*api.VerifyPerformanceBatchResponse, error) {
	if err := services.RequireBeaconClientSynced(c); err != nil {
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

	if (megapoolAddress == common.Address{}) {
		if err := services.RequireNodeRegistered(c); err != nil {
			return nil, fmt.Errorf("no megapool address supplied and node is not registered: %w", err)
		}
		w, err := services.GetWallet(c)
		if err != nil {
			return nil, err
		}
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		megapoolAddress, err = node.GetMegapoolAddress(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error looking up node's megapool address: %w", err)
		}
		if (megapoolAddress == common.Address{}) {
			return nil, fmt.Errorf("node has no megapool deployed; pass --megapool to specify one")
		}
	}

	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating megapool binding for %s: %w", megapoolAddress.Hex(), err)
	}

	validatorIds, err := resolveMegapoolTargets(mp, targets)
	if err != nil {
		return nil, err
	}
	if len(validatorIds) == 0 {
		return nil, fmt.Errorf("no megapool validators to verify")
	}

	// Resolve each validator's pubkey up front. Per-validator pubkey failures
	// are recorded and excluded from the beacon batch.
	pubkeys := make([]rptypes.ValidatorPubkey, len(validatorIds))
	pubkeyErrs := make([]string, len(validatorIds))
	for i, validatorId := range validatorIds {
		pubkey, err := mp.GetValidatorPubkey(validatorId, nil)
		if err != nil {
			pubkeyErrs[i] = fmt.Sprintf("error getting megapool %s validator %d pubkey: %s", megapoolAddress.Hex(), validatorId, err.Error())
			continue
		}
		pubkeys[i] = pubkey
	}

	batch, err := performance.VerifyPerformanceBatch(rp, bc, pubkeys, startEpoch, endEpoch)
	if err != nil {
		return nil, err
	}

	response := &api.VerifyPerformanceBatchResponse{
		Results: make([]api.VerifyPerformanceResult, 0, len(validatorIds)),
	}
	for i, validatorId := range validatorIds {
		if !batch[i].Active {
			continue
		}
		result := api.VerifyPerformanceResult{ValidatorId: validatorId}
		switch {
		case pubkeyErrs[i] != "":
			result.Error = pubkeyErrs[i]
		case batch[i].Err != nil:
			result.Error = batch[i].Err.Error()
		default:
			result.Performance = batch[i].Response
		}
		response.Results = append(response.Results, result)
	}

	return response, nil
}

// resolveMegapoolTargets turns the targets string into a concrete list of validator IDs
func resolveMegapoolTargets(mp megapool.Megapool, targets string) ([]uint32, error) {
	if strings.EqualFold(strings.TrimSpace(targets), "all") {
		count, err := mp.GetValidatorCount(nil)
		if err != nil {
			return nil, fmt.Errorf("error getting megapool validator count: %w", err)
		}
		ids := make([]uint32, 0, count)
		for i := uint32(0); i < count; i++ {
			ids = append(ids, i)
		}
		return ids, nil
	}

	var ids []uint32
	for _, raw := range strings.Split(targets, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := cliutils.ValidateUint32("validator-id", raw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
