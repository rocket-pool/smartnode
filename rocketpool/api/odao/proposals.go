package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoProposalsContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoProposalsContextFactory) Create(vars map[string]string) (*oracleDaoProposalsContext, error) {
	c := &oracleDaoProposalsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposalsContext, api.OracleDaoMembersData](
		router, "proposals", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoProposalsContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address
	hasAddress  bool

	odaoMgr *oracle.OracleDaoManager
	dpm     *proposals.DaoProposalManager
}

func (c *oracleDaoProposalsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, c.hasAddress = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	return nil
}

func (c *oracleDaoProposalsContext) GetState(mc *batch.MultiCaller) {
	c.dpm.GetProposalCount(mc)
}

func (c *oracleDaoProposalsContext) PrepareData(data *api.OracleDaoProposalsData, opts *bind.TransactOpts) error {
	_, odaoProps, err := c.dpm.GetProposals(c.dpm.ProposalCount.Formatted(), true, nil)
	if err != nil {
		return fmt.Errorf("error getting proposals: %w", err)
	}

	for _, odaoProp := range odaoProps {
		prop := api.OracleDaoProposalDetails{
			ID:              odaoProp.ID.Formatted(),
			ProposerAddress: odaoProp.ProposerAddress,
			Message:         odaoProp.Message,
			CreatedTime:     odaoProp.CreatedTime.Formatted(),
			StartTime:       odaoProp.StartTime.Formatted(),
			EndTime:         odaoProp.EndTime.Formatted(),
			ExpiryTime:      odaoProp.ExpiryTime.Formatted(),
			VotesRequired:   odaoProp.VotesRequired.Formatted(),
			VotesFor:        odaoProp.VotesFor.Formatted(),
			VotesAgainst:    odaoProp.VotesAgainst.Formatted(),
			IsCancelled:     odaoProp.IsCancelled,
			IsExecuted:      odaoProp.IsExecuted,
			Payload:         odaoProp.Payload,
			PayloadStr:      c.dpm.GetPayloadAsString(),
		}

		if c.hasAddress {
			prop.MemberVoted
		}
	}

	return nil
}

func getProposals(c *cli.Context) (*api.OracleDaoProposalsData, error) {

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
	response := api.OracleDaoProposalsData{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposals
	proposals, err := dao.GetDAOProposalsWithMember(rp, "rocketDAONodeTrustedProposals", nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	response.Proposals = proposals

	// Return response
	return &response, nil

}

func getProposal(c *cli.Context, id uint64) (*api.OracleDaoProposalData, error) {

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
	response := api.OracleDaoProposalData{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposals
	proposal, err := dao.GetProposalDetailsWithMember(rp, id, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	response.Proposal = proposal

	// Return response
	return &response, nil

}
