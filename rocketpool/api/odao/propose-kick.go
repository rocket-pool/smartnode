package odao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func canProposeKick(c *cli.Context, memberAddress common.Address, fineAmountWei *big.Int) (*api.CanProposeTNDAOKickResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
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
	response := api.CanProposeTNDAOKickResponse{}

	// Sync
	var wg errgroup.Group

	// Check if proposal cooldown is active
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		proposalCooldownActive, err := getProposalCooldownActive(rp, nodeAccount.Address)
		if err == nil {
			response.ProposalCooldownActive = proposalCooldownActive
		}
		return err
	})

	// Check member's RPL bond amount
	wg.Go(func() error {
		rplBondAmount, err := trustednode.GetMemberRPLBondAmount(rp, memberAddress, nil)
		if err == nil {
			response.InsufficientRplBond = (fineAmountWei.Cmp(rplBondAmount) > 0)
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		memberId, err := trustednode.GetMemberID(rp, memberAddress, nil)
		if err != nil {
			return err
		}
		memberUrl, err := trustednode.GetMemberUrl(rp, memberAddress, nil)
		if err != nil {
			return err
		}
		message := fmt.Sprintf("kick %s (%s) with %.6f RPL fine", memberId, memberUrl, math.RoundDown(eth.WeiToEth(fineAmountWei), 6))
		gasInfo, err := trustednode.EstimateProposeKickMemberGas(rp, message, memberAddress, fineAmountWei, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanPropose = !(response.ProposalCooldownActive || response.InsufficientRplBond)
	return &response, nil

}

func proposeKick(c *cli.Context, memberAddress common.Address, fineAmountWei *big.Int) (*api.ProposeTNDAOKickResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
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
	response := api.ProposeTNDAOKickResponse{}

	// Data
	var wg errgroup.Group
	var memberId string
	var memberUrl string

	// Get member details
	wg.Go(func() error {
		var err error
		memberId, err = trustednode.GetMemberID(rp, memberAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		memberUrl, err = trustednode.GetMemberUrl(rp, memberAddress, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	message := fmt.Sprintf("kick %s (%s) with %.6f RPL fine", memberId, memberUrl, math.RoundDown(eth.WeiToEth(fineAmountWei), 6))
	proposalId, hash, err := trustednode.ProposeKickMember(rp, message, memberAddress, fineAmountWei, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}
