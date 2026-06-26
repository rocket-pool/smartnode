package minipool

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// verifyPerformance computes the RPIP-73 target-vote performance over the
// inclusive epoch range [startEpoch, endEpoch] for one or more of the node's
// minipools
func verifyPerformance(
	c *cli.Command,
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

	addresses, err := resolveMinipoolTargets(c, rp, targets)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no minipools to verify")
	}

	// Resolve each minipool's validator pubkey up front. Per-minipool pubkey
	// failures are recorded and excluded from the beacon batch.
	pubkeys := make([]rptypes.ValidatorPubkey, len(addresses))
	pubkeyErrs := make([]string, len(addresses))
	for i, address := range addresses {
		pubkey, err := minipool.GetMinipoolPubkey(rp, address, nil)
		if err != nil {
			pubkeyErrs[i] = fmt.Sprintf("error getting minipool %s pubkey: %s", address.Hex(), err.Error())
			continue
		}
		pubkeys[i] = pubkey
	}

	batch, err := performance.VerifyPerformanceBatch(rp, bc, pubkeys, startEpoch, endEpoch)
	if err != nil {
		return nil, err
	}

	response := &api.VerifyPerformanceBatchResponse{
		Results: make([]api.VerifyPerformanceResult, 0, len(addresses)),
	}
	for i, address := range addresses {
		if !batch[i].Active {
			continue
		}
		result := api.VerifyPerformanceResult{MinipoolAddress: address}
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

// resolveMinipoolTargets turns the targets string ("all" or a comma-separated
// list of addresses) into a concrete list of minipool addresses
func resolveMinipoolTargets(c *cli.Command, rp *rocketpool.RocketPool, targets string) ([]common.Address, error) {
	if strings.EqualFold(strings.TrimSpace(targets), "all") {
		if err := services.RequireNodeRegistered(c); err != nil {
			return nil, err
		}
		w, err := services.GetWallet(c)
		if err != nil {
			return nil, err
		}
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting node minipool addresses: %w", err)
		}
		return addresses, nil
	}

	var addresses []common.Address
	for _, raw := range strings.Split(targets, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		address, err := cliutils.ValidateAddress("minipool address", raw)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}
	return addresses, nil
}
