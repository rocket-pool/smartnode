package pdao

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/shared/eip712"
)

// ===============
// === Factory ===
// ===============

type protocolDaoSetSignallingAddressFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoSetSignallingAddressFactory) Create(args url.Values) (*protocolDaoSetSignallingAddressContext, error) {
	c := &protocolDaoSetSignallingAddressContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("signallingAddress", args, input.ValidateAddress, &c.signallingAddress),
		server.GetStringFromVars("signature", args, &c.signature),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoSetSignallingAddressFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoSetSignallingAddressContext, types.TxInfoData](
		router, "set-signalling-address", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoSetSignallingAddressContext struct {
	handler  *ProtocolDaoHandler
	rp       *rocketpool.RocketPool
	registry *contracts.RocketSignerRegistry

	node              *node.Node
	signallingAddress common.Address
	nodeToSigner      common.Address
	signature         string
}

func (c *protocolDaoSetSignallingAddressContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()
	cfg := sp.GetConfig()
	network := cfg.GetNetworkResources().Network

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	c.registry = sp.GetRocketSignerRegistry()
	if c.registry == nil {
		return types.ResponseStatus_Error, fmt.Errorf("Network [%v] does not have a signer registry contract.", network)
	}

	// Binding
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error creating node %s binding: %w", nodeAddress.Hex(), err)
	}

	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoSetSignallingAddressContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.Exists,
		c.node.IsVotingInitialized,
	)
	// Check if the node already has a signer
	if c.registry != nil {
		c.registry.NodeToSigner(mc, &c.nodeToSigner, c.node.Address)
	}
}

func (c *protocolDaoSetSignallingAddressContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {

	if c.signallingAddress == c.nodeToSigner {
		return types.ResponseStatus_Error, fmt.Errorf("Signer address already in use")
	}

	if !c.node.IsVotingInitialized.Get() {
		return types.ResponseStatus_Error, fmt.Errorf("Voting must be initialized to set a signalling address. Use 'rocketpool pdao initialize-voting' to initialize voting first")
	}

	signatureSanitized, err := eip712.SanitizeEIP712String(c.signature)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error sanitizing signature input")
	}

	components, err := eip712.ParseEIP712Components(signatureSanitized)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error parsing EIP-712 components")
	}

	data.TxInfo, err = c.registry.SetSigner(c.signallingAddress, opts, components.V, components.R, components.S)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error getting the TX info for SetSigner: %w", err)
	}

	return types.ResponseStatus_Success, nil
}
