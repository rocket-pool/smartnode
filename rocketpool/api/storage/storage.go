package storage

import (
    "encoding/hex"
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


// Get a value from RocketStorage
func getStorage(c *cli.Context, dataType string, key [32]byte) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Handle data types
    switch dataType {

        // Get address
        case "address":
            if value, err := p.CM.RocketStorage.GetAddress(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, value.Hex(), "")
            }

        // Get bool
        case "bool":
            if value, err := p.CM.RocketStorage.GetBool(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, value, "")
            }

        // Get bytes
        case "bytes":
            if value, err := p.CM.RocketStorage.GetBytes(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, hexutil.AddPrefix(hex.EncodeToString(value)), "")
            }

        // Get bytes32
        case "bytes32":
            if value, err := p.CM.RocketStorage.GetBytes32(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, hexutil.AddPrefix(hex.EncodeToString(value[:])), "")
            }

        // Get int
        case "int":
            if value, err := p.CM.RocketStorage.GetInt(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, value, "")
            }

        // Get string
        case "string":
            if value, err := p.CM.RocketStorage.GetString(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, value, "")
            }

        // Get uint
        case "uint":
            if value, err := p.CM.RocketStorage.GetUint(nil, key); err != nil {
                return errors.New("Error retrieving data from storage: " + err.Error())
            } else {
                api.PrintResponse(p.Output, value, "")
            }

    }

    // Return
    return nil

}

