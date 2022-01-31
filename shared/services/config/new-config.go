package config

import (
	"runtime"

	"github.com/pbnjay/memory"
)

type ContainerID int
type Network int
type ParameterType int


// Enum to describe which container(s) a parameter impacts, so the Smartnode knows which
// ones to restart upon a settings change
const (
    ContainerID_Unknown ContainerID = iota
    ContainerID_Api
    ContainerID_Node
    ContainerID_Watchtower
    ContainerID_Eth1
    ContainerID_Eth2
    ContainerID_Validator
    ContainerID_Grafana
    ContainerID_Prometheus
    ContainerID_Exporter
)


// Enum to describe which network the system is on
const (
    Network_Unknown Network = iota
    Network_Mainnet
    Network_Prater
)


// Enum to describe which data type a parameter's value will have, which
// informs the corresponding UI element and value validation
const (
    ParameterType_Unknown ParameterType = iota
    ParameterType_Int
    ParameterType_Uint16
    ParameterType_String
    ParameterType_Bool
    ParameterType_Choice
)


// A parameter that can be configured by the user
type Parameter struct {
    Name string
    ID string
    Description string
    Type ParameterType
    Default interface{}
    AffectsContainers []ContainerID
    EnvironmentVariable string
    Required bool
}


// The value for a parameter
type Setting struct {
    Parameter *Parameter
    Value interface{}
    UsingDefault bool
}


// Configuration for the Smartnode itself
type SmartnodeConfig struct {
    // Smartnode parameters
    ProjectName *Parameter
    DataPath *Parameter
    ValidatorRestartCommand *Parameter
    Network *Parameter

    // Network fee parameters
    ManualMaxFee *Parameter
    PriorityFee *Parameter
    RplClaimGasThreshold *Parameter
    MinipoolStakeGasThreshold *Parameter
}


// Configuration for the Execution client
type ExecutionConfig struct {
    ReconnectDelay *Parameter

    // External clients (Hybrid mode)
    UseExternalClient *Parameter
    ExternalClientHttpUrl *Parameter
    ExternalClientWsUrl *Parameter

    // Local clients (Docker mode)
    Client *Parameter
    ClientConfig interface{}
}


// Configuration for Geth
type GethConfig struct {
    EthstatsLabel *Parameter
    EthstatsLogin *Parameter
    CacheSize *Parameter
    MaxPeers *Parameter
    P2pPort *Parameter
    HttpPort *Parameter
    WsPort *Parameter
    OpenRpcPorts *Parameter
    ContainerName *Parameter
    AdditionalFlags *Parameter
}


// Configuration for Infura
type InfuraConfig struct {
    ProjectID *Parameter
    HttpPort *Parameter
    WsPort *Parameter
    OpenRpcPorts *Parameter
}


// Configuration for Pocket
type PocketConfig struct {
    GatewayID *Parameter
    HttpPort *Parameter
    OpenRpcPorts *Parameter
}


// Configuration for Lighthouse
type LighthouseConfig struct {
    Graffiti *Parameter
    CheckpointSyncProvider *Parameter
    TargetPeers *Parameter
    P2pPort *Parameter
    ApiPort *Parameter
    OpenApiPort *Parameter
    ContainerName *Parameter
    AdditionalBnFlags *Parameter
    AdditionalVcFlags *Parameter
}


// Configuration for Nimbus
type NimbusConfig struct {
    Graffiti *Parameter
    MaxPeers *Parameter
    P2pPort *Parameter
    ApiPort *Parameter
    OpenApiPort *Parameter
    ContainerName *Parameter
    AdditionalFlags *Parameter
}


// Configuration for Prysm
type PrysmConfig struct {
    Graffiti *Parameter
    TargetPeers *Parameter
    P2pPort *Parameter
    HttpPort *Parameter
    RpcPort *Parameter
    OpenApiPorts *Parameter
    ContainerName *Parameter
    AdditionalBnFlags *Parameter
    AdditionalVcFlags *Parameter
}


