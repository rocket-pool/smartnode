package minipool

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the node's minipool statuses
func getMinipoolStatus(c *cli.Context) error {

    // Command setup
    if message, err := setup(c, []string{"rocketPoolToken", "utilAddressSetStorage"}, []string{"rocketMinipool"}); message != "" {
        fmt.Println(message)
        return nil
    } else if err != nil {
        return err
    }

    // Get minipool addresses
    minipoolAddresses, err := node.GetMinipoolAddresses(am.GetNodeAccount().Address, cm)
    if err != nil {
        return err
    }
    minipoolCount := len(minipoolAddresses)

    // Get minipool details
    detailsChannels := make([]chan *minipool.Details, minipoolCount)
    errorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        detailsChannels[mi] = make(chan *minipool.Details)
        go (func(mi int) {
            if details, err := minipool.GetDetails(cm, minipoolAddresses[mi]); err != nil {
                errorChannel <- err
            } else {
                detailsChannels[mi] <- details
            }
        })(mi)
    }

    // Receive minipool details
    minipoolDetails := make([]*minipool.Details, minipoolCount)
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case details := <-detailsChannels[mi]:
                minipoolDetails[mi] = details
            case err := <-errorChannel:
                return err
        }
    }

    // Log status & return
    fmt.Println(fmt.Sprintf("Node has %d minipools", minipoolCount))
    for _, details := range minipoolDetails {
        fmt.Println("--------")
        fmt.Println("Address:               ", details.Address.Hex())
        fmt.Println("Status:                ", details.StatusType)
        fmt.Println("Status Updated Time:   ", details.StatusTime.Format("2006-01-02, 15:04 -0700 MST"))
        fmt.Println("Staking Duration:      ", details.StakingDurationId)
        fmt.Println("Node ETH Deposited:    ", fmt.Sprintf("%.2f", eth.WeiToEth(details.NodeEtherBalanceWei)))
        fmt.Println("Node RPL Deposited:    ", fmt.Sprintf("%.2f", eth.WeiToEth(details.NodeRplBalanceWei)))
        fmt.Println("User Count:            ", details.UserCount.String())
        fmt.Println("User Deposit Capacity: ", fmt.Sprintf("%.2f", eth.WeiToEth(details.UserDepositCapacityWei)))
        fmt.Println("User Deposit Total:    ", fmt.Sprintf("%.2f", eth.WeiToEth(details.UserDepositTotalWei)))
    }
    return nil

}

