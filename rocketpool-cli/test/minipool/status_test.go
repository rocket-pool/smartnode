package minipool

import (
	"io"
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
	if !strings.Contains(errorstr, "no response defined, call CLITest.RepliesWith") {
		t.Fatalf("rocketpool-cli should return expected errors when no mock response was provided, got %s", errorstr)
	}

	httpTrace, err := io.ReadAll(result.HTTPTraceFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(httpTrace), "GotFirstResponseByte") {
		t.Fatalf("test should have successfully traced http request to mock, instead got: %s", string(httpTrace))
	}
}
