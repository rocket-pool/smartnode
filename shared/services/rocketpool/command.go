package rocketpool

import (
	"io"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

// A command to be executed either locally or remotely
type command struct {
	cmd     *exec.Cmd
	session *ssh.Session
	cmdText string
}

// Create a command to be run by the Rocket Pool client
func (c *Client) newCommand(cmdText string) (*command, error) {
	if c.client == nil {
		return &command{
			cmd:     exec.Command("sh", "-c", cmdText),
			cmdText: cmdText,
		}, nil
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	return &command{
		session: session,
		cmdText: cmdText,
	}, nil
}

// Close the command session
func (c *command) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}

// Run the command
func (c *command) Run() error {
	if c.cmd != nil {
		return c.cmd.Run()
	}

	return c.session.Run(c.cmdText)
}

// Start executes the command. Don't forget to call Wait
func (c *command) Start() error {
	if c.cmd != nil {
		return c.cmd.Start()
	}

	return c.session.Start(c.cmdText)
}

// Wait for the command to exit
func (c *command) Wait() error {
	if c.cmd != nil {
		return c.cmd.Wait()
	}

	return c.session.Wait()
}

// Run the command and return its output
func (c *command) Output() ([]byte, error) {
	if c.cmd != nil {
		return c.cmd.Output()
	}

	return c.session.Output(c.cmdText)
}

// Get a pipe to the command's stdout
func (c *command) StdoutPipe() (io.Reader, error) {
	if c.cmd != nil {
		return c.cmd.StdoutPipe()
	}
	return c.session.StdoutPipe()
}

// Get a pipe to the command's stderr
func (c *command) StderrPipe() (io.Reader, error) {
	if c.cmd != nil {
		return c.cmd.StderrPipe()
	}

	return c.session.StderrPipe()
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
