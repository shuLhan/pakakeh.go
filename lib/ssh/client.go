// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/shuLhan/share/lib/ssh/config"
)

// Client for SSH connection.
type Client struct {
	*ssh.Client

	cfg    *config.Section
	stdout io.Writer
	stderr io.Writer
}

// NewClientFromConfig create a new SSH connection using predefined
// configuration. This function may dial twice to find appropriate
// authentication method when SSH_AUTH_SOCK environment variable is
// set and IdentityFile directive is specified in Host section.
func NewClientFromConfig(cfg *config.Section) (cl *Client, err error) {
	if cfg == nil {
		return nil, nil
	}

	var (
		logp       = `NewClient`
		remoteAddr = fmt.Sprintf(`%s:%s`, cfg.Hostname, cfg.Port)
		sshConfig  = &ssh.ClientConfig{
			User:            cfg.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		agentClient agent.ExtendedAgent
	)

	cl = &Client{
		cfg:    cfg,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	sshAgentSockPath := cfg.GetIdentityAgent()
	if len(sshAgentSockPath) > 0 {
		sshAgentSock, err := net.Dial("unix", sshAgentSockPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}
		agentClient = agent.NewClient(sshAgentSock)
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		}
		sshClient, err := ssh.Dial("tcp", remoteAddr, sshConfig)
		if err == nil {
			cl.Client = sshClient
			return cl, nil
		}
		if len(cfg.IdentityFile) == 0 || !strings.HasSuffix(err.Error(), "no supported methods remain") {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}
	}

	sshConfig.Auth = []ssh.AuthMethod{
		ssh.PublicKeysCallback(cfg.Signers),
	}
	sshClient, err := ssh.Dial("tcp", remoteAddr, sshConfig)
	if err == nil {
		cl.Client = sshClient
		return cl, nil
	}
	return nil, fmt.Errorf("%s: %w", logp, err)
}

// Execute a command on remote server.
func (cl *Client) Execute(cmd string) (err error) {
	sess, err := cl.NewSession()
	if err != nil {
		return fmt.Errorf("ssh: NewSession: " + err.Error())
	}

	sess.Stdout = cl.stdout
	sess.Stderr = cl.stderr

	for k, v := range cl.cfg.Environments {
		err = sess.Setenv(k, v)
		if err != nil {
			log.Printf("Execute: Setenv %q=%q:%s\n", k, v, err.Error())
		}
	}

	err = sess.Run(cmd)
	if err != nil {
		err = fmt.Errorf("ssh: Run %q: %s", cmd, err.Error())
	}

	sess.Close()

	return err
}

// ScpGet copy file from remote into local storage using scp.
//
// The local file should be use the absolute path, or relative to the file in
// config.Section.WorkingDir.
func (cl *Client) ScpGet(remote, local string) (err error) {
	logp := "ScpGet"

	if len(remote) == 0 {
		return fmt.Errorf("%s: empty remote file", logp)
	}
	if len(local) == 0 {
		return fmt.Errorf("%s: empty local file", logp)
	}

	remote = fmt.Sprintf("%s@%s:%s", cl.cfg.User, cl.cfg.Hostname, remote)

	args := []string{"-r", "-P", cl.cfg.Port}
	if len(cl.cfg.PrivateKeyFile) > 0 {
		args = append(args, "-i")
		args = append(args, cl.cfg.PrivateKeyFile)
	}
	args = append(args, remote)
	args = append(args, local)

	cmd := exec.Command("scp", args...)

	cmd.Dir = cl.cfg.WorkingDir
	cmd.Stdout = cl.stdout
	cmd.Stderr = cl.stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %q: %s", logp, cmd.Args, err.Error())
	}

	return nil
}

// ScpPut copy a file from local storage to remote using scp command.
//
// The local file should be use the absolute path, or relative to the file in
// config.Section's WorkingDir.
func (cl *Client) ScpPut(local, remote string) (err error) {
	logp := "ScpPut"

	if len(local) == 0 {
		return fmt.Errorf("%s: empty local file", logp)
	}
	if len(remote) == 0 {
		return fmt.Errorf("%s: empty remote file", logp)
	}

	remote = fmt.Sprintf("%s@%s:%s", cl.cfg.User, cl.cfg.Hostname, remote)

	args := []string{"-r", "-P", cl.cfg.Port}
	if len(cl.cfg.PrivateKeyFile) > 0 {
		args = append(args, "-i")
		args = append(args, cl.cfg.PrivateKeyFile)
	}
	args = append(args, local)
	args = append(args, remote)

	cmd := exec.Command("scp", args...)

	cmd.Dir = cl.cfg.WorkingDir
	cmd.Stdout = cl.stdout
	cmd.Stderr = cl.stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %q: %s", logp, cmd.Args, err.Error())
	}

	return nil
}

// SetSessionOutputError set the standard output and error for future remote
// execution.
func (cl *Client) SetSessionOutputError(stdout, stderr io.Writer) {
	if stdout != nil {
		cl.stdout = stdout
	}
	if stderr != nil {
		cl.stderr = stderr
	}
}

func (cl *Client) String() string {
	return cl.cfg.User + "@" + cl.cfg.Hostname + ":" + cl.cfg.Port
}
