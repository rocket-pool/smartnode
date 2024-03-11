package config

import (
	"fmt"
	"os"
	"path"

	"github.com/rocket-pool/smartnode/shared/types/config"
	"gopkg.in/yaml.v2"
)

const defaultNetworksConfigPath = "networks-default.yml"

type NetworksConfig struct {
	Networks []*config.NetworkInfo `yaml:"networks"`
}

func NewNetworksConfig() *NetworksConfig {
	return &NetworksConfig{
		Networks: make([]*config.NetworkInfo, 0),
	}
}

func LoadNetworksFromFile(configPath string) (*NetworksConfig, error) {
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config path does not exist: %s", configPath)
	}

	filePath := path.Join(configPath, defaultNetworksConfigPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("default networks file does not exist: %s", filePath)
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read default networks file: %w", err)
	}

	networks := NetworksConfig{}
	err = yaml.Unmarshal(bytes, &networks)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal default networks file: %w", err)
	}

	return &networks, nil
}

func (nc *NetworksConfig) GetNetworks() []*config.NetworkInfo {
	return nc.Networks
}

func (nc *NetworksConfig) SaveToFile(filePath string) error {
	// TODO: this signature doesn't support saving to default/extras files
	d, err := yaml.Marshal(nc)
	if err != nil {
		return fmt.Errorf("could not marshal networks config: %w", err)
	}

	err = os.WriteFile(filePath, d, 0644)
	if err != nil {
		return fmt.Errorf("could not write networks config to file to %s: %w", filePath, err)
	}
	return nil
}
