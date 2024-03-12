package config

import "github.com/rocket-pool/node-manager-core/config"

type RewardsMode string
type MevRelayID string
type MevSelectionMode string

const (
	// ID for the Watchtower
	ContainerID_Watchtower config.ContainerID = "watchtower"

	// Rocket Pool networks
	Network_Devnet config.Network = "devnet"
)

// Enum to describe the rewards tree acquisition modes
const (
	RewardsMode_Unknown  RewardsMode = ""
	RewardsMode_Download RewardsMode = "download"
	RewardsMode_Generate RewardsMode = "generate"
)

// Enum to identify MEV-boost relays
const (
	MevRelayID_Unknown            MevRelayID = ""
	MevRelayID_Flashbots          MevRelayID = "flashbots"
	MevRelayID_BloxrouteEthical   MevRelayID = "bloxrouteEthical"
	MevRelayID_BloxrouteMaxProfit MevRelayID = "bloxrouteMaxProfit"
	MevRelayID_BloxrouteRegulated MevRelayID = "bloxrouteRegulated"
	MevRelayID_Eden               MevRelayID = "eden"
	MevRelayID_Ultrasound         MevRelayID = "ultrasound"
	MevRelayID_Aestus             MevRelayID = "aestus"
)

// Enum to describe MEV-Boost relay selection mode
const (
	MevSelectionMode_Profile MevSelectionMode = "profile"
	MevSelectionMode_Relay   MevSelectionMode = "relay"
)
