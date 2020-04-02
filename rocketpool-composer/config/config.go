package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    //cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    configutils "github.com/rocket-pool/smartnode/shared/utils/config"
)


// Configure the Rocket Pool service
func configureService() error {

    // Get config paths
    rpPath := os.Getenv("RP_PATH")
    globalPath := filepath.Join(rpPath, "config.yml")
    userPath := filepath.Join(rpPath, "settings.yml")

    // Load config
    globalConfig, rpConfig, err := configutils.Load(globalPath, userPath)
    if err != nil { return err }

    // Check config options
    if len(rpConfig.Chains.Eth1.Client.Options) == 0 {
        return errors.New("There are no available Eth1 client options.")
    }
    if len(rpConfig.Chains.Eth2.Client.Options) == 0 {
        return errors.New("There are no available Eth2 client options.")
    }

    // Select some shit
    rpConfig.Chains.Eth1.Client.Selected = "Infura"
    rpConfig.Chains.Eth2.Client.Selected = "Lighthouse"

    // Update config
    if err := configutils.Save(userPath, globalConfig, rpConfig); err != nil { return err }

    /*
    // Prompt for eth1 client
    eth1ClientOptions := []string{}
    for _, option := range rpConfig.Chains.Eth1.Client.Options { eth1ClientOptions = append(eth1ClientOptions, option.Name) }
    eth1Client := cliutils.PromptSelect(nil, nil, "Which ethereum 1.0 client would you like to run?", eth1ClientOptions)
    rpConfig.Chains.Eth1.Client.Selected = eth1Client

    // Log
    fmt.Println(fmt.Sprintf("%s ethereum 1.0 client selected.", eth1Client))
    fmt.Println("")

    // Prompt for eth1 client params
    eth1Params := []string{}
    for _, param := range rpConfig.GetSelectedEth1Client().Params {
        var value string
        switch param {
            case "ETHSTATS_LABEL":    value = cliutils.Prompt(nil, nil, "Please enter your ethstats label (or leave blank for none)", "^.*$", "Invalid ethstats label")
            case "ETHSTATS_LOGIN":    value = cliutils.Prompt(nil, nil, "Please enter your ethstats login (or leave blank for none)", "^.*$", "Invalid ethstats login")
            case "INFURA_PROJECT_ID": value = cliutils.Prompt(nil, nil, "Please enter your Infura project ID", "^[0-9a-fA-F]{32}$", "Invalid Infura project ID")
        }
        eth1Params = append(eth1Params, fmt.Sprintf("%s=%s", param, value))
        fmt.Println("")
    }
    rpConfig.Chains.Eth1.Client.Params = eth1Params

    // Prompt for eth2 client
    eth2ClientOptions := []string{}
    for _, option := range rpConfig.Chains.Eth2.Client.Options { eth2ClientOptions = append(eth2ClientOptions, option.Name) }
    eth2Client := cliutils.PromptSelect(nil, nil, "Which ethereum 2.0 client would you like to run?", eth2ClientOptions)
    rpConfig.Chains.Eth2.Client.Selected = eth2Client

    // Log
    fmt.Println(fmt.Sprintf("%s ethereum 2.0 client selected.", eth2Client))
    fmt.Println("")

    // Prompt for eth2 client params
    eth2Params := []string{}
    for _, param := range rpConfig.GetSelectedEth2Client().Params {
        var value string
        eth2Params = append(eth2Params, fmt.Sprintf("%s=%s", param, value))
        fmt.Println("")
    }
    rpConfig.Chains.Eth2.Client.Params = eth2Params

    // Update config
    if err := configutils.Save(rpPath, rpConfig); err != nil { return err }
    */

    // Log
    fmt.Println("Done! Run 'rocketpool service start' to apply new configuration settings.")

    // Return
    return nil

}

