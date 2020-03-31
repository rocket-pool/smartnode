package config

import (
    "errors"
    "fmt"
    "os"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    configutils "github.com/rocket-pool/smartnode/shared/utils/config"
)


// Configure the Rocket Pool service
func configureService() error {

    // Load config
    rpPath := os.Getenv("RP_PATH")
    rpConfig, err := configutils.Load(rpPath)
    if err != nil { return err }

    // Check config options
    if len(rpConfig.Chains.Eth1.Client.Options) == 0 {
        return errors.New("There are no available Eth1 client options.")
    }
    if len(rpConfig.Chains.Eth2.Client.Options) == 0 {
        return errors.New("There are no available Eth2 client options.")
    }

    // Prompt for eth1 client
    eth1ClientOptions := []string{}
    for _, option := range rpConfig.Chains.Eth1.Client.Options { eth1ClientOptions = append(eth1ClientOptions, option.Name) }
    eth1Client := cliutils.PromptSelect(nil, nil, "Which ethereum 1.0 client would you like to run?", eth1ClientOptions)

    // Log
    fmt.Println(fmt.Sprintf("%s ethereum 1.0 client selected.", eth1Client))
    fmt.Println("")

    // Prompt for eth2 client
    eth2ClientOptions := []string{}
    for _, option := range rpConfig.Chains.Eth2.Client.Options { eth2ClientOptions = append(eth2ClientOptions, option.Name) }
    eth2Client := cliutils.PromptSelect(nil, nil, "Which ethereum 2.0 client would you like to run?", eth2ClientOptions)

    // Log
    fmt.Println(fmt.Sprintf("%s ethereum 2.0 client selected.", eth2Client))
    fmt.Println("")

    // Update config
    rpConfig.Chains.Eth1.Client.Selected = eth1Client
    rpConfig.Chains.Eth2.Client.Selected = eth2Client
    if err := configutils.Save(rpPath, rpConfig); err != nil { return err }

    // Log
    fmt.Println("Done! Run 'rocketpool service start' to apply new configuration settings.")

    // Return
    return nil

}

