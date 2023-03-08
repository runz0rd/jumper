package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Cmd struct {
	useDebug       bool
	useBash        bool
	ctx            context.Context
	env            map[string]string
	stdout, stderr io.Writer
}

func New(stdout, stderr io.Writer) *Cmd {
	return &Cmd{env: make(map[string]string), stdout: stdout, stderr: stderr}
}

func (c *Cmd) WithContext(ctx context.Context) *Cmd {
	c.ctx = ctx
	return c
}

func (c *Cmd) WithEnv(env map[string]string) *Cmd {
	c.env = env
	return c
}

func (c *Cmd) WithDebug() *Cmd {
	c.useDebug = true
	return c
}

func (c *Cmd) WithBash() *Cmd {
	c.useBash = true
	return c
}

func (c *Cmd) String(format string, a ...any) (string, error) {
	return c.exec(format, a...)
}

func (c *Cmd) Slice(format string, a ...any) ([]string, error) {
	out, err := c.exec(format, a...)
	return strings.Split(out, "\n"), err
}

func (c *Cmd) exec(format string, a ...any) (string, error) {
	cmdString := fmt.Sprintf(format, a...)
	if c.useBash {
		cmdString = fmt.Sprintf("bash -c %q", cmdString)
	}
	cmdSlice := strings.Split(cmdString, " ")
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
	if c.ctx != nil {
		cmd = exec.CommandContext(c.ctx, cmdSlice[0], cmdSlice[1:]...)
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(buf, c.stdout)
	cmd.Stderr = c.stderr
	for k, v := range c.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", k, v))
	}
	if c.useDebug {
		fmt.Fprint(c.stdout, fmt.Sprintf("executing %q", cmdString))
	}
	return buf.String(), cmd.Run()
}
