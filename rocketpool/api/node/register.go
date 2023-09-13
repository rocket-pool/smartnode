package node

import (
	"fmt"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type nodeRegisterHandler struct {
	timezoneLocation string
	pSettings        *settings.ProtocolDaoSettings
}

func (h *nodeRegisterHandler) CreateBindings(ctx *callContext) error {
	rp := ctx.rp
	var err error

	h.pSettings, err = settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error getting pDAO settings binding: %w", err)
	}
	return nil
}

func (h *nodeRegisterHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	node := ctx.node
	node.GetExists(mc)
	h.pSettings.GetNodeRegistrationEnabled(mc)
}

func (h *nodeRegisterHandler) PrepareResponse(ctx *callContext, response *api.NodeRegisterResponse) error {
	node := ctx.node
	opts := ctx.opts

	response.AlreadyRegistered = node.Details.Exists
	response.RegistrationDisabled = !h.pSettings.Details.Node.IsRegistrationEnabled
	response.CanRegister = !(response.AlreadyRegistered || response.RegistrationDisabled)
	if !response.CanRegister {
		return nil
	}

	// Get tx info
	txInfo, err := node.Register(h.timezoneLocation, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for Register: %w", err)
	}
	response.TxInfo = txInfo
	return nil
}
