// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

//
// Client for SSH connection.
//
type Client struct {
	cfg  *ConfigSection
	conn *ssh.Client
}

//
// NewClient create a new SSH connection using predefined configuration.
//
func NewClient(cfg *ConfigSection) (cl *Client, err error) {
	if cfg == nil {
		return nil, nil
	}

	var (
		logp             = "NewClient"
		sshAgentSockPath string
		sshAgentSock     net.Conn
		sshConfig        *ssh.ClientConfig
		remoteAddr       string
		agentClient      agent.ExtendedAgent
	)

	cfg.postConfig("")

	sshConfig = &ssh.ClientConfig{
		User:            cfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshAgentSockPath = os.Getenv("SSH_AUTH_SOCK")
	if len(sshAgentSockPath) > 0 {
		sshAgentSock, err = net.Dial("unix", sshAgentSockPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		agentClient = agent.NewClient(sshAgentSock)

		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		}
	} else {
		err = cfg.generateSigners(nil)
		if err != nil {
			return nil, err
		}

		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(cfg.signers...),
		}
	}

	cl = &Client{
		cfg: cfg,
	}

	remoteAddr = fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)

	cl.conn, err = ssh.Dial("tcp", remoteAddr, sshConfig)
	if err != nil {
		err = fmt.Errorf("ssh: Dial: " + err.Error())
		return nil, err
	}

	return cl, nil
}

//
// Execute a command on remote server.
//
func (cl *Client) Execute(cmd string) (err error) {
	sess, err := cl.conn.NewSession()
	if err != nil {
		return fmt.Errorf("ssh: NewSession: " + err.Error())
	}

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

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

//
// Get copy file from remote into local storage.
//
// The local file should be use the absolute path, or relative to the file in
// ConfigSection's workingDir.
//
func (cl *Client) Get(remote, local string) (err error) {
	if len(remote) == 0 {
		log.Println("ssh: Get: empty remote file")
		return nil
	}
	if len(local) == 0 {
		log.Println("ssh: Get: empty local file")
		return nil
	}

	remote = fmt.Sprintf("%s@%s:%s", cl.cfg.User, cl.cfg.Hostname, remote)

	args := []string{"-r", "-P", cl.cfg.stringPort}
	if len(cl.cfg.privateKeyFile) > 0 {
		args = append(args, "-i")
		args = append(args, cl.cfg.privateKeyFile)
	}
	args = append(args, remote)
	args = append(args, local)

	cmd := exec.Command("scp", args...)

	cmd.Dir = cl.cfg.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ssh: Get %q: %s", cmd.Args, err.Error())
	}

	return nil
}

//
// Put copy a file from local storage to remote using scp command.
//
// The local file should be use the absolute path, or relative to the file in
// ConfigSection's workingDir.
//
func (cl *Client) Put(local, remote string) (err error) {
	if len(local) == 0 {
		log.Println("ssh: Put: empty local file")
		return nil
	}
	if len(remote) == 0 {
		log.Println("ssh: Put: empty remote file")
		return nil
	}

	remote = fmt.Sprintf("%s@%s:%s", cl.cfg.User, cl.cfg.Hostname, remote)

	args := []string{"-r", "-P", cl.cfg.stringPort}
	if len(cl.cfg.privateKeyFile) > 0 {
		args = append(args, "-i")
		args = append(args, cl.cfg.privateKeyFile)
	}
	args = append(args, local)
	args = append(args, remote)

	cmd := exec.Command("scp", args...)

	cmd.Dir = cl.cfg.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ssh: Put: %q: %s", cmd.Args, err.Error())
	}

	return nil
}

func (cl *Client) String() string {
	return cl.cfg.User + "@" + cl.cfg.Hostname + ":" + cl.cfg.stringPort
}
