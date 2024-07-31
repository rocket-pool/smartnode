package ens

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

type AddressOrENSData struct {
	address       common.Address
	addressString string
}

// Constructor Function
func NewAddressOrENSData() *AddressOrENSData {
	return &AddressOrENSData{}
}

// Getter for Address
func (e *AddressOrENSData) GetToAddress() common.Address {
	return e.address
}

// Getter for AddressString
func (e *AddressOrENSData) GetToAddressString() string {
	return e.addressString
}

// ResolveAddressOrEns checks if an input contains an ENS address, resolves it to a hex
// address, then stores it as a common.Address and a string. If the input is a hex address,
// the address is validated then stored as a common.Address and a string
func (e *AddressOrENSData) ResolveAddressOrEns(c *cli.Context, AddressOrEns string) error {
	// Get the RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	// Get the Address
	if strings.Contains(AddressOrEns, ".") {
		response, err := rp.Api.Node.ResolveEns(common.Address{}, AddressOrEns)
		if err != nil {
			return err
		}
		e.address = response.Data.Address
		e.addressString = fmt.Sprintf("%s (%s)", AddressOrEns, e.address.Hex())
		return nil
	} else {
		var err error
		e.address, err = input.ValidateAddress("address", AddressOrEns)
		if err != nil {
			return err
		}
		e.addressString = e.address.Hex()
		return nil
	}
}
