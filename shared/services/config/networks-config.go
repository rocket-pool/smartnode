package config

import (
	"fmt"
	"os"
	"path"

	"github.com/rocket-pool/smartnode/shared/types/config"
	"gopkg.in/yaml.v2"
)

const (
	defaultNetworksConfigPath = "networks-default.yml"
	extraNetworksConfigPath   = "networks-extra.yml"
)

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
	defaultNetworks, err := loadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not load default networks file: %w", err)
	}

	filePath = path.Join(configPath, extraNetworksConfigPath)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		// if the file didn't exist, we just use the default networks
		return defaultNetworks, nil
	}
	extraNetworks, err := loadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not load extra networks file: %w", err)
	}
	defaultNetworks.Networks = append(defaultNetworks.Networks, extraNetworks.Networks...)

	return defaultNetworks, nil
}

func loadFile(filePath string) (*NetworksConfig, error) {
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
