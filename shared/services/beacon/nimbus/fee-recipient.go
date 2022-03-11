package nimbus

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Creates a fee recipient file that points all of this node's validators to the node distributor address.
func (c *Client) GenerateFeeRecipientFile(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]byte, error) {

	return nil, fmt.Errorf("Nimbus currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

}
