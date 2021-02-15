package network

import (
    "bytes"
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/network"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestSetSystemWithdrawalContractAddress(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set SWC address
    swcAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    if _, err := network.SetSystemWithdrawalContractAddress(rp, swcAddress, ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check SWC address
    if networkSwcAddress, err := network.GetSystemWithdrawalContractAddress(rp, nil); err != nil {
        t.Error(err)
    } else if !bytes.Equal(networkSwcAddress.Bytes(), swcAddress.Bytes()) {
        t.Errorf("Incorrect system withdrawal contract address %s", networkSwcAddress.Hex())
    }

}

