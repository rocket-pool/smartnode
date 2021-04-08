package service

import (
	"fmt"
	"math/rand"
	"strconv"
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

    // Load configs
    globalConfig, err := rp.LoadGlobalConfig()
    if err != nil {
        return err
    }
    userConfig, err := rp.LoadUserConfig()
    if err != nil {
        return err
    }

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
            optionText := option.Name
            if option.Desc != "" {
                optionText += fmt.Sprintf(" %s\n\t\t%s\n", option.Desc, option.Link)
            }
            clientOptions[oi] = optionText
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

        // Get param label
        paramText := param.Name
        if !param.Required {
            blankText := "none"
            if param.BlankText != "" {
                blankText = param.BlankText
            }
            paramText += fmt.Sprintf(" (leave blank for %s)", blankText)
        }
        if param.Desc != "" {
            paramText += fmt.Sprintf("\n(%s)", param.Desc)
        }

        // Prompt for value
        var value string
        for {
            value = cliutils.Prompt(fmt.Sprintf("Please enter the %s", paramText), expectedFormat, fmt.Sprintf("Invalid %s", param.Name))
            isValid := true

            // Allow blanks for optional params
            if !param.Required && value == "" {
                value = param.Default
                break
            }

            // Type checking
            switch param.Type {
            case "uint":
                if _, err := strconv.ParseUint(value, 0, 0); err != nil {
                    fmt.Printf("'%s' is not a valid value for %s, try again.\n", value, param.Name)
                    isValid = false
                }
    
            case "uint16":
                if _, err := strconv.ParseUint(value, 0, 16); err != nil {
                    fmt.Printf("'%s' is not a valid value for %s, try again.\n", value, param.Name)
                    isValid = false
                }
            }
            
            if isValid {
                break
            }
        }

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

