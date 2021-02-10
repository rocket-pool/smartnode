package dao

import (
    "fmt"
    "strings"
    "sync"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get the string representation of a proposal payload
var getProposalPayloadStringLock sync.Mutex
func GetProposalPayloadString(rp *rocketpool.RocketPool, daoName string, payload []byte) (string, error) {

    // Lock while getting proposal payload string
    getProposalPayloadStringLock.Lock()
    defer getProposalPayloadStringLock.Unlock()

    // Get proposal DAO contract ABI
    abi, err := rp.GetABI(daoName)
    if err != nil {
        return "", fmt.Errorf("Could not get '%s' DAO contract ABI: %w", daoName, err)
    }

    // Get proposal payload method
    method, err := abi.MethodById(payload)
    if err != nil {
        return "", fmt.Errorf("Could not get proposal payload method: %w", err)
    }

    // Get proposal payload argument values
    args, err := method.Inputs.UnpackValues(payload)
    if err != nil {
        return "", fmt.Errorf("Could not get proposal payload arguments: %w", err)
    }

    // Format argument values as strings
    argStrs := []string{}
    for _, arg := range args {
        argStrs = append(argStrs, fmt.Sprintf("%v", arg))
    }

    // Build & return payload string
    return fmt.Sprintf("%s(%s)", method.Sig, strings.Join(argStrs, ",")), nil

}

