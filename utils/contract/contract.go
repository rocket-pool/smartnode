package contract

import (
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get a contract's address and ABI
func GetDetails(rp *rocketpool.RocketPool, contractName string) (*common.Address, *abi.ABI, error) {

    // Data
    var wg errgroup.Group
    var address *common.Address
    var abi *abi.ABI

    // Load data
    wg.Go(func() error {
        var err error
        address, err = rp.GetAddress(contractName)
        return err
    })
    wg.Go(func() error {
        var err error
        abi, err = rp.GetABI(contractName)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, nil, err
    }

    // Return
    return address, abi, nil

}

