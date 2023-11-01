package client

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ===============
// === Command ===
// ===============

// A command to be executed either locally or remotely
type command struct {
	cmd *exec.Cmd
}

// Create a command to be run by the Rocket Pool client
func (c *Client) newCommand(cmdText string) *command {
	return &command{
		cmd: exec.Command("sh", "-c", cmdText),
	}
}

// Run the command
func (c *command) Run() error {
	return c.cmd.Run()
}

// Start executes the command. Don't forget to call Wait
func (c *command) Start() error {
	return c.cmd.Start()
}

// Wait for the command to exit
func (c *command) Wait() error {
	return c.cmd.Wait()
}

func (c *command) SetStdin(r io.Reader) {
	c.cmd.Stdin = r
}

func (c *command) SetStdout(w io.Writer) {
	c.cmd.Stdout = w
}

func (c *command) SetStderr(w io.Writer) {
	c.cmd.Stderr = w
}

// Run the command and return its output
func (c *command) Output() ([]byte, error) {
	return c.cmd.Output()
}

// Get a pipe to the command's stdout
func (c *command) StdoutPipe() (io.Reader, error) {
	return c.cmd.StdoutPipe()
}

// Get a pipe to the command's stderr
func (c *command) StderrPipe() (io.Reader, error) {
	return c.cmd.StderrPipe()
}

// OutputPipes pipes for stdout and stderr
func (c *command) OutputPipes() (io.Reader, io.Reader, error) {
	cmdOut, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	cmdErr, err := c.StderrPipe()

	return cmdOut, cmdErr, err
}

// ==============
// === Client ===
// ==============

// Get the command used to escalate privileges on the system
func (c *Client) getEscalationCommand() (string, error) {
	// Check for sudo first
	sudo := "sudo"
	exists, err := c.checkIfCommandExists(sudo)
	if err != nil {
		return "", fmt.Errorf("error checking if %s exists: %w", sudo, err)
	}
	if exists {
		return sudo, nil
	}

	// Check for doas next
	doas := "doas"
	exists, err = c.checkIfCommandExists(doas)
	if err != nil {
		return "", fmt.Errorf("error checking if %s exists: %w", doas, err)
	}
	if exists {
		return doas, nil
	}

	return "", fmt.Errorf("no privilege escalation command found")
}

func (c *Client) checkIfCommandExists(command string) (bool, error) {
	// Run `type` to check for existence
	cmd := fmt.Sprintf("type %s", command)
	output, err := c.readOutput(cmd)

	if err != nil {
		exitErr, isExitErr := err.(*exec.ExitError)
		if isExitErr && exitErr.ProcessState.ExitCode() == 127 {
			// Command not found
			return false, nil
		} else {
			return false, fmt.Errorf("error checking if %s exists: %w", command, err)
		}
	} else {
		if strings.Contains(string(output), fmt.Sprintf("%s is", command)) {
			return true, nil
		} else {
			return false, fmt.Errorf("unexpected output when checking for %s: %s", command, string(output))
		}
	}
}

// Run a command and print its output
func (c *Client) printOutput(cmdText string) error {
	// Initialize command
	cmd := c.newCommand(cmdText)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for the command to exit
	return cmd.Wait()
}

// Run a command and return its output
func (c *Client) readOutput(cmdText string) ([]byte, error) {
	// Initialize command
	cmd := c.newCommand(cmdText)

	// Run command and return output
	return cmd.Output()
}
