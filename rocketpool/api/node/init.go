package node

import (
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Load Rocket Pool node contracts
func loadContracts(c *cli.Context) (*rocketpool.ContractManager, error) {

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("powHost"))
    if err != nil {
        return nil, err
    }

    // Load rocket Pool node contracts
    contractManager := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    contractManager.LoadContracts([]string{"rocketNodeAPI"})

    // Return contract manager
    return contractManager, nil

}

