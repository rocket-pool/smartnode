package pdao

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func setSignallingAddress(c *cli.Context, signallingAddress common.Address, signature string) error {
	// // Get RP client
	// rp, err := client.NewClientFromCtx(c)
	// if err != nil {
	// 	return err
	// }

	// // Build the TX
	// response, err := rp.Api.PDao.SetSignallingAddress()

	return nil

}

func clearSignallingAddress(c *cli.Context) error {

	// // Get RP client
	// rp, err := client.NewClientFromCtx(c)
	// if err != nil {
	// 	return err
	// }

	// // Build the TX
	// response, err := rp.Api.PDao.ClearSignallingAddress()

	return nil
}
