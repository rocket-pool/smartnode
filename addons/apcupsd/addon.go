package apcupsd

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/addons"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	ContainerID_Apcupsd                 cfgtypes.ContainerID = "apcupsd"
	ContainerID_ApcupsdExporter         cfgtypes.ContainerID = "apcupsd_exporter"
	ApcupsdContainerName                string               = "addon_apcupsd"
	ApcupsdNetworkComposeTemplateName   string               = "addon_apcupsd.network"
	ApcupsdContainerComposeTemplateName string               = "addon_apcupsd.container"
	ApcupsdConfigTemplateName           string               = "addon_apcupsd_config"
	ApcupsdConfigName                   string               = "addon_apcupsd.conf"
)

type Apcupsd struct {
	cfg *ApcupsdConfig `yaml:"config,omitempty"`
}

func NewApcupsd() addons.SmartnodeAddon {
	return &Apcupsd{
		cfg: NewConfig(),
	}
}

func (apcupsd *Apcupsd) GetName() string {
	return "APCUPS Monitor"
}

func (apcupsd *Apcupsd) GetDescription() string {
	return "This addon adds UPS monitoring to your node so you can monitor the status of your APCUPSD compatible UPS within grafana \n\nMade with love by killjoy.eth."
}

func (apcupsd *Apcupsd) GetConfig() cfgtypes.Config {
	return apcupsd.cfg
}

func (apcupsd *Apcupsd) GetContainerName() string {
	return fmt.Sprint(ContainerID_Apcupsd)
}

func (apcupsd *Apcupsd) GetEnabledParameter() *cfgtypes.Parameter {
	return &apcupsd.cfg.Enabled
}

func (apcupsd *Apcupsd) GetContainerTag() string {
	return containerTag
}

func (apcupsd *Apcupsd) UpdateEnvVars(envVars map[string]string) error {
	if apcupsd.cfg.Enabled.Value == true {
		cfgtypes.AddParametersToEnvVars(apcupsd.cfg.GetParameters(), envVars)
	}
	return nil
}
