package tx

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func signTx(c *cli.Context, txInfo *core.TransactionInfo) (*api.SignedTxData, error) {
	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Rocket Pool binding: %w", err)
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, fmt.Errorf("error getting wallet: %w", err)
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, fmt.Errorf("error getting node account transactor: %w", err)
	}

	// Response
	response := api.SignedTxData{}

	// Sign it
	tx, err := rp.SignTransaction(txInfo, opts)
	if err != nil {
		return nil, fmt.Errorf("error signing transaction: %w", err)
	}

	// Marshal it
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("error marshalling signed transaction: %w", err)
	}

	// Return
	response.TxBytes = txBytes
	return &response, nil
}
