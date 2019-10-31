package minipool

import (
    "errors"
    "fmt"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    minipoolapi "github.com/rocket-pool/smartnode/shared/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
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
        NodeContractAddress: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings", "utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool", "rocketMinipoolDelegateNode", "rocketNodeContract"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check node withdrawals are enabled
    canWithdraw, err := minipoolapi.CanWithdrawMinipools(p)
    if err != nil { return err }

    // Cancel if node withdrawals are disabled
    if !canWithdraw {
        fmt.Fprintln(p.Output, "Node withdrawals are currently disabled in Rocket Pool")
        return nil
    }

    // Get withdrawable minipools
    withdrawableMinipools, err := minipoolapi.GetWithdrawableMinipools(p)
    if err != nil { return err }

    // Cancel if no minipools are withdrawable
    if len(withdrawableMinipools) == 0 {
        fmt.Fprintln(p.Output, "No minipools are currently available for withdrawal")
        return nil
    }

    // Prompt for minipools to withdraw
    prompt := []string{"Please select a minipool to withdraw from by entering a number, or enter 'A' for all (excluding initialized):"}
    options := []string{}
    for mi, minipoolStatus := range withdrawableMinipools {
        prompt = append(prompt, fmt.Sprintf(
            "%-4v %-42v  %-11v @ %v  %-3v",
            strconv.Itoa(mi + 1) + ":",
            minipoolStatus.Address.Hex(),
            strings.Title(minipoolStatus.StatusType),
            minipoolStatus.StatusTime.Format("2006-01-02, 15:04 -0700 MST"),
            minipoolStatus.StakingDurationId))
        options = append(options, strconv.Itoa(mi + 1))
    }
    response := cliutils.Prompt(p.Input, p.Output, strings.Join(prompt, "\n"), fmt.Sprintf("(?i)^(%s|a|all)$", strings.Join(options, "|")), "Please enter a minipool number or 'A' for all (excluding initialized)")

    // Get addresses of minipools to withdraw
    withdrawMinipoolAddresses := []*common.Address{}
    if strings.ToLower(response[:1]) == "a" {
        for _, minipoolStatus := range withdrawableMinipools {
            if minipoolStatus.Status != minipool.INITIALIZED {
                withdrawMinipoolAddresses = append(withdrawMinipoolAddresses, minipoolStatus.Address)
            }
        }
    } else {
        index, _ := strconv.Atoi(response)
        withdrawMinipoolAddresses = append(withdrawMinipoolAddresses, withdrawableMinipools[index - 1].Address)
    }
    withdrawMinipoolCount := len(withdrawMinipoolAddresses)

    // Cancel if no minipools to withdraw
    if withdrawMinipoolCount == 0 {
        fmt.Fprintln(p.Output, "No minipools to withdraw")
        return nil
    }

    // Withdraw node deposits
    withdrawErrors := []string{"Error withdrawing deposits from one or more minipools:"}
    for mi := 0; mi < withdrawMinipoolCount; mi++ {
        minipoolAddress := withdrawMinipoolAddresses[mi]

        // Withdraw from minipool & print output
        if withdrawn, err := minipoolapi.WithdrawMinipool(p, *minipoolAddress); err != nil {
            withdrawErrors = append(withdrawErrors, fmt.Sprintf("Error withdrawing from minipool %s: %s", minipoolAddress.Hex(), err.Error()))
        } else {
            fmt.Fprintln(p.Output, fmt.Sprintf(
                "Successfully withdrew deposit of %.2f ETH, %.2f rETH and %.2f RPL from minipool %s",
                eth.WeiToEth(withdrawn.EtherWithdrawnWei),
                eth.WeiToEth(withdrawn.RethWithdrawnWei),
                eth.WeiToEth(withdrawn.RplWithdrawnWei),
                minipoolAddress.Hex()))
        }

    }

    // Return
    if len(withdrawErrors) > 1 {
        return errors.New(strings.Join(withdrawErrors, "\n"))
    }
    return nil

}

