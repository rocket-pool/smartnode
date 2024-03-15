package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	sharedutils "github.com/rocket-pool/smartnode/shared/utils"
)

const (
	signatureVersion int = 1
)

type PersonalSignature struct {
	Address   common.Address `json:"address"`
	Message   string         `json:"msg"`
	Signature string         `json:"sig"`
	Version   string         `json:"version"` // beaconcha.in expects a string
}

var (
	signMessageFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "message",
		Aliases: []string{"m"},
		Usage:   "The 'quoted message' to be signed",
	}
)

func signMessage(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get & check wallet status
	status, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	if !sharedutils.IsWalletReady(status.Data.WalletStatus) {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	// Get the message
	message := c.String(signMessageFlag.Name)
	for message == "" {
		message = utils.Prompt("Please enter the message you want to sign: (EIP-191 personal_sign)", "^.+$", "Please enter the message you want to sign: (EIP-191 personal_sign)")
	}

	// Build the TX
	response, err := rp.Api.Wallet.SignMessage([]byte(message))
	if err != nil {
		return err
	}

	// Print the signature
	formattedSignature := PersonalSignature{
		Address:   status.Data.AccountAddress,
		Message:   message,
		Signature: response.Data.SignedMessage,
		Version:   fmt.Sprint(signatureVersion),
	}
	bytes, err := json.MarshalIndent(formattedSignature, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf("Signed Message:\n\n%s\n", string(bytes))
	return nil
}
