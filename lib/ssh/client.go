// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

//
// Client for SSH connection.
//
type Client struct {
	cfg  *ClientConfig
	conn *ssh.Client
}

//
// NewClient create a new SSH connection using predefined configuration.
//
func NewClient(cfg *ClientConfig) (cl *Client, err error) {
	if cfg == nil {
		return nil, nil
	}

	err = cfg.initialize()
	if err != nil {
		return
	}

	pkeyRaw, err := ioutil.ReadFile(cfg.PrivateKeyFile)
	if err != nil {
		err = fmt.Errorf("ssh: error when reading private key file %q: %s\n",
			cfg.PrivateKeyFile, err)
		return nil, err
	}

	sshSigner, err := ssh.ParsePrivateKey(pkeyRaw)
	if err != nil {
		return nil, fmt.Errorf("ssh: ParsePrivateKey: " + err.Error())
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.RemoteUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(sshSigner),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cl = &Client{
		cfg: cfg,
	}

	cl.conn, err = ssh.Dial("tcp", cfg.remoteAddr, sshConfig)
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
// ClientConfig.WorkingDir.
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

	remote = fmt.Sprintf("%s@%s:%s", cl.cfg.RemoteUser, cl.cfg.RemoteHost,
		remote)

	cmd := exec.Command("scp", "-i", cl.cfg.PrivateKeyFile,
		"-P", cl.cfg.remotePort, remote, local)

	cmd.Dir = cl.cfg.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ssh: Get %q: %s\n", cmd.Args, err.Error())
	}

	return nil
}

//
// Put copy a file from local storage to remote using scp command.
//
// The local file should be use the absolute path, or relative to the file in
// ClientConfig.WorkingDir.
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

	remote = fmt.Sprintf("%s@%s:%s", cl.cfg.RemoteUser, cl.cfg.RemoteHost,
		remote)

	cmd := exec.Command("scp", "-i", cl.cfg.PrivateKeyFile,
		"-P", cl.cfg.remotePort, local, remote)

	cmd.Dir = cl.cfg.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ssh: Put: %q: %s\n", cmd.Args, err.Error())
	}

	return nil
}

func (cl *Client) String() string {
	return cl.cfg.RemoteUser + "@" + cl.cfg.remoteAddr
}
