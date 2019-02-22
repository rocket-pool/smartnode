package beacon

import (
    "context"
    "encoding/hex"
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "15s"


// Start beacon withdrawal process
func StartWithdrawalProcess(c *cli.Context, client *ethclient.Client, vm *node.ValidatorManager, fatalErrorChannel chan error) {

    // Check staking minipools for withdrawal on interval
    checkStakingMinipools(client, vm)
    checkMinipoolsInterval, _ := time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)
    checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
    for _ = range checkMinipoolsTimer.C {
        checkStakingMinipools(client, vm)
    }

}


// Check staking minipools for withdrawal
func checkStakingMinipools(client *ethclient.Client, vm *node.ValidatorManager) {

    // Get latest block header
    header, err := client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Println(errors.New("Error retrieving latest block header: " + err.Error()))
        return
    }

    // Log
    log.Println(fmt.Sprintf("Checking staking minipools for withdrawal at block %s...", header.Number.String()))

    // Check minipools
    for _, minipool := range vm.Validators {
        var exitBlock big.Int
        exitBlock.Add(minipool.StatusBlock, minipool.StakingDuration)
        if header.Number.Cmp(&exitBlock) > -1 {
            log.Println(fmt.Sprintf("Validator %s ready to withdraw, since block %s...", hex.EncodeToString(minipool.ValidatorPubkey), exitBlock.String()))
        } else {
            log.Println(fmt.Sprintf("Validator %s not ready to withdraw until block %s...", hex.EncodeToString(minipool.ValidatorPubkey), exitBlock.String()))
        }
    }

}

