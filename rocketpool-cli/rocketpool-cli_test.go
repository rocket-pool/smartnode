package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prysmaticlabs/prysm/v5/testing/require"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/settings"
	"github.com/urfave/cli/v2"
)

func TestGlobalFlagsDefaults(t *testing.T) {
	// Drop args except the binary name
	args := os.Args[0:1]
	// Create a temp directory for trace files
	tempPath := t.TempDir()
	httpTraceFile := filepath.Join(tempPath, "http-trace.txt")
	apiUrl := "http://localhost:1337"
	// Test a subcommand, otherwise before/after fp are not executed
	args = append(args,
		"--native-mode=true",
		"-f=300",
		"-i=30",
		"--nonce=101",
		"--debug",
		"--api-address="+apiUrl,
		"--http-trace-path="+httpTraceFile,
		"-s",
		"auction",
		"--help",
	)

	// Make a system path
	systemPath := filepath.Join(tempPath, "rocketpool")
	err := os.MkdirAll(systemPath, 0755)
	require.NoError(t, err)

	// Emulating a (partial) installation by deploying a networks folder
	// Other installation files can be copied here later if necessary
	networkSettingsPath := filepath.Join(systemPath, "networks")
	err = os.MkdirAll(networkSettingsPath, 0755)
	require.NoError(t, err)

	// Copy the settings files
	networksSource := filepath.Join("..", "install", "deploy", "networks")
	entries, err := os.ReadDir(networksSource)
	require.NoError(t, err)
	for _, file := range entries {
		fileName := file.Name()

		sourceFile, err := os.Open(filepath.Join(networksSource, fileName))
		require.NoError(t, err)
		defer sourceFile.Close()

		targetFile, err := os.Create(filepath.Join(networkSettingsPath, fileName))
		require.NoError(t, err)
		defer targetFile.Close()

		_, err = io.Copy(targetFile, sourceFile)
		require.NoError(t, err)
		err = targetFile.Sync()
		require.NoError(t, err)
	}

	// Set the install dir env var
	err = os.Setenv(settings.TestSystemDirEnvVar, systemPath)
	require.NoError(t, err)

	app := newCliApp()

	// Capture stdout
	stdout := new(bytes.Buffer)
	app.Writer = stdout
	t.Cleanup(func() {
		t.Log("test stdout:")
		t.Log(stdout.String())
	})

	errs := make(chan error) // any assertions we want to run will be passed up as errors

	// Capture the default 'before'  and 'after' fps
	before := app.Before
	after := app.After

	// And overwrite it for our purposes
	app.Before = func(ctx *cli.Context) error {
		// Run the captured fp
		errs <- before(ctx)

		// Now we can inspect ctx for a SmartNodeSettings
		snSettings := settings.GetSmartNodeSettings(ctx)
		if snSettings == nil {
			errs <- errors.New("expected SmartNodeSettings to be instantiated, got nil")
			// return now to avoid segfaulting
			return nil
		}
		t.Logf("%+v\n", snSettings)

		// The default config path should have expanded
		userDir, err := os.UserHomeDir()
		if err != nil {
			errs <- err
		}
		if snSettings.ConfigPath != userDir+"/.rocketpool" {
			errs <- fmt.Errorf("expected config path in /home/%s/.rocketpool, got %s", userDir, snSettings.ConfigPath)
		}

		if snSettings.HttpTraceFile.Name() != httpTraceFile {
			errs <- fmt.Errorf("expected traces to be saved to %s, got %s", httpTraceFile, snSettings.HttpTraceFile.Name())
		}

		if !snSettings.DebugEnabled {
			errs <- errors.New("expected --debug to enable debug")
		}

		if !snSettings.SecureSession {
			errs <- errors.New("expected -s to enable a secure session")
		}

		if !strings.HasPrefix(snSettings.ApiUrl.String(), apiUrl) {
			errs <- fmt.Errorf("expected ApiUrl to start with %s, got %s", apiUrl, snSettings.ApiUrl)
		}

		if snSettings.Nonce.Uint64() != 101 {
			errs <- fmt.Errorf("expected nonce to be %d, got %d", 101, snSettings.Nonce.Uint64())
		}

		if snSettings.MaxFee != 300.0 {
			errs <- fmt.Errorf("expected max fee to be %f, got %f", 300.0, snSettings.MaxFee)
		}

		if snSettings.MaxPriorityFee != 30.0 {
			errs <- fmt.Errorf("expected max priority fee to be %f, got %f", 30.0, snSettings.MaxPriorityFee)
		}

		return nil
	}

	app.After = func(ctx *cli.Context) error {
		// Just run the captured fp and close the error channel
		errs <- after(ctx)
		return nil
	}

	// Parse the args
	go func() {
		errs <- app.Run(args)
		close(errs)
	}()

	// Check for errors
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