// Configuration for Teku
type TekuConfig struct {
    Graffiti *Parameter
    CheckpointSyncProvider *Parameter
    MaxPeers *Parameter
    P2pPort *Parameter
    ApiPort *Parameter
    OpenApiPort *Parameter
    ContainerName *Parameter
    AdditionalBnFlags *Parameter
    AdditionalVcFlags *Parameter
}


// Configuration for Grafana
type GrafanaConfig struct {
    Port *Parameter
    ContainerName *Parameter
    AdditionalFlags *Parameter
}


// Configuration for Prometheus
type PrometheusConfig struct {
    Port *Parameter
    OpenPort *Parameter
    ContainerName *Parameter
    AdditionalFlags *Parameter
}


// Configuration for Exporter
type ExporterConfig struct {
    RootFs *Parameter
    Port *Parameter
    ContainerName *Parameter
    AdditionalFlags *Parameter
}


// Generates a new Smartnode configuration
func NewSmartnodeConfig() *SmartnodeConfig {

    return &SmartnodeConfig{
        ProjectName: &Parameter{
            ID: "projectName",
            Name: "Project Name",
            Description: "This is the prefix that will be attached to all of the Docker containers managed by the Smartnode.",
            Type: ParameterType_String,
            Default: "rocketpool",
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2, ContainerID_Validator, ContainerID_Grafana, ContainerID_Prometheus, ContainerID_Exporter },
        },

        DataPath: &Parameter{
            ID: "passwordPath",
            Name: "Password Path",
            Description: "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
            Type: ParameterType_String,
            Default: "$HOME/.rocketpool/data",
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Validator },
        },

        ValidatorRestartCommand: &Parameter{
            ID: "validatorRestartCommand",
            Name: "Validator Restart Command",
            Description: "The absolute path to a custom script that will be invoked when Rocket Pool needs to restart your validator container to load the new key after a minipool is staked. **For Native mode only.**",
            Type: ParameterType_String,
            Default: "$HOME/.rocketpool/chains/eth2/restart-validator.sh",
            AffectsContainers: []ContainerID { ContainerID_Node },
        },

        Network: &Parameter{
            ID: "network",
            Name: "Network",
            Description: "The Ethereum network you want to use - select Prater Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
            Type: ParameterType_Choice,
            Default: "",
            Required: true,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2, ContainerID_Validator },
        },

        ManualMaxFee: &Parameter{
            ID: "manualMaxFee",
            Name: "Manual Max Fee",
            Description: "Set this if you want all of the Smartnode's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*). This will ignore the recommended max fee based on the current network conditions, and explicitly use this value instead. This applies to automated transactions (such as claiming RPL and staking minipools) as well.",
            Default: 0,
        },

        PriorityFee: &Parameter{
            ID: "priorityFee",
            Name: "Priority Fee",
            Description: "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the miners for including your transaction, which generally means it will be mined faster (as long as your max fee is sufficiently high to cover the current network conditions).",
            Default: 2,
            AffectsContainers: []ContainerID { ContainerID_Node, ContainerID_Watchtower },
        },

        RplClaimGasThreshold: &Parameter{
            ID: "rplClaimGasThreshold",
            Name: "RPL Claim Gas Threshold",
            Description: "Automatic RPL rewards claims will use the `Rapid` suggestion from the gas estimator, based on current network conditions. This threshold is a limit (in gwei) you can put on that suggestion; your node will not try to claim RPL rewards automatically until the suggestion is below this limit.",
            Default: 150,
            AffectsContainers: []ContainerID { ContainerID_Node, ContainerID_Watchtower },
        },

        MinipoolStakeGasThreshold: &Parameter{
            ID: "minipoolStakeGasThreshold",
            Name: "Minipool Stake Gas Threshold",
            Description: "Once a newly created minipool passes the scrub check and is ready to perform its second 16 ETH deposit (the `stake` transaction), your node will try to do so automatically using the `Rapid` suggestion from the gas estimator as its max fee. This threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" + 
            "Note that to ensure your minipool does not get dissolved, the node will ignore this limit and automatically execute the `stake` transaction at whatever the suggested fee happens to be once too much time has passed since its first deposit (currently 7 days).",
            Default: 150,
            AffectsContainers: []ContainerID { ContainerID_Node },
        },
    }
    
}


