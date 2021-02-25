package minipool

import (
    "fmt"
    "sort"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/hex"
)


func getLeader(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolLeader()
    if err != nil {
        return err
    }

    // Get minipools by status
    minipools := []api.MinipoolDetails{}
    statusCounts := map[string]int{}
    for _, minipool := range status.Minipools {

        // Add to status list
        if minipool.Validator.Exists {
            minipools = append(minipools, minipool)
        }

        // status count
        statusName := minipool.Status.Status.String()
        if _, ok := statusCounts[statusName]; !ok {
            statusCounts[statusName] = 0
        }
        statusCounts[statusName] = statusCounts[statusName]+1
    }

    fmt.Printf("Total minipools: %d\n", len(status.Minipools))
    fmt.Println("Status,Count")

    for status, count := range statusCounts {
        fmt.Printf("%s,%d", status, count)
        fmt.Println("")
    }

    // Print & return
    if len(status.Minipools) == 0 {
        fmt.Println("No active minipools")
        return nil
    }
    fmt.Println("")

    sort.SliceStable(minipools, func(i, j int) bool { return eth.WeiToEth(minipools[i].Validator.Balance) > eth.WeiToEth(minipools[j].Validator.Balance) })

    fmt.Printf("Minipools with validators: %d\n", len(minipools))
    fmt.Println("Rank,Node address,Validator pubkey,RP status update time,Accumulated reward/penalty (ETH)")

    for i, minipool := range minipools {
        nodeAddress := hex.AddPrefix(minipool.Node.Address.Hex())
        validatorAddress := hex.AddPrefix(minipool.ValidatorPubkey.Hex())
        statusTime := minipool.Status.StatusTime.Format("2006-01-02T15:04:05-0700")
        diffBalance := eth.WeiToEth(minipool.Validator.Balance) - eth.WeiToEth(minipool.Node.DepositBalance) - eth.WeiToEth(minipool.User.DepositBalance)
        fmt.Printf("%4d,%s,%s,%s,%+0.10f", i+1, nodeAddress, validatorAddress, statusTime, diffBalance)
        fmt.Println("")
    }
    return nil

}
