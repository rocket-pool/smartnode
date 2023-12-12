package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoRewardsPercentagesContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoRewardsPercentagesContextFactory) Create(vars map[string]string) (*protocolDaoRewardsPercentagesContext, error) {
	c := &protocolDaoRewardsPercentagesContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoRewardsPercentagesContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoRewardsPercentagesContext, api.ProtocolDaoRewardsPercentagesData](
		router, "rewards-percentages", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoRewardsPercentagesContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	percentages protocol.RplRewardsPercentages
	pdaoMgr     *protocol.ProtocolDaoManager
}

func (c *protocolDaoRewardsPercentagesContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Bindings
	var err error
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	return nil
}

func (c *protocolDaoRewardsPercentagesContext) GetState(mc *batch.MultiCaller) {
	c.pdaoMgr.GetRewardsPercentages(mc, &c.percentages)
}

func (c *protocolDaoRewardsPercentagesContext) PrepareData(data *api.ProtocolDaoRewardsPercentagesData, opts *bind.TransactOpts) error {
	data.OracleDao = c.percentages.OdaoPercentage
	data.Node = c.percentages.NodePercentage
	data.ProtocolDao = c.percentages.PdaoPercentage
	return nil
}
