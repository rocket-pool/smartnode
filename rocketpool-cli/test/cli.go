package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/settings"
	"github.com/urfave/cli/v2"
)

// CLITest provides a interface for testing rocketpool-cli commands against a mock daemon.
type CLITest struct {
	t *testing.T

	TestDir string
	App     *cli.App

	commands uint
	server   *httptest.Server

	response any
}

// CLITestResult contains artifacts of running a cli command
type CLITestResult struct {
	Error         error // Error returned by "run" if any
	HTTPTraceFile *os.File
}

func NewCLITest(t *testing.T) *CLITest {
	out := &CLITest{
		t: t,
	}

	out.TestDir = t.TempDir()
	t.Logf("CLI Test using %s as .rocketpool\n", out.TestDir)

	// Set up the mock daemon http server
	out.server = httptest.NewServer(out)
	t.Cleanup(func() {
		out.server.Close()
	})

	// Set up the urfav/cli.App
	out.App = cli.NewApp()
	out.App.Name = "rocketpool-test"
	// Add global flags
	out.App.Flags = settings.AppendSmartNodeSettingsFlags(out.App.Flags)

	// Register commands
	commands.RegisterCommands(out.App)

	out.App.Before = func(c *cli.Context) error {
		_, err := settings.NewSmartNodeSettings(c)
		if err != nil {
			t.Fatal(err)
			return err
		}

		return nil
	}
	out.App.After = func(c *cli.Context) error {
		snSettings := settings.GetSmartNodeSettings(c)
		if snSettings != nil && snSettings.HttpTraceFile != nil {
			snSettings.HttpTraceFile.Close()
		}
		return nil
	}

	return out
}

func (cliTest *CLITest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cliTest.response == nil {
		w.WriteHeader(http.StatusBadGateway)
		_, err := w.Write([]byte(`{"error": "no response defined, call CLITest.RepliesWith before CLITest.Run"}`))
		if err != nil {
			cliTest.t.Fatal(err)
		}
		return
	}
	body, err := json.Marshal(cliTest.response)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		if err != nil {
			cliTest.t.Fatal(err)
		}
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			cliTest.t.Fatal(err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		cliTest.t.Fatal(err)
	}
}

func (cliTest *CLITest) RepliesWith(response any) *CLITest {
	cliTest.response = response
	return cliTest
}

func (cliTest *CLITest) Run(cmd ...string) *CLITestResult {
	// Create a path for this command in the temp dir
	cmdPath := filepath.Join(cliTest.TestDir, fmt.Sprintf("%d:%s", cliTest.commands, strings.Join(cmd[:min(len(cmd), 3)], "_")))
	err := os.Mkdir(cmdPath, 0755)
	if err != nil {
		cliTest.t.Fatal(err)
		return nil
	}
	cliTest.t.Logf("file artifacts for command '%s' will be saved in %s", strings.Join(cmd, " "), cmdPath)

	// Make a trace file that can be opened
	httpTraceFilePath := filepath.Join(cmdPath, "http-trace.txt")
	httpTraceFile, err := os.Create(httpTraceFilePath)
	if err != nil {
		cliTest.t.Fatal(err)
		return nil
	}

	// Prepend http trace path
	cmd = append([]string{"--http-trace-path", httpTraceFilePath}, cmd...)
	// Prepend config location
	cmd = append([]string{"--config-path", cliTest.TestDir}, cmd...)
	// Prepend mock http server
	cmd = append([]string{"--api-address", cliTest.server.URL}, cmd...)
	// Prepend expected test app name
	cmd = append([]string{"rocketpool-test"}, cmd...)
	cliTest.t.Logf("Running command %s\n", strings.Join(cmd, " "))

	defer func() {
		cliTest.response = nil
	}()

	return &CLITestResult{
		Error:         cliTest.App.Run(cmd),
		HTTPTraceFile: httpTraceFile,
	}
}
