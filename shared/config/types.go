package config

import "github.com/rocket-pool/node-manager-core/config"

type PBSubmissionRef int

// A MEV relay
type MevRelay struct {
	ID          MevRelayID
	Name        string
	Description string
	Urls        map[config.Network]string
	Regulated   bool
}