// Generates a new Geth configuration
func NewGethConfig() *GethConfig {
    return &GethConfig{
        EthstatsLabel: &Parameter{
            ID: "ethstatsLabel",
            Name: "ETHStats Label",
            Description: "If you would like to report your Execution client statistics to https://ethstats.net/, enter the label you want to use here.",
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "ETHSTATS_LABEL",
        },

        EthstatsLogin: &Parameter{
            ID: "ethstatsLogin",
            Name: "ETHStats Login",
            Description: "If you would like to report your Execution client statistics to https://ethstats.net/, enter the login you want to use here.",
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "ETHSTATS_LOGIN",
        },

        CacheSize: &Parameter{
            ID: "cache",
            Name: "Cache Size",
            Description: "The amount of RAM (in MB) you want Geth's cache to use. Larger values mean your disk space usage will increase slower, and you will have to prune less frequently. The default is based on how much total RAM your system has but you can adjust it manually.",
            Default: calculateGethCache(),
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "GETH_CACHE_SIZE",
        },

        MaxPeers: &Parameter{
            ID: "maxPeers",
            Name: "Max Peers",
            Description: "The maximum number of peers Geth should connect to. This can be lowered to improve performance on low-power systems or constrained networks. We recommend keeping it at 12 or higher.",
            Default: calculateGethPeers(),
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "GETH_MAX_PEERS",
        },

        P2pPort: &Parameter{
            ID: "p2pPort",
            Name: "P2P Port",
            Description: "The port Geth should use for P2P (blockchain) traffic to communicate with other nodes.",
            Default: 30303,
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "EC_P2P_PORT",
            Required: true,
        },

        HttpPort: &Parameter{
            ID: "httpPort",
            Name: "HTTP Port",
            Description: "The port Geth should use for its HTTP RPC endpoint.",
            Default: 8545,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2 },
            EnvironmentVariable: "EC_HTTP_PORT",
            Required: true,
        },

        WsPort: &Parameter{
            ID: "wsPort",
            Name: "Websocket Port",
            Description: "The port Geth should use for its Websocket RPC endpoint.",
            Default: 8546,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2 },
            EnvironmentVariable: "EC_WS_PORT",
            Required: true,
        },

        OpenRpcPorts: &Parameter{
            ID: "openRpcPorts",
            Name: "Open RPC Ports",
            Description: "Open the HTTP and Websocket RPC ports to your local network, so other local machines can access your Execution Client's RPC endpoint.",
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "EC_OPEN_RPC_PORTS",
            Required: false,
        },

        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Geth container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "ethereum/client-go:v1.10.15",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            Required: true,
        },

        AdditionalFlags: &Parameter{
            ID: "additionalFlags",
            Name: "Additional Flags",
            Description: "Additional custom command line flags you want to pass to Geth, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "EC_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Infura configuration
func NewInfuraConfig() *InfuraConfig {
    return &InfuraConfig{
        ProjectID: &Parameter{
            ID: "projectID",
            Name: "Project ID",
            Description: "The ID of your `Ethereum` project in Infura. Note: This is your Project ID, not your Project Secret!",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "INFURA_PROJECT_ID",
            Required: true,
        },

        HttpPort: &Parameter{
            ID: "httpPort",
            Name: "HTTP Port",
            Description: "The port the Infura proxy should use for its HTTP RPC endpoint.",
            Default: 8545,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2 },
            EnvironmentVariable: "EC_HTTP_PORT",
            Required: true,
        },

        WsPort: &Parameter{
            ID: "wsPort",
            Name: "Websocket Port",
            Description: "The port the Infura proxy should use for its Websocket RPC endpoint.",
            Default: 8546,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2 },
            EnvironmentVariable: "EC_WS_PORT",
            Required: true,
        },

        OpenRpcPorts: &Parameter{
            ID: "openRpcPorts",
            Name: "Open RPC Ports",
            Description: "Open the HTTP and Websocket RPC ports to your local network, so other local machines can access the Infura proxy's RPC endpoint.",
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "EC_OPEN_RPC_PORTS",
            Required: false,
        },
    }
}


