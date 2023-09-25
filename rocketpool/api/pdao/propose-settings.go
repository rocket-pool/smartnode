package pdao

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func canProposeSetting(c *cli.Context, settingName string, value string) (*api.CanProposePDAOSettingResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.CanProposePDAOSettingResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Sync
	var stakedRpl *big.Int
	var lockedRpl *big.Int
	var proposalBond *big.Int
	var wg errgroup.Group

	// Get the node's RPL stake
	wg.Go(func() error {
		var err error
		stakedRpl, err = node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the node's locked RPL
	wg.Go(func() error {
		var err error
		lockedRpl, err = node.GetNodeRplLocked(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the node's RPL stake
	wg.Go(func() error {
		var err error
		proposalBond, err = protocol.GetProposalBond(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	response.StakedRpl = stakedRpl
	response.LockedRpl = lockedRpl
	response.ProposalBond = proposalBond

	freeRpl := big.NewInt(0).Sub(stakedRpl, lockedRpl)
	response.InsufficientRpl = (freeRpl.Cmp(proposalBond) < 0)

	// Update & return response
	response.CanPropose = !(response.InsufficientRpl)
	return &response, nil

}
