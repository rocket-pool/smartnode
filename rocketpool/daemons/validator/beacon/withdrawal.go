package beacon

import (
    "context"
    "encoding/hex"
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "15s"
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Start beacon withdrawal process
func StartWithdrawalProcess(p *services.Provider) {

    // Check staking minipools for withdrawal on interval
    go (func() {
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        for _ = range checkMinipoolsTimer.C {
            checkStakingMinipools(p)
        }
    })()

}


// Check staking minipools for withdrawal
func checkStakingMinipools(p *services.Provider) {

    // Get latest block header
    header, err := p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Println(errors.New("Error retrieving latest block header: " + err.Error()))
        return
    }

    // Log
    log.Println(fmt.Sprintf("Checking staking minipools for withdrawal at block %s...", header.Number.String()))

    // Check minipools
    for _, minipool := range p.VM.Validators {
        var exitBlock big.Int
        exitBlock.Add(minipool.StatusBlock, minipool.StakingDuration)
        if header.Number.Cmp(&exitBlock) > -1 {
            log.Println(fmt.Sprintf("Validator %s ready to withdraw, since block %s...", hex.EncodeToString(minipool.ValidatorPubkey), exitBlock.String()))
        } else {
            log.Println(fmt.Sprintf("Validator %s not ready to withdraw until block %s...", hex.EncodeToString(minipool.ValidatorPubkey), exitBlock.String()))
        }
    }

}

