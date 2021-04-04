// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
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

	cfg.postConfig("")

	err = cfg.generateSigners()
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(cfg.signers...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cl = &Client{
		cfg: cfg,
	}

	remoteAddr := fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)

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

	cmd := exec.Command("scp", "-r", "-i", cl.cfg.privateKeyFile,
		"-P", cl.cfg.stringPort, remote, local)

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

	cmd := exec.Command("scp", "-r", "-i", cl.cfg.privateKeyFile,
		"-P", cl.cfg.stringPort, local, remote)

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
