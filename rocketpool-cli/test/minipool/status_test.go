package minipool

import (
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/test"
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

func TestMinipoolStatus(t *testing.T) {
	cliTest := newMinipoolTest(t)
	result := cliTest.Run("minipool", "status")
	if result.Error == nil {
		t.Fatal("running 'minipool status' without a mock response should produce an error")
	}

	errorstr := result.Error.Error()
	if !strings.Contains(errorstr, "connect: connection refused") {
		t.Fatalf("rocketpool-cli should return dial errors when unable to contact the api, got %s", errorstr)
	}
}
