package minipool

import (
    "errors"
    "fmt"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Withdraw node deposit from a minipool
func withdrawMinipool(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings", "utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool", "rocketNodeContract"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil {
        return err
    }

    // Check withdrawals are allowed
    withdrawalsAllowed := new(bool)
    if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, withdrawalsAllowed, "getWithdrawalAllowed"); err != nil {
        return errors.New("Error checking node withdrawals enabled status: " + err.Error())
    } else if !*withdrawalsAllowed {
        fmt.Fprintln(p.Output, "Node withdrawals are currently disabled in Rocket Pool")
        return nil
    }

    // Get minipool addresses
    nodeAccount, _ := p.AM.GetNodeAccount()
    minipoolAddresses, err := node.GetMinipoolAddresses(nodeAccount.Address, p.CM)
    if err != nil {
        return err
    }
    minipoolCount := len(minipoolAddresses)

    // Get minipool node statuses
    nodeStatusChannel := make([]chan *minipool.NodeStatus, minipoolCount)
    nodeStatusErrorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        nodeStatusChannel[mi] = make(chan *minipool.NodeStatus)
        go (func(mi int) {
            if nodeStatus, err := minipool.GetNodeStatus(p.CM, minipoolAddresses[mi]); err != nil {
                nodeStatusErrorChannel <- err
            } else {
                nodeStatusChannel[mi] <- nodeStatus
            }
        })(mi)
    }

    // Receive minipool node statuses & filter withdrawable minipools
    withdrawableMinipoolAddresses := []*common.Address{}
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case nodeStatus := <-nodeStatusChannel[mi]:
                if nodeStatus.Status == minipool.WITHDRAWN && nodeStatus.DepositExists {
                    withdrawableMinipoolAddresses = append(withdrawableMinipoolAddresses, minipoolAddresses[mi])
                }
            case err := <-nodeStatusErrorChannel:
                return err
        }
    }

    // Cancel if no minipools are withdrawable
    if len(withdrawableMinipoolAddresses) == 0 {
        fmt.Fprintln(p.Output, "No minipools are currently available for withdrawal")
        return nil
    }

    // Prompt for minipools to withdraw
    prompt := []string{"Please select a minipool to withdraw from by entering a number, or enter 'A' for all:"}
    options := []string{}
    for mi, minipoolAddress := range withdrawableMinipoolAddresses {
        prompt = append(prompt, fmt.Sprintf("%d: %s", mi + 1, minipoolAddress.Hex()))
        options = append(options, strconv.Itoa(mi + 1))
    }
    response := cliutils.Prompt(p.Input, p.Output, strings.Join(prompt, "\n"), fmt.Sprintf("(?i)^(%s|a|all)$", strings.Join(options, "|")), "Please enter a minipool number or 'A' for all")

    // Get addresses of minipools to withdraw
    var withdrawMinipoolAddresses []*common.Address
    if strings.ToLower(response[:1]) == "a" {
        withdrawMinipoolAddresses = withdrawableMinipoolAddresses
    } else {
        index, _ := strconv.Atoi(response)
        withdrawMinipoolAddresses = []*common.Address{withdrawableMinipoolAddresses[index - 1]}
    }
    withdrawMinipoolCount := len(withdrawMinipoolAddresses)

    // Status channels
    withdrawSuccessChannel := make(chan bool)
    withdrawErrorChannel := make(chan error)

    // Withdraw node deposits
    for mi := 0; mi < withdrawMinipoolCount; mi++ {
        go (func(mi int) {
            if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
                withdrawErrorChannel <- errors.New(fmt.Sprintf("Error creating transactor for minipool %s: " + err.Error(), withdrawMinipoolAddresses[mi].Hex()))
            } else {
                fmt.Fprintln(p.Output, fmt.Sprintf("Withdrawing deposit from minipool %s...", withdrawMinipoolAddresses[mi].Hex()))
                if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "withdrawMinipoolDeposit", withdrawMinipoolAddresses[mi]); err != nil {
                    withdrawErrorChannel <- errors.New(fmt.Sprintf("Error withdrawing deposit from minipool %s: " + err.Error(), withdrawMinipoolAddresses[mi].Hex()))
                } else {
                    fmt.Fprintln(p.Output, "Successfully withdrew deposit from minipool", withdrawMinipoolAddresses[mi].Hex())
                    withdrawSuccessChannel <- true
                }
            }
        })(mi)
    }

    // Receive status & errors
    withdrawErrors := []string{"Error withdrawing deposits from one or more minipools:"}
    for received := 0; received < withdrawMinipoolCount; {
        select {
            case <-withdrawSuccessChannel:
                received++
            case err := <-withdrawErrorChannel:
                withdrawErrors = append(withdrawErrors, err.Error())
                received++
        }
    }

    // Return
    if len(withdrawErrors) > 1 { return errors.New(strings.Join(withdrawErrors, "\n")) }
    return nil

}