// Generates a new Pocket configuration
func NewPocketConfig() *PocketConfig {
    return &PocketConfig{
        GatewayID: &Parameter{
            ID: "gatewayID",
            Name: "Gateway ID",
            Description: "If you would like to use a custom gateway for Pocket instead of the default Rocket Pool gateway, enter it here.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "Pocket_PROJECT_ID",
            Required: true,
        },

        HttpPort: &Parameter{
            ID: "httpPort",
            Name: "HTTP Port",
            Description: "The port the Pocket proxy should use for its HTTP RPC endpoint.",
            Default: 8545,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2 },
            EnvironmentVariable: "EC_HTTP_PORT",
            Required: true,
        },

        OpenRpcPorts: &Parameter{
            ID: "openRpcPorts",
            Name: "Open RPC Ports",
            Description: "Open the HTTP RPC port to your local network, so other local machines can access the Pocket proxy's RPC endpoint.",
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "EC_OPEN_RPC_PORTS",
            Required: false,
        },
    }
}


// Generates a new Lighthouse configuration
func NewLighthouseConfig() *LighthouseConfig {
    return &LighthouseConfig{
        Graffiti: &Parameter{
            ID: "graffiti",
            Name: "Custom Graffiti",
            Description: "Add a short message to any blocks you propose, so the world can see what you have to say! It has a 16 character limit.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "CUSTOM_GRAFFITI",
            Required: false,
        },

        CheckpointSyncProvider: &Parameter{
            ID: "checkpointSyncUrl",
            Name: "Checkpoint Sync URL",
            Description: "If you would like to instantly sync using an existing Beacon node, enter its URL.\n" +
                "Example: https://<project ID>:<secret>@eth2-beacon-prater.infura.io\n" +
                "Leave this blank if you want to sync normally from the start of the chain.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "CHECKPOINT_SYNC_URL",
            Required: false,
        },

        TargetPeers: &Parameter{
            ID: "targetPeers",
            Name: "Target Peers",
            Description: "The number of peers Lighthouse should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
            Type: ParameterType_Uint16,
            Default: 50,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_MAX_PEERS",
            Required: true,
        },

        P2pPort: &Parameter{
            ID: "p2pPort",
            Name: "P2P Port",
            Description: "The port to use for P2P (blockchain) traffic.",
            Type: ParameterType_Uint16,
            Default: 9001,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_P2P_PORT",
            Required: true,
        },

        ApiPort: &Parameter{
            ID: "apiPort",
            Name: "HTTP API Port",
            Description: "The port Lighthouse should run its HTTP API on.",
            Type: ParameterType_Uint16,
            Default: 5052,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus },
            EnvironmentVariable: "BN_API_PORT",
            Required: true,
        },

        OpenApiPort: &Parameter{
            ID: "openApiPort",
            Name: "Open API Port",
            Description: "Enable this to open Lighthouse's API port to your local network, so other machines can access it too.",
            Type: ParameterType_Bool,
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_OPEN_API_PORT",
        },

        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Lighthouse container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "sigp/lighthouse:v2.1.1",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            Required: true,
        },

        AdditionalBnFlags: &Parameter{
            ID: "additionalBnFlags",
            Name: "Additional Beacon Client Flags",
            Description: "Additional custom command line flags you want to pass Lighthouse's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_ADDITIONAL_FLAGS",
            Required: false,
        },

        AdditionalVcFlags: &Parameter{
            ID: "additionalVcFlags",
            Name: "Additional Validator Client Flags",
            Description: "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "VC_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Nimbus configuration
func NewNimbusConfig() *NimbusConfig {
    return &NimbusConfig{
        Graffiti: &Parameter{
            ID: "graffiti",
            Name: "Custom Graffiti",
            Description: "Add a short message to any blocks you propose, so the world can see what you have to say! It has a 16 character limit.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "CUSTOM_GRAFFITI",
            Required: false,
        },

        MaxPeers: &Parameter{
            ID: "maxPeers",
            Name: "Max Peers",
            Description: "The maximum number of peers Nimbus should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
            Type: ParameterType_Uint16,
            Default: 50,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_MAX_PEERS",
            Required: true,
        },

        P2pPort: &Parameter{
            ID: "p2pPort",
            Name: "P2P Port",
            Description: "The port to use for P2P (blockchain) traffic.",
            Type: ParameterType_Uint16,
            Default: 9001,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_P2P_PORT",
            Required: true,
        },

        ApiPort: &Parameter{
            ID: "apiPort",
            Name: "HTTP API Port",
            Description: "The port Nimbus should run its HTTP API on.",
            Type: ParameterType_Uint16,
            Default: 5052,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus },
            EnvironmentVariable: "BN_API_PORT",
            Required: true,
        },

        OpenApiPort: &Parameter{
            ID: "openApiPort",
            Name: "Open API Port",
            Description: "Enable this to open Nimbus's API port to your local network, so other machines can access it too.",
            Type: ParameterType_Bool,
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_OPEN_API_PORT",
        },

        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Nimbus container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "statusim/nimbus-eth2:multiarch-v1.6.0",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            Required: true,
        },

        AdditionalFlags: &Parameter{
            ID: "additionalFlags",
            Name: "Additional Flags",
            Description: "Additional custom command line flags you want to pass to Nimbus, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth1 },
            EnvironmentVariable: "BN_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Prysm configuration
func NewPrysmConfig() *PrysmConfig {
    return &PrysmConfig{
        Graffiti: &Parameter{
            ID: "graffiti",
            Name: "Custom Graffiti",
            Description: "Add a short message to any blocks you propose, so the world can see what you have to say! It has a 16 character limit.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "CUSTOM_GRAFFITI",
            Required: false,
        },

        TargetPeers: &Parameter{
            ID: "targetPeers",
            Name: "Target Peers",
            Description: "The number of peers Prysm should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
            Type: ParameterType_Uint16,
            Default: 45,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_MAX_PEERS",
            Required: true,
        },

        P2pPort: &Parameter{
            ID: "p2pPort",
            Name: "P2P Port",
            Description: "The port to use for P2P (blockchain) traffic.",
            Type: ParameterType_Uint16,
            Default: 9001,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_P2P_PORT",
            Required: true,
        },

        HttpPort: &Parameter{
            ID: "httpPort",
            Name: "HTTP Port",
            Description: "The port Prysm should run its HTTP (gRPC) API on.",
            Type: ParameterType_Uint16,
            Default: 5052,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus },
            EnvironmentVariable: "BN_HTTP_PORT",
            Required: true,
        },

        RpcPort: &Parameter{
            ID: "rpcPort",
            Name: "RPC Port",
            Description: "The port Prysm should run its JSON-RPC API on.",
            Type: ParameterType_Uint16,
            Default: 5053,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus },
            EnvironmentVariable: "BN_RPC_PORT",
            Required: true,
        },

        OpenApiPorts: &Parameter{
            ID: "openApiPorts",
            Name: "Open API Ports",
            Description: "Enable this to open Prysm's API ports to your local network, so other machines can access it too.",
            Type: ParameterType_Bool,
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_OPEN_API_PORTS",
        },

        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Prysm container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "prysmaticlabs/prysm-beacon-chain:HEAD-d6338f-debug",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            Required: true,
        },

        AdditionalBnFlags: &Parameter{
            ID: "additionalBnFlags",
            Name: "Additional Beacon Client Flags",
            Description: "Additional custom command line flags you want to pass Prysm's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_ADDITIONAL_FLAGS",
            Required: false,
        },

        AdditionalVcFlags: &Parameter{
            ID: "additionalVcFlags",
            Name: "Additional Validator Client Flags",
            Description: "Additional custom command line flags you want to pass Prysm's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "VC_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Teku configuration
func NewTekuConfig() *TekuConfig {
    return &TekuConfig{
        Graffiti: &Parameter{
            ID: "graffiti",
            Name: "Custom Graffiti",
            Description: "Add a short message to any blocks you propose, so the world can see what you have to say! It has a 16 character limit.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "CUSTOM_GRAFFITI",
            Required: false,
        },

        CheckpointSyncProvider: &Parameter{
            ID: "checkpointSyncUrl",
            Name: "Checkpoint Sync URL",
            Description: "If you would like to instantly sync using an existing Beacon node, enter its URL.\n" +
                "Example: https://<project ID>:<secret>@eth2-beacon-prater.infura.io\n" +
                "Leave this blank if you want to sync normally from the start of the chain.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "CHECKPOINT_SYNC_URL",
            Required: false,
        },

        MaxPeers: &Parameter{
            ID: "maxPeers",
            Name: "Max Peers",
            Description: "The maximum number of peers Teku should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
            Type: ParameterType_Uint16,
            Default: 74,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_MAX_PEERS",
            Required: true,
        },

        P2pPort: &Parameter{
            ID: "p2pPort",
            Name: "P2P Port",
            Description: "The port to use for P2P (blockchain) traffic.",
            Type: ParameterType_Uint16,
            Default: 9001,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_P2P_PORT",
            Required: true,
        },

        ApiPort: &Parameter{
            ID: "apiPort",
            Name: "HTTP API Port",
            Description: "The port Teku should run its HTTP API on.",
            Type: ParameterType_Uint16,
            Default: 5052,
            AffectsContainers: []ContainerID { ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus },
            EnvironmentVariable: "BN_API_PORT",
            Required: true,
        },

        OpenApiPort: &Parameter{
            ID: "openApiPort",
            Name: "Open API Port",
            Description: "Enable this to open Teku's API port to your local network, so other machines can access it too.",
            Type: ParameterType_Bool,
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_OPEN_API_PORT",
        },

        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Teku container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "consensys/teku:22.1.0",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            Required: true,
        },

        AdditionalBnFlags: &Parameter{
            ID: "additionalBnFlags",
            Name: "Additional Beacon Client Flags",
            Description: "Additional custom command line flags you want to pass Teku's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Eth2 },
            EnvironmentVariable: "BN_ADDITIONAL_FLAGS",
            Required: false,
        },

        AdditionalVcFlags: &Parameter{
            ID: "additionalVcFlags",
            Name: "Additional Validator Client Flags",
            Description: "Additional custom command line flags you want to pass Teku's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Validator },
            EnvironmentVariable: "VC_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Grafana config
func NewGrafanaConfig() *GrafanaConfig {
    return &GrafanaConfig{
        Port: &Parameter{
            ID: "port",
            Name: "HTTP Port",
            Description: "The port Grafana should run its HTTP server on - this is the port you will connect to in your browser.",
            Type: ParameterType_Uint16,
            Default: 3100,
            AffectsContainers: []ContainerID { ContainerID_Grafana },
            EnvironmentVariable: "GRAFANA_PORT",
            Required: true,
        },
        
        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Grafana container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "grafana/grafana:8.3.2",
            AffectsContainers: []ContainerID { ContainerID_Grafana },
            Required: true,
        },

        AdditionalFlags: &Parameter{
            ID: "additionalFlags",
            Name: "Additional Flags",
            Description: "Additional custom command line flags you want to pass to Grafana, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Grafana },
            EnvironmentVariable: "GRAFANA_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Prometheus config
func NewPrometheusConfig() *PrometheusConfig {
    return &PrometheusConfig{
        Port: &Parameter{
            ID: "port",
            Name: "API Port",
            Description: "The port Prometheus should make its statistics available on.",
            Type: ParameterType_Uint16,
            Default: 9091,
            AffectsContainers: []ContainerID { ContainerID_Prometheus },
            EnvironmentVariable: "PROMETHEUS_PORT",
            Required: true,
        },

        OpenPort: &Parameter{
            ID: "openPort",
            Name: "Open Port",
            Description: "Enable this to open Prometheus's port to your local network, so other machines can access it too.",
            Type: ParameterType_Bool,
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Prometheus },
            EnvironmentVariable: "PROMETHEUS_PORT",
        },
        
        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Prometheus container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "prom/prometheus:v2.31.1",
            AffectsContainers: []ContainerID { ContainerID_Prometheus },
            Required: true,
        },

        AdditionalFlags: &Parameter{
            ID: "additionalFlags",
            Name: "Additional Flags",
            Description: "Additional custom command line flags you want to pass to Prometheus, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Prometheus },
            EnvironmentVariable: "PROMETHEUS_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Generates a new Exporter config
func NewExporterConfig() *ExporterConfig {
    return &ExporterConfig{
        RootFs: &Parameter{
            ID: "enableRootFs",
            Name: "Allow Root Filesystem Access",
            Description: "Give the exporter permission to view your root filesystem instead of being limited to its own Docker container.\nThis is needed if you want the Grafana dashboard to report the used disk space of a second SSD.",
            Type: ParameterType_Bool,
            Default: false,
            AffectsContainers: []ContainerID { ContainerID_Exporter },
            EnvironmentVariable: "EXPORTER_ROOT_FS",
            Required: false,
        },

        Port: &Parameter{
            ID: "port",
            Name: "API Port",
            Description: "The port the Exporter should make its statistics available on.",
            Type: ParameterType_Uint16,
            Default: 9103,
            AffectsContainers: []ContainerID { ContainerID_Exporter },
            EnvironmentVariable: "EXPORTER_PORT",
            Required: true,
        },
        
        ContainerName: &Parameter{
            ID: "containerName",
            Name: "Container Name",
            Description: "The tag name of the Exporter container you want to use on Docker hub.",
            Type: ParameterType_String,
            Default: "prom/node-exporter:v1.3.1",
            AffectsContainers: []ContainerID { ContainerID_Exporter },
            Required: true,
        },

        AdditionalFlags: &Parameter{
            ID: "additionalFlags",
            Name: "Additional Flags",
            Description: "Additional custom command line flags you want to pass to the Exporter, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
            Type: ParameterType_String,
            Default: "",
            AffectsContainers: []ContainerID { ContainerID_Exporter },
            EnvironmentVariable: "EXPORTER_ADDITIONAL_FLAGS",
            Required: false,
        },
    }
}


// Calculate the recommended size for Geth's cache based on the amount of system RAM
func calculateGethCache() uint64 {
    totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024

    if totalMemoryGB == 0 {
        return 0
    } else if totalMemoryGB < 9 {
        return 256
    } else if totalMemoryGB < 13 {
        return 2048
    } else if totalMemoryGB < 17 {
        return 4096
    } else if totalMemoryGB < 25 {
        return 8192
    } else if totalMemoryGB < 33 {
        return 12288
    } else {
        return 16384
    }
}


// Calculate the default number of Geth peers
func calculateGethPeers() int {
    if runtime.GOARCH == "arm64" {
        return 25
    }
    return 50
}