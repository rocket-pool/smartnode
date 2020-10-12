package service

import (
    "fmt"
    "math/rand"
    "time"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Configure the Rocket Pool service
func configureService(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Load global config
    globalConfig, err := rp.LoadGlobalConfig()
    if err != nil {
        return err
    }

    // Initialize user config
    userConfig := config.RocketPoolConfig{}

    // Log
    fmt.Println("Configuring the Rocket Pool service...")
    fmt.Println("* Using the Infura Eth 1.0 client with a free account may be unstable, and is not recommended in production.")
    fmt.Println("* The Eth 1.0 Ethstats configuration is for reporting node status to ethstats.net, and is optional.")
    fmt.Println("* The Eth 2.0 Custom Graffiti configuration is for adding custom text to signed blocks, and is optional.")
    fmt.Println("")

    // Configure chains
    if err := configureChain(&(globalConfig.Chains.Eth1), &(userConfig.Chains.Eth1), "Eth 1.0", false); err != nil {
        return err
    }
    if err := configureChain(&(globalConfig.Chains.Eth2), &(userConfig.Chains.Eth2), "Eth 2.0", true); err != nil {
        return err
    }

    // Save user config
    if err := rp.SaveUserConfig(userConfig); err != nil {
        return err
    }

    // Log & return
    fmt.Println("Done! Run 'rocketpool service start' to apply new configuration settings.")
    return nil

}


// Configure a chain
func configureChain(globalChain, userChain *config.Chain, chainName string, defaultRandomClient bool) error {

    // Check client options
    if len(globalChain.Client.Options) == 0 {
        return fmt.Errorf("There are no available %s client options", chainName)
    }

    // Prompt for random client selection
    var randomClient bool
    if defaultRandomClient {
        randomClient = cliutils.Confirm(fmt.Sprintf("Would you like to run a random %s client (recommended)?", chainName))
    }

    // Select client
    var selected int
    if randomClient {
        rand.Seed(time.Now().UnixNano())
        selected = rand.Intn(len(globalChain.Client.Options))
    } else {
        clientOptions := make([]string, len(globalChain.Client.Options))
        for oi, option := range globalChain.Client.Options {
            clientOptions[oi] = option.Name
        }
        selected, _ = cliutils.Select(fmt.Sprintf("Which %s client would you like to run?", chainName), clientOptions)
    }

    // Set selected client
    globalChain.Client.Selected = globalChain.Client.Options[selected].ID
    userChain.Client.Selected = globalChain.Client.Options[selected].ID

    // Log
    fmt.Printf("%s %s client selected.\n", globalChain.GetSelectedClient().Name, chainName)
    fmt.Println("")

    // Prompt for params
    params := []config.UserParam{}
    for _, param := range globalChain.GetSelectedClient().Params {

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
        if !param.Required {
            optionalLabel = " (leave blank for none)"
        }

        // Prompt for value
        value := cliutils.Prompt(fmt.Sprintf("Please enter the %s%s", param.Name, optionalLabel), expectedFormat, fmt.Sprintf("Invalid %s", param.Name))

        // Add param
        params = append(params, config.UserParam{
            Env: param.Env,
            Value: value,
        })

    }

    // Set unselected client params to blank strings to prevent docker-compose warnings
    for _, option := range globalChain.Client.Options {
        if option.ID == globalChain.Client.Selected { continue }
        for _, param := range option.Params {

            // Cancel if param already set in selected client
            paramSet := false
            for _, userParam := range params {
                if param.Env == userParam.Env {
                    paramSet = true
                    break
                }
            }
            if paramSet { continue }

            // Add param
            params = append(params, config.UserParam{
                Env: param.Env,
                Value: "",
            })

        }
    }

    // Set config params
    userChain.Client.Params = params

    // Return
    return nil

}

