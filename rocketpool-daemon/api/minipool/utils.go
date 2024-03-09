package minipool

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/minipool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	minipoolAddressBatchSize int = 100
)

// Get transaction info for an operation on all of the provided minipools, using the common minipool API (for version-agnostic functions)
func prepareMinipoolBatchTxData(sp *services.ServiceProvider, minipoolAddresses []common.Address, data *api.BatchTxInfoData, txCreator func(mp minipool.IMinipool, opts *bind.TransactOpts) (*eth.TransactionInfo, error), txName string) error {
	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}

	// TX opts
	rp := sp.GetRocketPool()
	opts, err := sp.GetWallet().GetTransactor()
	if err != nil {
		return fmt.Errorf("error creating node transactor: %w", err)
	}

	// Response
	response := api.BatchTxInfoData{}

	// Create minipools
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	mps, err := mpMgr.CreateMinipoolsFromAddresses(minipoolAddresses, false, nil)
	if err != nil {
		return fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the TXs
	txInfos := make([]*eth.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		txInfo, err := txCreator(mp, opts)
		if err != nil {
			return fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, mp.Common().Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return nil
}
