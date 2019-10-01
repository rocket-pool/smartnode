package node

import (
    "context"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Send resources from the node to an address
func sendFromNode(c *cli.Context, toAddressStr string, sendAmount float64, unit string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketPoolToken"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get parameters
    toAddress := common.HexToAddress(toAddressStr)
    sendAmountWei := eth.EthToWei(sendAmount)

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

    // Handle unit types
    switch unit {
        case "ETH":

            // Check balance
            if etherBalanceWei, err := p.Client.BalanceAt(context.Background(), nodeAccount.Address, nil); err != nil {
                return errors.New("Error retrieving node account ETH balance: " + err.Error())
            } else if etherBalanceWei.Cmp(sendAmountWei) == -1 {
                fmt.Fprintln(p.Output, "Send amount exceeds node account ETH balance")
                return nil
            }

            // Send
            if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
                return err
            } else {
                fmt.Fprintln(p.Output, "Sending ETH to address...")
                if _, err := eth.SendEther(p.Client, txor, &toAddress, sendAmountWei); err != nil {
                    return errors.New("Error transferring ETH to address: " + err.Error())
                }
            }

        case "RETH": fallthrough
        case "RPL":

            // Get token properties
            var tokenName string
            var tokenContract string
            switch unit {
                case "RETH":
                    tokenName = "rETH"
                    tokenContract = "rocketETHToken"
                case "RPL":
                    tokenName = "RPL"
                    tokenContract = "rocketPoolToken"
            }

            // Check balance
            tokenBalanceWei := new(*big.Int)
            if err := p.CM.Contracts[tokenContract].Call(nil, tokenBalanceWei, "balanceOf", nodeAccount.Address); err != nil {
                return errors.New(fmt.Sprintf("Error retrieving node account %s balance: " + err.Error(), tokenName))
            } else if (*tokenBalanceWei).Cmp(sendAmountWei) == -1 {
                fmt.Fprintln(p.Output, fmt.Sprintf("Send amount exceeds node account %s balance", tokenName))
                return nil
            }

            // Send
            if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
                return err
            } else {
                fmt.Fprintln(p.Output, fmt.Sprintf("Sending %s to address...", tokenName))
                if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses[tokenContract], p.CM.Abis[tokenContract], "transfer", toAddress, sendAmountWei); err != nil {
                    return errors.New(fmt.Sprintf("Error transferring %s to address: " + err.Error(), tokenName))
                }
            }

    }

    // Log & return
    fmt.Fprintln(p.Output, fmt.Sprintf("Successfully sent %.2f %s from node account to %s", sendAmount, unit, toAddress.String()))
    return nil

}

