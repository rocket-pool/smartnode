package minipools

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/docker/docker/api/types"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
    "github.com/rocket-pool/smartnode/shared/utils/validator"
)


// Docker config
const VALIDATOR_CONTAINER_NAME string = "rocketpool_validator_1"
const VALIDATOR_RESTART_TIMEOUT string = "5s"
var validatorRestartTimeout, _ = time.ParseDuration(VALIDATOR_RESTART_TIMEOUT)


/**
 * Stake pre-launch minipools
 */
func (p *MinipoolsProcess) stakePrelaunchMinipools(minipoolAddresses []*common.Address) {

    // Check address count
    if len(minipoolAddresses) == 0 { return }

    // Log
    p.p.Log.Println("Staking prelaunch minipools...")

    // Get Rocket Pool withdrawal credentials
    withdrawalCredentialsBytes32 := new([32]byte)
    if err := p.p.CM.Contracts["rocketNodeAPI"].Call(nil, withdrawalCredentialsBytes32, "getWithdrawalCredentials"); err != nil {
        p.p.Log.Println(errors.New("Error retrieving Rocket Pool withdrawal credentials: " + err.Error()))
        return
    }
    withdrawalCredentials := (*withdrawalCredentialsBytes32)[:]

    // Check withdrawal credentials
    if bytes.Equal(withdrawalCredentials, make([]byte, 32)) {
        p.p.Log.Println(errors.New("Rocket Pool withdrawal credentials have not been initialized"))
        return
    }

    // Get eth2 config
    eth2Config, err := p.p.Beacon.GetEth2Config()
    if err != nil {
        p.p.Log.Println(err)
        return
    }

    // Stake minipools
    done := make(chan struct{})
    for _, minipoolAddress := range minipoolAddresses {
        go p.stakePrelaunchMinipool(minipoolAddress, withdrawalCredentials, eth2Config, done)
    }
    for _,_ = range minipoolAddresses { <-done }

    // Restart validator container
    p.restartValidator()

}


/**
 * Stake a pre-launch minipool
 */
func (p *MinipoolsProcess) stakePrelaunchMinipool(minipoolAddress *common.Address, withdrawalCredentials []byte, eth2Config *beacon.Eth2ConfigResponse, done chan struct{}) {

    // Send done signal on return
    defer (func() {
        done <- struct{}{}
    })()

    // Log
    p.p.Log.Println(fmt.Sprintf("Staking prelaunch minipool %s...", minipoolAddress.Hex()))

    // Generate new validator key
    validatorKey, err := p.p.KM.CreateValidatorKey()
    if err != nil {
        p.p.Log.Println(err)
        return
    }
    validatorPubkey := validatorKey.PublicKey.Marshal()

    // Get validator deposit data
    depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config)
    if err != nil {
        p.p.Log.Println(errors.New("Error building validator deposit data: " + err.Error()))
        return
    }

    // Stake minipool
    p.txLock.Lock()
    defer p.txLock.Unlock()
    if txor, err := p.p.AM.GetNodeAccountTransactor(); err != nil {
         p.p.Log.Println(err)
    } else {
        if _, err := eth.ExecuteContractTransaction(p.p.Client, txor, p.p.NodeContractAddress, p.p.CM.Abis["rocketNodeContract"], "stakeMinipool", minipoolAddress, validatorPubkey, depositData.Signature[:], depositDataRoot); err != nil {
            p.p.Log.Println(errors.New(fmt.Sprintf("Error staking minipool %s: " + err.Error(), minipoolAddress.Hex())))
        } else {
            p.p.Log.Println(fmt.Sprintf("Successfully staked minipool %s...", minipoolAddress.Hex()))
        }
    }

}


/**
 * Restart validator container
 */
func (p *MinipoolsProcess) restartValidator() {

    // Log
    p.p.Log.Println("Restarting validator container...")

    // Get all containers
    containers, err := p.p.Docker.ContainerList(context.Background(), types.ContainerListOptions{All: true})
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving docker containers: " + err.Error()))
        return
    }

    // Get validator container ID
    var validatorContainerId string
    for _, container := range containers {
        if container.Names[0] == "/" + VALIDATOR_CONTAINER_NAME {
            validatorContainerId = container.ID
            break
        }
    }
    if validatorContainerId == "" {
        p.p.Log.Println(errors.New("Validator container not found"))
        return
    }

    // Restart validator container
    if err := p.p.Docker.ContainerRestart(context.Background(), validatorContainerId, &validatorRestartTimeout); err != nil {
        p.p.Log.Println(errors.New("Error restarting validator container: " + err.Error()))
    } else {
        p.p.Log.Println("Successfully restarted validator container...")
    }

}

