package rocketpool

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestGetAddress(t *testing.T) {

	// Get contract address
	address1, err := rp.GetAddress("rocketDepositPool")
	if err != nil {
		t.Fatalf("Could not get contract address: %s", err)
	} else if bytes.Equal(address1.Bytes(), common.Address{}.Bytes()) {
		t.Error("Contract address was not found")
	}

	// Get cached contract address
	address2, err := rp.GetAddress("rocketDepositPool")
	if err != nil {
		t.Fatalf("Could not get cached contract address: %s", err)
	} else if !bytes.Equal(address2.Bytes(), address1.Bytes()) {
		t.Error("Cached contract address did not match original contract address")
	}

}

func TestGetAddresses(t *testing.T) {

	// Get contract addresses
	addresses1, err := rp.GetAddresses("rocketNodeManager", "rocketNodeDeposit")
	if err != nil {
		t.Fatalf("Could not get contract addresses: %s", err)
	} else {
		for ai, address := range addresses1 {
			if bytes.Equal(address.Bytes(), common.Address{}.Bytes()) {
				t.Errorf("Contract address %d was not found", ai)
			}
		}
	}

	// Get cached contract addresses
	addresses2, err := rp.GetAddresses("rocketNodeManager", "rocketNodeDeposit")
	if err != nil {
		t.Fatalf("Could not get cached contract addresses: %s", err)
	} else {
		for ai := 0; ai < len(addresses2); ai++ {
			if !bytes.Equal(addresses2[ai].Bytes(), addresses1[ai].Bytes()) {
				t.Errorf("Cached contract address %d did not match original contract address", ai)
			}
		}
	}

}

func TestGetABI(t *testing.T) {

	// Get ABI
	abi1, err := rp.GetABI("rocketDepositPool")
	if err != nil {
		t.Fatalf("Could not get contract ABI: %s", err)
	}

	// Get cached ABI
	abi2, err := rp.GetABI("rocketDepositPool")
	if err != nil {
		t.Fatalf("Could not get cached contract ABI: %s", err)
	} else {
		abi2Json, err := json.Marshal(abi2)
		if err != nil {
			t.Fatal(err)
		}
		abi1Json, err := json.Marshal(abi1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(abi2Json, abi1Json) {
			t.Error("Cached contract ABI did not match original contract ABI")
		}
	}

}

func TestGetABIs(t *testing.T) {

	// Get ABIs
	abis1, err := rp.GetABIs("rocketNodeManager", "rocketNodeDeposit")
	if err != nil {
		t.Fatalf("Could not get contract ABIs: %s", err)
	}

	// Get cached ABIs
	abis2, err := rp.GetABIs("rocketNodeManager", "rocketNodeDeposit")
	if err != nil {
		t.Fatalf("Could not get cached contract ABIs: %s", err)
	} else {
		for ai := 0; ai < len(abis2); ai++ {
			abi2Json, err := json.Marshal(abis2[ai])
			if err != nil {
				t.Fatal(err)
			}
			abi1Json, err := json.Marshal(abis1[ai])
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(abi2Json, abi1Json) {
				t.Errorf("Cached contract ABI %d did not match original contract ABI", ai)
			}
		}
	}

}

func TestGetContract(t *testing.T) {

	// Get contract
	if _, err := rp.GetContract("rocketDepositPool"); err != nil {
		t.Fatalf("Could not get contract: %s", err)
	}

	// Get cached contract
	if _, err := rp.GetContract("rocketDepositPool"); err != nil {
		t.Fatalf("Could not get cached contract: %s", err)
	}

}

func TestGetContracts(t *testing.T) {

	// Get contracts
	if _, err := rp.GetContracts("rocketNodeManager", "rocketNodeDeposit"); err != nil {
		t.Fatalf("Could not get contracts: %s", err)
	}

	// Get cached contracts
	if _, err := rp.GetContracts("rocketNodeManager", "rocketNodeDeposit"); err != nil {
		t.Fatalf("Could not get cached contracts: %s", err)
	}

}

func TestMakeContract(t *testing.T) {

	// Make contract
	if _, err := rp.MakeContract("rocketMinipool", common.HexToAddress("0x1111111111111111111111111111111111111111")); err != nil {
		t.Fatalf("Could not make contract: %s", err)
	}

	// Make contract with cached ABI
	if _, err := rp.MakeContract("rocketMinipool", common.HexToAddress("0x2222222222222222222222222222222222222222")); err != nil {
		t.Fatalf("Could not make contract with cached ABI: %s", err)
	}

}
