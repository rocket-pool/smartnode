package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/config"
)


// Configure the Rocket Pool service
func configureService() error {

    // Get config paths
    rpPath := os.Getenv("RP_PATH")
    globalPath := filepath.Join(rpPath, "config.yml")
    userPath := filepath.Join(rpPath, "settings.yml")

    // Load config
    globalConfig, rpConfig, err := config.Load(globalPath, userPath)
    if err != nil { return err }

    // Configure chains
    if err := configureChain(&(rpConfig.Chains.Eth1), "Ethereum 1.0"); err != nil { return err }
    if err := configureChain(&(rpConfig.Chains.Eth2), "Ethereum 2.0"); err != nil { return err }

    // Update config
    if err := config.Save(userPath, globalConfig, rpConfig); err != nil { return err }

    // Log
    fmt.Println("Done! Run 'rocketpool service start' to apply new configuration settings.")
    fmt.Println("")

    // Return
    return nil

}


// Configure a chain
func configureChain(chain *config.Chain, chainName string) error {

    // Check client options
    if len(chain.Client.Options) == 0 {
        return errors.New(fmt.Sprintf("There are no available %s client options", chainName))
    }

    // Prompt for client
    clientOptions := []string{}
    for _, option := range chain.Client.Options { clientOptions = append(clientOptions, option.Name) }
    chain.Client.Selected = cli.PromptSelect(nil, nil, fmt.Sprintf("Which %s client would you like to run?", chainName), clientOptions)

    // Log
    fmt.Println(fmt.Sprintf("%s %s client selected.", chain.Client.Selected, chainName))
    fmt.Println("")

    // Prompt for params
    params := []config.UserParam{}
    for _, param := range chain.GetSelectedClient().Params {

        // Get expected param format
        var expectedFormat string
        if param.Regex != "" {
            expectedFormat = param.Regex
        } else if param.Required {
            expectedFormat = "^.+$"
        } else {
            expectedFormat = "^.*$"
        }

        // Optional field text
        optionalLabel := ""
        if !param.Required { optionalLabel = " (leave blank for none)" }

        // Prompt for value
        value := cli.Prompt(nil, nil, fmt.Sprintf("Please enter the %s%s", param.Name, optionalLabel), expectedFormat, fmt.Sprintf("Invalid %s", param.Name))
        fmt.Println("")

        // Add param
        params = append(params, config.UserParam{
            Env: param.Env,
            Value: value,
        })

    }
    chain.Client.Params = params

    // Return
    return nil

}

