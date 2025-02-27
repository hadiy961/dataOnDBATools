package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClientManager handles SSH connections
func GetSSHClient(host string, port int, user, password string) (*ssh.Client, error) {
	// Create SSH client config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	return client, nil
}
