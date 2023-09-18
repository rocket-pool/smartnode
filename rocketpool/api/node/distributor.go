package node

import (
	"context"
	"fmt"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/node"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ==================================
// === Initialize Fee Distributor ===
// ==================================
type nodeFeeDistributorInitHandler struct {
}

func (h *nodeFeeDistributorInitHandler) CreateBindings(ctx *callContext) error {
	return nil
}

func (h *nodeFeeDistributorInitHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	ctx.node.GetFeeDistributorInitialized(mc)
	ctx.node.GetDistributorAddress(mc)
}

func (h *nodeFeeDistributorInitHandler) PrepareResponse(ctx *callContext, response *api.NodeInitializeFeeDistributorResponse) error {
	node := ctx.node
	opts := ctx.opts

	response.Distributor = node.DistributorAddress
	response.IsInitialized = node.IsFeeDistributorInitialized
	if response.IsInitialized {
		return nil
	}

	// Get tx info
	txInfo, err := node.InitializeFeeDistributor(opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for InitializeFeeDistributor: %w", err)
	}
	response.TxInfo = txInfo
	return nil
}

// ==================
// === Distribute ===
// ==================
type nodeDistributeHandler struct {
}

func (h *nodeDistributeHandler) CreateBindings(ctx *callContext) error {
	return nil
}

func (h *nodeDistributeHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	ctx.node.GetFeeDistributorInitialized(mc)
	ctx.node.GetDistributorAddress(mc)
}

func (h *nodeDistributeHandler) PrepareResponse(ctx *callContext, response *api.NodeDistributeResponse) error {
	rp := ctx.rp
	nodeBinding := ctx.node
	opts := ctx.opts

	// Make sure it's initialized before proceeding
	response.IsInitialized = nodeBinding.IsFeeDistributorInitialized
	if !response.IsInitialized {
		return nil
	}

	// Create the distributor
	distributorAddress := nodeBinding.DistributorAddress
	distributor, err := node.NewNodeDistributor(rp, nodeBinding.Address, distributorAddress)
	if err != nil {
		return fmt.Errorf("error creating node distributor binding: %w", err)
	}

	// Get its balance
	response.Balance, err = rp.Client.BalanceAt(context.Background(), distributorAddress, nil)
	if err != nil {
		return fmt.Errorf("error getting fee distributor balance: %w", err)
	}

	// Get the node share of the balance
	err = rp.Query(func(mc *batch.MultiCaller) error {
		distributor.GetNodeShare(mc)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting node share for distributor %s: %w", distributorAddress.Hex(), err)
	}
	response.NodeShare = distributor.NodeShare

	// Get tx info
	txInfo, err := distributor.Distribute(opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for Distribute: %w", err)
	}
	response.TxInfo = txInfo
	return nil
}
