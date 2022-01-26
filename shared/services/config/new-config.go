package config

type ContainerID int

type ParameterCategory int

const (
    Unknown ContainerID = iota
    Api
    Node
    Watchtower
    Eth1
    Eth2
    Validator
    Grafana
    Prometheus
    Exporter
)

type Parameter struct {
    Name string
    ID string
    Description string
    Default interface{}
    AffectsContainers []ContainerID
    EnvironmentVariable string
}


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

    // Network fee parameters
    ManualMaxFee *Parameter
    PriorityFee *Parameter
    RplClaimGasThreshold *Parameter
    MinipoolStakeGasThreshold *Parameter
}


// Smartnode Parameters
func NewSmartnodeConfig() *SmartnodeConfig {

    return &SmartnodeConfig{
        ProjectName: &Parameter{
            ID: "projectName",
            Name: "Project Name",
            Description: "This is the prefix that will be attached to all of the Docker containers managed by the Smartnode.",
            Default: "rocketpool",
            AffectsContainers: []ContainerID { Api, Node, Watchtower, Eth1, Eth2, Validator, Grafana, Prometheus, Exporter },
        },

        DataPath: &Parameter{
            ID: "passwordPath",
            Name: "Password Path",
            Description: "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
            Default: "$HOME/.rocketpool/data",
            AffectsContainers: []ContainerID { Api, Node, Watchtower, Validator },
        },

        ValidatorRestartCommand: &Parameter{
            ID: "validatorRestartCommand",
            Name: "Validator Restart Command",
            Description: "The absolute path to a custom script that will be invoked when Rocket Pool needs to restart your validator container to load the new key after a minipool is staked. **For Native mode only.**",
            Default: "$HOME/.rocketpool/chains/eth2/restart-validator.sh",
            AffectsContainers: []ContainerID { Node },
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
            AffectsContainers: []ContainerID { Node, Watchtower },
        },

        RplClaimGasThreshold: &Parameter{
            ID: "rplClaimGasThreshold",
            Name: "RPL Claim Gas Threshold",
            Description: "Automatic RPL rewards claims will use the `Rapid` suggestion from the gas estimator, based on current network conditions. This threshold is a limit (in gwei) you can put on that suggestion; your node will not try to claim RPL rewards automatically until the suggestion is below this limit.",
            Default: 150,
            AffectsContainers: []ContainerID { Node, Watchtower },
        },

        MinipoolStakeGasThreshold: &Parameter{
            ID: "minipoolStakeGasThreshold",
            Name: "Minipool Stake Gas Threshold",
            Description: "Once a newly created minipool passes the scrub check and is ready to perform its second 16 ETH deposit (the `stake` transaction), your node will try to do so automatically using the `Rapid` suggestion from the gas estimator as its max fee. This threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" + 
            "Note that to ensure your minipool does not get dissolved, the node will ignore this limit and automatically execute the `stake` transaction at whatever the suggested fee happens to be once too much time has passed since its first deposit (currently 7 days).",
            Default: 150,
            AffectsContainers: []ContainerID { Node },
        },
    }
    

}