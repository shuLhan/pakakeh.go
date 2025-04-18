// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type command struct {
	buf       *bytes.Buffer
	execGoRun *exec.Cmd
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func newCommand(opts GoOptions, req *Request) (cmd *command) {
	cmd = &command{
		buf: &bytes.Buffer{},
	}
	var ctxParent = context.Background()
	cmd.ctx, cmd.ctxCancel = context.WithTimeout(ctxParent, opts.Timeout)

	var listArg = []string{`run`}
	if !req.WithoutRace {
		listArg = append(listArg, `-race`)
	}
	listArg = append(listArg, `.`)

	cmd.execGoRun = exec.CommandContext(cmd.ctx, `go`, listArg...)
	cmd.execGoRun.Env = append(cmd.execGoRun.Env, `CGO_ENABLED=1`)
	cmd.execGoRun.Env = append(cmd.execGoRun.Env, `HOME=`+userHomeDir)
	cmd.execGoRun.Env = append(cmd.execGoRun.Env,
		`PATH=/usr/bin:/usr/local/bin`)
	cmd.execGoRun.Dir = req.UnsafeRun
	cmd.execGoRun.Stdout = cmd.buf
	cmd.execGoRun.Stderr = cmd.buf
	cmd.execGoRun.WaitDelay = 100 * time.Millisecond

	return cmd
}

func newTestCommand(opts GoOptions, treq *Request) (cmd *command) {
	cmd = &command{
		buf: &bytes.Buffer{},
	}
	var ctxParent = context.Background()
	cmd.ctx, cmd.ctxCancel = context.WithTimeout(ctxParent, opts.Timeout)

	var listArg = []string{`test`, `-count=1`}
	if !treq.WithoutRace {
		listArg = append(listArg, `-race`)
	}
	listArg = append(listArg, `.`)

	cmd.execGoRun = exec.CommandContext(cmd.ctx, `go`, listArg...)
	cmd.execGoRun.Env = append(cmd.execGoRun.Env, `CGO_ENABLED=1`)
	cmd.execGoRun.Env = append(cmd.execGoRun.Env, `HOME=`+userHomeDir)
	cmd.execGoRun.Env = append(cmd.execGoRun.Env,
		`PATH=/usr/bin:/usr/local/bin`)
	cmd.execGoRun.Dir = treq.UnsafeRun
	cmd.execGoRun.Stdout = cmd.buf
	cmd.execGoRun.Stderr = cmd.buf
	cmd.execGoRun.WaitDelay = 100 * time.Millisecond

	return cmd
}

// run the command using [exec.Command.Start] and [exec.Command.Wait].
// The Start method is used to get the process ID.
// When the Start or Wait failed, it will write the error or ProcessState
// into the last line of out.
func (cmd *command) run() (out []byte) {
	defer cmd.ctxCancel()

	var err = cmd.execGoRun.Run()
	if err != nil {
		cmd.buf.WriteString("\n" + err.Error() + "\n")
	}

	out = cmd.buf.Bytes()
	return out
}
