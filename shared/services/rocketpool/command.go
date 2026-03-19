package rocketpool

import (
	"io"
	"os/exec"
)

// A command to be executed locally
type command struct {
	cmd     *exec.Cmd
	cmdText string
}

func (c *Client) newCommand(cmdText string) (*command, error) {
	return &command{
		cmd:     exec.Command("sh", "-c", cmdText),
		cmdText: cmdText,
	}, nil
}

func (c *command) Close() error { return nil }

func (c *command) Run() error { return c.cmd.Run() }

func (c *command) Start() error { return c.cmd.Start() }

func (c *command) Wait() error { return c.cmd.Wait() }

func (c *command) SetStdin(r io.Reader)  { c.cmd.Stdin = r }
func (c *command) SetStdout(w io.Writer) { c.cmd.Stdout = w }
func (c *command) SetStderr(w io.Writer) { c.cmd.Stderr = w }

func (c *command) Output() ([]byte, error) { return c.cmd.Output() }

func (c *command) StdoutPipe() (io.Reader, error) { return c.cmd.StdoutPipe() }

func (c *command) StderrPipe() (io.Reader, error) { return c.cmd.StderrPipe() }

func (c *command) OutputPipes() (io.Reader, io.Reader, error) {
	cmdOut, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	cmdErr, err := c.StderrPipe()
	return cmdOut, cmdErr, err
}
