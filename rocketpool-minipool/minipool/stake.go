package minipool

import (
    "bytes"
    "context"
    "encoding/hex"
    "errors"
    "time"

    "github.com/docker/docker/api/types"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
    "github.com/rocket-pool/smartnode/shared/utils/validator"
)


// Docker config
const VALIDATOR_CONTAINER_NAME string = "rocketpool_validator_1"
const VALIDATOR_RESTART_TIMEOUT string = "5s"
var validatorRestartTimeout, _ = time.ParseDuration(VALIDATOR_RESTART_TIMEOUT)


// Stake minipool
func Stake(p *services.Provider, pool *Minipool) error {

    // Check minipool status
    if status, err := minipool.GetStatusCode(p.CM, pool.Address); err != nil {
        return errors.New("Error retrieving minipool status: " + err.Error())
    } else if status != minipool.PRELAUNCH {
        return nil
    }

    // Get Rocket Pool withdrawal credentials
    withdrawalCredentialsBytes32 := new([32]byte)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, withdrawalCredentialsBytes32, "getWithdrawalCredentials"); err != nil {
        return errors.New("Error retrieving Rocket Pool withdrawal credentials: " + err.Error())
    }
    withdrawalCredentials := (*withdrawalCredentialsBytes32)[:]

    // Check withdrawal credentials
    if bytes.Equal(withdrawalCredentials, make([]byte, 32)) {
        return errors.New("Rocket Pool withdrawal credentials have not been initialized")
    }

    // Generate new validator key
    validatorKey, err := p.KM.CreateValidatorKey()
    if err != nil { return err }
    validatorPubkey := validatorKey.PublicKey.Marshal()

    // Get validator deposit data
    depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials)
    if err != nil { return errors.New("Error building validator deposit data: " + err.Error()) }

    // Stake minipool
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "stakeMinipool", pool.Address, validatorPubkey, depositData.Signature[:], depositDataRoot); err != nil {
            return errors.New("Error staking minipool: " + err.Error())
        }
    }

    // Log
    p.Log.Println("Successfully staked minipool...")

    // Restart validator container
    if err := restartValidator(p); err != nil { return err }

    // Encode validator pubkey and add to minipool data
    validatorPubkeyHex := make([]byte, hex.EncodedLen(len(validatorPubkey)))
    hex.Encode(validatorPubkeyHex, validatorPubkey)
    validatorPubkeyStr := string(validatorPubkeyHex)
    pool.Pubkey = validatorPubkeyStr

    // Return
    return nil

}


// Restart validator container
func restartValidator(p *services.Provider) error {

    // Get all containers
    containers, err := p.Docker.ContainerList(context.Background(), types.ContainerListOptions{All: true})
    if err != nil { return errors.New("Error retrieving docker containers: " + err.Error()) }

    // Get validator container ID
    var validatorContainerId string
    for _, container := range containers {
        if container.Names[0] == "/" + VALIDATOR_CONTAINER_NAME {
            validatorContainerId = container.ID
            break
        }
    }
    if validatorContainerId == "" { return errors.New("Validator container not found") }

    // Restart validator container
    if err := p.Docker.ContainerRestart(context.Background(), validatorContainerId, &validatorRestartTimeout); err != nil {
        return errors.New("Error restarting validator container: " + err.Error())
    }

    // Log
    p.Log.Println("Successfully restarted validator container...")

    // Return
    return nil

}

