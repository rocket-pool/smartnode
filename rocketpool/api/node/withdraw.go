package node

import (
    "context"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Withdraw resources from the node
func withdrawFromNode(c *cli.Context, amount float64, unit string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketNodeContract"},
    })
    if err != nil {
        return err
    }

    // Convert withdrawal amount to wei
    amountWei := eth.EthToWei(amount)

    // Get contract method names
    var balanceMethod string
    var withdrawMethod string
    switch unit {
        case "ETH":
            balanceMethod = "getBalanceETH"
            withdrawMethod = "withdrawEther"
        case "RPL":
            balanceMethod = "getBalanceRPL"
            withdrawMethod = "withdrawRPL"
    }

    // Check withdrawal amount is available
    balanceWei := new(*big.Int)
    if err := p.NodeContract.Call(nil, balanceWei, balanceMethod); err != nil {
        return errors.New("Error retrieving node balance: " + err.Error())
    } else if amountWei.Cmp(*balanceWei) > 0 {
        fmt.Println("Withdrawal amount exceeds available balance on node contract")
        return nil
    }

    // Withdraw amount
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        txor.GasLimit = 100000 // Gas estimates on this method are incorrect
        if tx, err := p.NodeContract.Transact(txor, withdrawMethod, amountWei); err != nil {
            return errors.New("Error withdrawing from node contract: " + err.Error())
        } else {

            // Wait for transaction to be mined before continuing
            fmt.Println("Withdrawal transaction awaiting mining...")
            bind.WaitMined(context.Background(), p.Client, tx)

        }
    }

    // Log & return
    fmt.Println(fmt.Sprintf("Successfully withdrew %.2f %s from node contract to account", amount, unit))
    return nil

}

