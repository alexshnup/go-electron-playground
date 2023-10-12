package main

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func ConnectToSSH(config SSHConfig) (*ssh.Client, error) {
	fmt.Println("Connecting to SSH server...")

	var authMethod ssh.AuthMethod

	if config.Password != "" {
		authMethod = ssh.Password(config.Password)
	} else if len(config.PrivateKey) > 0 {
		var signer ssh.Signer
		var err error

		if config.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(config.PrivateKey, []byte(config.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(config.PrivateKey)
		}

		if err != nil {
			return nil, err
		}
		authMethod = ssh.PublicKeys(signer)
	}
	clientConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For simplicity. In production, validate host key.
	}

	client, err := ssh.Dial("tcp", config.Host+":"+config.Port, clientConfig)
	return client, err
}

func ExecuteSSHCommand(command string, client *ssh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err = session.Run(command)
	return stdoutBuf.String(), err
}
