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
	cmd      []string
	useDebug bool
	useBash  bool
	ctx      context.Context
	env      map[string]string
}

func New(cmd ...string) *Cmd {
	return &Cmd{env: make(map[string]string), cmd: cmd}
}

func splitCommand(c string) []string {
	var command []string
	items := strings.Split(c, " ")
	var isQuoted bool
	var quoted []string
	for i := 0; i < len(items); i++ {
		// for strings.Contains(items[i], `\"`) {
		// 	items[i] = strings.Replace(items[i], `\"`, `"`, -1)
		// }
		if !isQuoted && strings.Contains(items[i], `"`) {
			// find beggining of quote
			isQuoted = true
			// remove first quote in case the end quote is in the same piece
			items[i] = strings.Replace(items[i], `"`, "", 1)
		}
		if isQuoted && strings.Contains(items[i], `"`) {
			// find end of quote
			quoted = append(quoted, strings.Replace(items[i], `"`, "", -1))
			command = append(command, strings.Join(quoted, " "))
			isQuoted = false
			quoted = nil
			continue
		}
		if isQuoted {
			// add up pieces under quote
			quoted = append(quoted, items[i])
		} else {
			// otherwise theyre a regular piece of the command
			command = append(command, items[i])
		}
	}
	return command
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
	// todo bash
	cmd := exec.Command(c.cmd[0], c.cmd[1:]...)
	if c.ctx != nil {
		cmd = exec.CommandContext(c.ctx, c.cmd[0], c.cmd[1:]...)
	}
	c.Cmd = cmd

	combinedBuf := new(bytes.Buffer)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(log.WriterLevel(logrus.DebugLevel), combinedBuf)
	cmd.Stderr = io.MultiWriter(log.WriterLevel(logrus.DebugLevel), combinedBuf)
	for k, v := range c.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", k, v))
	}
	log.Log().Debugf("executing %q", c.cmd)
	err := cmd.Run()
	return combinedBuf.String(), err
}
