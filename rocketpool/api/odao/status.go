package odao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoStatusContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoStatusContextFactory) Create(vars map[string]string) (*oracleDaoStatusContext, error) {
	c := &oracleDaoStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoStatusContext, api.OracleDaoStatusData](
		router, "status", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoStatusContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	odaoMember *trustednode.OracleDaoMember
	oSettings  *settings.OracleDaoSettings
	dnt        *trustednode.DaoNodeTrusted
	dp         *dao.DaoProposal
}

func (c *oracleDaoStatusContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.odaoMember, err = trustednode.NewOracleDaoMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating oracle DAO member binding: %w", err)
	}
	c.oSettings, err = settings.NewOracleDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating oracle DAO settings binding: %w", err)
	}
	c.dnt, err = trustednode.NewDaoNodeTrusted(c.rp)
	if err != nil {
		return fmt.Errorf("error creating DNT binding: %w", err)
	}
	c.dp, err = dao.NewDaoProposal(c.rp)
	if err != nil {
		return fmt.Errorf("error creating DP binding: %w", err)
	}
	return nil
}

func (c *oracleDaoStatusContext) GetState(mc *batch.MultiCaller) {
	c.odaoMember.GetExists(mc)
	c.odaoMember.GetInvitedTime(mc)
	c.odaoMember.GetReplacedTime(mc)
	c.odaoMember.GetLeftTime(mc)
	c.dnt.GetMemberCount(mc)
	c.dp.GetProposalCount(mc)
}

func (c *oracleDaoStatusContext) PrepareData(data *api.AuctionClaimFromLotData, opts *bind.TransactOpts) error {
	// Check for validity
	data.DoesNotExist = !c.lot.Details.Exists
	data.NoBidFromAddress = (c.addressBidAmount.Cmp(big.NewInt(0)) == 0)
	data.NotCleared = !c.lot.Details.IsCleared
	data.CanClaim = !(data.DoesNotExist || data.NoBidFromAddress || data.NotCleared)

	// Get tx info
	if data.CanClaim && opts != nil {
		txInfo, err := c.lot.ClaimBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}

func getStatus(c *cli.Context) (*api.OracleDaoStatusData, error) {

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
	response := api.OracleDaoStatusData{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get membership status
	isMember, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.IsMember = isMember

	// Sync
	var wg errgroup.Group

	// Get pending executed proposal statuses
	if isMember {

		// Check if node can leave
		wg.Go(func() error {
			leaveActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "leave")
			if err == nil {
				response.CanLeave = leaveActionable
			}
			return err
		})

		// Check if node can replace position
		wg.Go(func() error {
			replaceActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "replace")
			if err == nil {
				response.CanReplace = replaceActionable
			}
			return err
		})

	} else {

		// Check if node can join
		wg.Go(func() error {
			joinActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "invited")
			if err == nil {
				response.CanJoin = joinActionable
			}
			return err
		})

	}

	// Get total DAO members
	wg.Go(func() error {
		memberCount, err := trustednode.GetMemberCount(rp, nil)
		if err == nil {
			response.TotalMembers = memberCount
		}
		return err
	})

	// Get proposal counts
	wg.Go(func() error {
		proposalStates, err := getProposalStates(rp)
		if err == nil {
			response.ProposalCounts.Total = len(proposalStates)
			for _, state := range proposalStates {
				switch state {
				case rptypes.Pending:
					response.ProposalCounts.Pending++
				case rptypes.Active:
					response.ProposalCounts.Active++
				case rptypes.Cancelled:
					response.ProposalCounts.Cancelled++
				case rptypes.Defeated:
					response.ProposalCounts.Defeated++
				case rptypes.Succeeded:
					response.ProposalCounts.Succeeded++
				case rptypes.Expired:
					response.ProposalCounts.Expired++
				case rptypes.Executed:
					response.ProposalCounts.Executed++
				}
			}
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
