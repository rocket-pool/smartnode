package alerting

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	apiclient "github.com/rocket-pool/smartnode/shared/services/alerting/alertmanager/client"
	"github.com/rocket-pool/smartnode/shared/services/alerting/alertmanager/models"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// fetches the current alerts directly the alertmanager container/application's API.
func FetchAlerts(cfg *config.RocketPoolConfig) ([]*models.GettableAlert, error) {

	if cfg.EnableMetrics.Value == false {
		// metrics are disabled, so no alerts will be fetched.
		return make([]*models.GettableAlert, 0), nil
	}

	// create the transport
	host := fmt.Sprintf("%s:%d", config.AlertmanagerContainerName, cfg.Alertmanager.Port.Value)
	transport := apiclient.DefaultTransportConfig().WithHost(host)
	client := apiclient.NewHTTPClientWithConfig(strfmt.Default, transport)

	// request to get alerts:
	resp, err := client.Alert.GetAlerts(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting alerts: %w", err)
	}
	return resp.Payload, nil
}
