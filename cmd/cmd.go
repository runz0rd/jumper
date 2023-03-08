package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/runz0rd/jumper/log"
	"github.com/sirupsen/logrus"
)

type Cmd struct {
	*exec.Cmd
	cmd      string
	useDebug bool
	useBash  bool
	ctx      context.Context
	env      map[string]string
}

func New(format string, a ...any) *Cmd {
	return &Cmd{env: make(map[string]string), cmd: fmt.Sprintf(format, a...)}
}

func (c *Cmd) WithContext(ctx context.Context) *Cmd {
	c.ctx = ctx
	return c
}

func (c *Cmd) WithEnv(env map[string]string) *Cmd {
	c.env = env
	return c
}

func (c *Cmd) WithBash() *Cmd {
	c.useBash = true
	return c
}

func (c *Cmd) String() (string, error) {
	return c.exec()
}

func (c *Cmd) Slice() ([]string, error) {
	out, err := c.exec()
	return strings.Split(out, "\n"), err
}

func (c *Cmd) Kill() error {
	return c.Process.Kill()
}

func (c *Cmd) exec() (string, error) {
	cmdString := c.cmd
	if c.useBash {
		cmdString = fmt.Sprintf(`bash -c %v`, cmdString)
	}
	cmdSlice := strings.Split(cmdString, " ")
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
	if c.ctx != nil {
		cmd = exec.CommandContext(c.ctx, cmdSlice[0], cmdSlice[1:]...)
	}
	c.Cmd = cmd

	combinedBuf := new(bytes.Buffer)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(log.WriterLevel(logrus.DebugLevel), combinedBuf)
	cmd.Stderr = io.MultiWriter(log.WriterLevel(logrus.DebugLevel), combinedBuf)
	for k, v := range c.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", k, v))
	}
	log.Log().Debugf("executing %q", cmdString)
	err := cmd.Run()
	return combinedBuf.String(), err
}
