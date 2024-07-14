package pdao

import (
	"github.com/urfave/cli/v2"
)

func setSignallingAddress(c *cli.Context, signallingAddressString string, signature string) error {

	// // Input Validation
	// var err error
	// signallingAddress, err := input.ValidateAddress("signalling-address", signallingAddressString)
	// if err != nil {
	// 	return err
	// }
	// // Todo: input.ValidateSignature
	// sig, err := input.ValidateSignature("signature", signature)
	// if err != nil {
	// 	return err
	// }

	return nil

}

func clearSignallingAddress(c *cli.Context) error {

	return nil
}
