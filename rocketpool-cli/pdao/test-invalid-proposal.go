package pdao

import (
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

var testInvalidProposalFlag bool

func applyTestInvalidProposal(rp *rocketpool.Client) {
	if !testInvalidProposalFlag {
		return
	}
	rp.SetTestInvalidProposal(true)
	color.YellowPrintln("WARNING: --test-invalid-proposal is set.")
	color.YellowPrintln("The submitted proposal will use a voting tree with one corrupted leaf,")
	color.YellowPrintln("so the on-chain proposal root will be invalid and can be challenged.")
	color.YellowPrintln("Expect the proposal to be defeated and the RPL proposal bond to be lost.")
	color.YellowPrintln("This flag is intended for testing only and cannot be used on mainnet.")
}
