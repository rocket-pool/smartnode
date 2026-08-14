package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	daoprotocol "github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeSettingMulti(c *cli.Command, settings []api.PDAOBatchSetting, customMessage string) (*api.CanProposePDAOSettingMultiResponse, error) {
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	response := api.CanProposePDAOSettingMultiResponse{}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	var stakedRpl *big.Int
	var lockedRpl *big.Int
	var proposalBond *big.Int
	var isRplLockingAllowed bool
	var wg errgroup.Group

	wg.Go(func() error {
		var err error
		stakedRpl, err = node.GetNodeStakedRPL(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		lockedRpl, err = node.GetNodeLockedRPL(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		proposalBond, err = protocol.GetProposalBond(rp, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		isRplLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
		return err
	})
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	response.StakedRpl = stakedRpl
	response.LockedRpl = lockedRpl
	response.ProposalBond = proposalBond
	response.IsRplLockingDisallowed = !isRplLockingAllowed

	freeRpl := big.NewInt(0).Sub(stakedRpl, lockedRpl)
	response.InsufficientRpl = freeRpl.Cmp(proposalBond) < 0
	response.CanPropose = !response.InsufficientRpl && !response.IsRplLockingDisallowed
	if !response.CanPropose {
		return &response, nil
	}

	decoded, err := decodeBatchSettings(settings, customMessage)
	if err != nil {
		return nil, err
	}

	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, fmt.Errorf("error creating pollard: %w", err)
	}
	response.BlockNumber = blockNumber

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	response.GasLimits, err = daoprotocol.EstimateProposeSetMultiGas(rp, decoded.message, decoded.contractNames, decoded.settingPaths, decoded.settingTypes, decoded.values, blockNumber, pollard, opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating gas for multi-setting proposal: %w", err)
	}

	return &response, nil
}

func proposeSettingMulti(c *cli.Command, settings []api.PDAOBatchSetting, customMessage string, blockNumber uint32, opts *bind.TransactOpts) (*api.ProposePDAOSettingMultiResponse, error) {
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
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

	decoded, err := decodeBatchSettings(settings, customMessage)
	if err != nil {
		return nil, err
	}

	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error regenerating pollard: %w", err)
	}

	proposalID, hash, err := daoprotocol.ProposeSetMulti(rp, decoded.message, decoded.contractNames, decoded.settingPaths, decoded.settingTypes, decoded.values, blockNumber, pollard, opts)
	if err != nil {
		return nil, fmt.Errorf("error proposing multi-setting update: %w", err)
	}

	return &api.ProposePDAOSettingMultiResponse{
		ProposalId: proposalID,
		TxHash:     hash,
	}, nil
}
