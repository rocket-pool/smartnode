package node

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getAlerts(c *cli.Context) (*api.NodeAlertsResponse, error) {
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	response := api.NodeAlertsResponse{}

	rawAlerts, err := alerting.FetchAlerts(cfg)
	if err != nil {
		// Don't fail the whole call — alertmanager may not be reachable in all setups.
		// Return an empty list so the CLI can still proceed.
		response.Alerts = []api.NodeAlert{}
		return &response, nil
	}

	response.Alerts = make([]api.NodeAlert, len(rawAlerts))
	for i, a := range rawAlerts {
		response.Alerts[i] = api.NodeAlert{
			State:       *a.Status.State,
			Labels:      a.Labels,
			Annotations: a.Annotations,
		}
	}

	return &response, nil
}
