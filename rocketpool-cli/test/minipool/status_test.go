package minipool

import (
	"math/big"
	"strings"
	"testing"

	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/test"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type minipoolTest struct {
	*test.CLITest
}

func newMinipoolTest(t *testing.T) *minipoolTest {
	out := &minipoolTest{
		test.NewCLITest(t),
	}

	return out
}

func TestMinipoolStatusNoMinipools(t *testing.T) {
	cliTest := newMinipoolTest(t)
	result := cliTest.RepliesWith(
		types.ApiResponse[api.MinipoolStatusData]{
			Data: &api.MinipoolStatusData{},
		},
	).Run("minipool", "status")
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	// Cli output should mention no minipools available
	expectedResponse := "The node does not have any minipools yet."
	if !strings.Contains(result.Stdout, expectedResponse) {
		t.Fatalf(`expected message "%s", got "%s"`, expectedResponse, result.Stdout)
	}
}

func newMinipoolDetails() *api.MinipoolDetails {
	out := new(api.MinipoolDetails)

	// Avoid segfaults by initializing.
	// TODO: we shouldn't use big.Int pointers in api request/response types.
	out.Node.DepositBalance = big.NewInt(0)
	out.Node.RefundBalance = big.NewInt(0)
	out.User.DepositBalance = big.NewInt(0)

	out.Balances.Eth = big.NewInt(0)
	out.Balances.Reth = big.NewInt(0)
	out.Balances.Rpl = big.NewInt(0)
	out.Balances.FixedSupplyRpl = big.NewInt(0)

	out.NodeShareOfEthBalance = big.NewInt(0)

	out.Validator.Balance = big.NewInt(0)
	out.Validator.NodeBalance = big.NewInt(0)

	return out
}

func TestMinipoolStatusOnlyFinalized(t *testing.T) {
	cliTest := newMinipoolTest(t)
	response := types.ApiResponse[api.MinipoolStatusData]{
		Data: &api.MinipoolStatusData{
			Minipools: []api.MinipoolDetails{},
		},
	}
	minipool1 := newMinipoolDetails()
	minipool1.Finalised = true
	response.Data.Minipools = append(response.Data.Minipools, *minipool1)
	minipool2 := newMinipoolDetails()
	minipool2.Finalised = true
	response.Data.Minipools = append(response.Data.Minipools, *minipool2)
	result := cliTest.RepliesWith(response).Run("minipool", "status")
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	expectedResponse := "All of this node's minipools have been finalized."
	if !strings.Contains(result.Stdout, expectedResponse) {
		t.Fatalf(`expected message "%s", got "%s"`, expectedResponse, result.Stdout)
	}
}
