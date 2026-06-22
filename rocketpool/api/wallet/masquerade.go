package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func masquerade(c *cli.Command, address common.Address, observe bool) (*api.MasqueradeResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	if observe {
		hdw, err := services.GetHdWallet(c)
		if err != nil {
			return nil, err
		}
		rp, err := services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
		nodeAccount, err := hdw.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		isMember, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error checking Oracle DAO membership: %w", err)
		}
		if isMember {
			return nil, fmt.Errorf("Observe mode is not available for Oracle DAO nodes: oDAO duties would stop running while observing")
		}
	}

	if err := w.MasqueradeAsAddress(address, observe); err != nil {
		return nil, fmt.Errorf("error masquerading as address %s: %w", address.Hex(), err)
	}

	response := api.MasqueradeResponse{}

	return &response, nil

}
