package deposit

import (
    "fmt"
    "math/big"
    "strings"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Make a deposit
func makeDeposit(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        RPLExchangeAddress: true,
        RPLExchange: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketETHToken", "rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings", "rocketPool", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get deposit status
    status, err := deposit.GetDepositStatus(p)
    if err != nil { return err }

    // Reserve deposit if reservation doesn't exist
    var statusFormat string
    if status.ReservationExists {
        statusFormat = "Node already has a deposit reservation requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s"
    } else {
        statusFormat = "Deposit reservation made successfully, requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s"

        // Check node deposit can be reserved
        canReserve, err := deposit.CanReserveDeposit(p, durationId)
        if err != nil { return err }

        // Check response
        if canReserve.DepositsDisabled {
            fmt.Fprintln(p.Output, "Node deposits are currently disabled in Rocket Pool")
        }
        if canReserve.StakingDurationDisabled {
            fmt.Fprintln(p.Output, fmt.Sprintf("The staking duration '%s' is invalid or disabled", durationId))
        }
        if !canReserve.Success {
            return nil
        }

        // Reserve deposit
        _, err = deposit.ReserveDeposit(p, durationId)
        if err != nil { return err }

        // Get deposit status
        status, err = deposit.GetDepositStatus(p)
        if err != nil { return err }

    }

    // Get node balance totals
    nodeTotalBalanceEtherWei := big.NewInt(0)
    nodeTotalBalanceEtherWei.Add(status.NodeContractBalanceEtherWei, status.NodeAccountBalanceEtherWei)
    nodeTotalBalanceRplWei := big.NewInt(0)
    nodeTotalBalanceRplWei.Add(status.NodeContractBalanceRplWei, status.NodeAccountBalanceRplWei)

    // Print deposit status
    fmt.Fprintln(p.Output, fmt.Sprintf(
        statusFormat,
        eth.WeiToEth(status.ReservationEtherRequiredWei),
        eth.WeiToEth(status.ReservationRplRequiredWei),
        status.ReservationStakingDurationID,
        status.ReservationExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    fmt.Fprintln(p.Output, fmt.Sprintf(
        "Node has a combined balance of %.2f ETH and %.2f RPL",
        eth.WeiToEth(nodeTotalBalanceEtherWei),
        eth.WeiToEth(nodeTotalBalanceRplWei)))

    // Prompt for action
    action := cliutils.Prompt(p.Input, p.Output, "Would you like to:\n1. Complete the deposit;\n2. Cancel the deposit; or\n3. Finish later?", "^(1|2|3)$", "Please answer '1', '2' or '3'")
    switch action {

        // Complete deposit
        case "1":

            // Check deposit can be completed
            canComplete, err := deposit.CanCompleteDeposit(p)
            if err != nil { return err }

            // Check response
            if canComplete.DepositsDisabled {
                fmt.Fprintln(p.Output, "Node deposits are currently disabled in Rocket Pool")
            }
            if canComplete.MinipoolCreationDisabled {
                fmt.Fprintln(p.Output, "Minipool creation is currently disabled in Rocket Pool")
            }
            if canComplete.InsufficientNodeEtherBalance {
                fmt.Fprintln(p.Output, fmt.Sprintf(
                    "Combined node balance of %.2f ETH is not enough to cover requirement of %.2f ETH" + "\r\n" +
                    "(Node account contains %.2f ETH and node contract contains %.2f ETH)",
                    eth.WeiToEth(nodeTotalBalanceEtherWei),
                    eth.WeiToEth(status.ReservationEtherRequiredWei),
                    eth.WeiToEth(status.NodeAccountBalanceEtherWei),
                    eth.WeiToEth(status.NodeContractBalanceEtherWei)))
            }
            if canComplete.DepositsDisabled || canComplete.MinipoolCreationDisabled || canComplete.InsufficientNodeEtherBalance {
                return nil
            }

            // RPL purchase required
            if canComplete.RplShortByWei.Cmp(big.NewInt(0)) > 0 {

                // Get exchange liquidity
                liquidity, err := exchange.GetTokenLiquidity(p, "RPL")
                if err != nil { return err }

                // Check exchange has sufficient liquidity
                if liquidity.ExchangeTokenBalanceWei.Cmp(canComplete.RplShortByWei) < 0 {
                    fmt.Fprintln(p.Output, fmt.Sprintf(
                        "Combined node balance of %.2f RPL plus Uniswap exchange liquidity of %.2f RPL is not enough to cover requirement of %.2f RPL" + "\r\n" +
                        "(Node account contains %.2f RPL and node contract contains %.2f RPL)",
                        eth.WeiToEth(nodeTotalBalanceRplWei),
                        eth.WeiToEth(liquidity.ExchangeTokenBalanceWei),
                        eth.WeiToEth(status.ReservationRplRequiredWei),
                        eth.WeiToEth(status.NodeAccountBalanceRplWei),
                        eth.WeiToEth(status.NodeContractBalanceRplWei)))
                    return nil
                }

                // Get exchange token price
                price, err := exchange.GetTokenPrice(p, canComplete.RplShortByWei, "RPL")
                if err != nil { return err }

                // Get total ether required including exchange price
                totalEtherRequiredWei := big.NewInt(0)
                totalEtherRequiredWei.Add(canComplete.EtherRequiredWei, price.MaxEtherPriceWei)

                // Get total deposit amount including exchange price
                totalDepositEtherAmountWei := big.NewInt(0)
                totalDepositEtherAmountWei.Add(status.ReservationEtherRequiredWei, price.MaxEtherPriceWei)

                // Check total ether required
                if status.NodeAccountBalanceEtherWei.Cmp(totalEtherRequiredWei) < 0 {
                    fmt.Fprintln(p.Output, fmt.Sprintf(
                        "Node account balance of %.2f ETH is not enough to cover remaining requirement of %.2f ETH" + "\r\n" +
                        "(%.2f ETH required by node contract and %.2f ETH required for Uniswap RPL purchase)",
                        eth.WeiToEth(status.NodeAccountBalanceEtherWei),
                        eth.WeiToEth(totalEtherRequiredWei),
                        eth.WeiToEth(canComplete.EtherRequiredWei),
                        eth.WeiToEth(price.MaxEtherPriceWei)))
                    return nil
                }

                // Confirm purchase of RPL
                rplPurchaseConfirmed := cliutils.Prompt(p.Input, p.Output,
                    fmt.Sprintf(
                        "Deposit requires a total of %.2f ETH, would you like to continue? [y/n]" + "\r\n" +
                        "(%.2f ETH already on node contract with %.2f ETH still required, and %.2f ETH required for Uniswap RPL purchase)",
                        eth.WeiToEth(totalDepositEtherAmountWei),
                        eth.WeiToEth(status.NodeContractBalanceEtherWei),
                        eth.WeiToEth(canComplete.EtherRequiredWei),
                        eth.WeiToEth(price.MaxEtherPriceWei)),
                    "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
                if strings.ToLower(rplPurchaseConfirmed[:1]) == "n" {
                    fmt.Fprintln(p.Output, "Deposit not completed")
                    return nil
                }

                // Purchase RPL
                _, err = exchange.BuyTokens(p, price.MaxEtherPriceWei, canComplete.RplShortByWei, "RPL")
                if err != nil { return err }

            }

            // Transfer remaining required RPL
            _, err = node.SendFromNode(p, *(p.NodeContractAddress), canComplete.RplRequiredWei, "RPL")
            if err != nil { return err }

            // Complete deposit
            completed, err := deposit.CompleteDeposit(p, canComplete.EtherRequiredWei, canComplete.DepositDurationId)
            if err != nil { return err }

            // Print output
            if completed.Success {
                fmt.Fprintln(p.Output, "Deposit completed successfully, minipool created at", completed.MinipoolAddress.Hex())
            }

        // Cancel deposit
        case "2":

            // Cancel deposit
            cancelled, err := deposit.CancelDeposit(p)
            if err != nil { return err }

            // Print output
            if cancelled.Success {
                fmt.Fprintln(p.Output, "Deposit reservation cancelled successfully")
            }

    }

    // Return
    return nil

}

