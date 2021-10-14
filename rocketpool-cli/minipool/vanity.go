package minipool

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func findVanitySalt(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get the target prefix
    prefix := c.String("prefix")
    if prefix == "" {
        prefix = cliutils.Prompt("Please specify the address prefix you would like to search for (must start with 0x):", "^0x[0-9a-fA-F]+$", "Invalid hex string")
    }
    if !strings.HasPrefix(prefix, "0x") {
        return fmt.Errorf("Prefix must start with 0x.")
    }
    targetPrefix, success := big.NewInt(0).SetString(prefix, 16)
    if !success {
        return fmt.Errorf("Invalid prefix: %s")
    }

    // Get the starting salt
    saltString := c.String("salt")
    var salt *big.Int
    if saltString == "" {
        salt = big.NewInt(0)
    } else {
        salt, success = big.NewInt(0).SetString(saltString, 0)
        if !success {
            return fmt.Errorf("Invalid starting salt: %s")
        }
    }

    // Get deposit amount
    var amount float64
    if c.String("amount") != "" {

        // Parse amount
        depositAmount, err := strconv.ParseFloat(c.String("amount"), 64)
        if err != nil {
            return fmt.Errorf("Invalid deposit amount '%s': %w", c.String("amount"), err)
        }
        amount = depositAmount

    } else {

        // Get node status
        status, err := rp.NodeStatus()
        if err != nil {
            return err
        }

        // Get deposit amount options
        amountOptions := []string{
            "32 ETH (minipool begins staking immediately)",
            "16 ETH (minipool begins staking after ETH is assigned)",
        }
        if status.Trusted {
            amountOptions = append(amountOptions, "0 ETH  (minipool begins staking after ETH is assigned)")
        }

        // Prompt for amount
        selected, _ := cliutils.Select("Please choose a deposit type to search for:", amountOptions)
        switch selected {
            case 0: amount = 32
            case 1: amount = 16
            case 2: amount = 0
        }

    }
    amountWei := eth.EthToWei(amount)

    // Get the vanity generation artifacts
    vanityArtifacts, err := rp.GetVanityArtifacts(amountWei)
    if err != nil {
        return err
    }

    // Set up some variables
    zero := big.NewInt(0)
    one := big.NewInt(1)
    saltBytes := [32]byte{}
    hashInt := big.NewInt(0)
    nodeAddress := vanityArtifacts.NodeAddress.Bytes()
    minipoolManagerAddress := vanityArtifacts.MinipoolManagerAddress
    initHash := vanityArtifacts.InitHash.Bytes()
    shiftAmount := uint(42 - len(prefix))

    // Run the search
    start := time.Now()
    for {
        salt.FillBytes(saltBytes[:])
        nodeSalt := crypto.Keccak256Hash(nodeAddress, saltBytes[:])
    
        address := crypto.CreateAddress2(minipoolManagerAddress, nodeSalt, initHash)
        hashInt.SetBytes(address.Bytes())
        hashInt.Rsh(hashInt, shiftAmount * 4)
        hashInt.Xor(hashInt, targetPrefix)
        if hashInt.Cmp(zero) == 0 {
            fmt.Printf("Salt 0x%x = %s\n", salt, address.Hex())
            break
        } else {
            salt.Add(salt, one)
        }
    }
    end := time.Now()
    elapsed := end.Sub(start)
    fmt.Printf("Ran %s iterations in %s\n", salt.String(), elapsed)

    // Return
    return nil

}

