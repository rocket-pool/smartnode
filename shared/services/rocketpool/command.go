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
	} else {
		session, err := c.client.NewSession()
		if err != nil {
			return nil, err
		}
		return &command{
			session: session,
			cmdText: cmdText,
		}, nil
	}
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
	} else {
		return c.session.Run(c.cmdText)
	}
}

// Run the command and return its output
func (c *command) Output() ([]byte, error) {
	if c.cmd != nil {
		return c.cmd.Output()
	} else {
		return c.session.Output(c.cmdText)
	}
}

// Get a pipe to the command's stdout
func (c *command) StdoutPipe() (io.Reader, error) {
	if c.cmd != nil {
		return c.cmd.StdoutPipe()
	} else {
		return c.session.StdoutPipe()
	}
}

// Get a pipe to the command's stderr
func (c *command) StderrPipe() (io.Reader, error) {
	if c.cmd != nil {
		return c.cmd.StderrPipe()
	} else {
		return c.session.StderrPipe()
	}
}
