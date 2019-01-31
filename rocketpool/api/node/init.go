package node

import (
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Initialise ethereum client & node contracts
func initClient(c *cli.Context) (*ethclient.Client, *rocketpool.ContractManager, error) {

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("powHost"))
    if err != nil {
        return nil, nil, err
    }

    // Initialise Rocket Pool contract manager
    contractManager, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return nil, nil, err
    }

    // Load Rocket Pool node contracts
    contractManager.LoadContracts([]string{"rocketNodeAPI"})

    // Return contract manager
    return client, contractManager, nil

}

