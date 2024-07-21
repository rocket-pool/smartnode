package minipool

import (
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
