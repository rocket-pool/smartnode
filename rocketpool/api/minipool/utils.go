package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get transaction info for an operation on all of the provided minipools, using the common minipool API (for version-agnostic functions)
func createBatchTxResponseForCommon(c *cli.Context, minipoolAddresses []common.Address, txCreator func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error), txName string) (*api.BatchTxInfoData, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxInfoData{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		txInfo, err := txCreator(mpCommon, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, mpCommon.Details.Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}

// Get transaction info for an operation on all of the provided minipools, using the v3 minipool API (for Atlas-specific functions)
func createBatchTxResponseForV3(c *cli.Context, minipoolAddresses []common.Address, txCreator func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error), txName string) (*api.BatchTxInfoData, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxInfoData{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			return nil, fmt.Errorf("minipool %s is too old (current version: %d); please upgrade the delegate for it first", minipoolAddress.Hex(), mpCommon.Details.Version)
		}
		txInfo, err := txCreator(mpv3, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, minipoolAddress.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}
