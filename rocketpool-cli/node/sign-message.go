package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

const signatureVersion = 1

type PersonalSignature struct {
	Address   common.Address `json:"address"`
	Message   string         `json:"msg"`
	Signature string         `json:"sig"`
	Version   string         `json:"version"` // beaconcha.in expects a string
}

func signMessage(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get & check wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}

	if !status.WalletInitialized {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	message := c.String("message")
	for message == "" {
		message = prompt.Prompt("Please enter the message you want to sign: (EIP-191 personal_sign)", "^.+$", "Please enter the message you want to sign: (EIP-191 personal_sign)")
	}

	response, err := rp.SignMessage(message)
	if err != nil {
		return err
	}

	// Print the signature
	formattedSignature := PersonalSignature{
		Address:   status.AccountAddress,
		Message:   message,
		Signature: response.SignedData,
		Version:   fmt.Sprint(signatureVersion),
	}
	bytes, err := json.MarshalIndent(formattedSignature, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf("Signed Message:\n\n%s\n", string(bytes))

	return nil

}
