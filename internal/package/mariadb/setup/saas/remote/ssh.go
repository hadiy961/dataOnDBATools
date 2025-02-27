package saas_remote

import (
	"dbaTools/internal/logger"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConnection handles connecting to a remote server via SSH
func SSHConnection(serverIP, port, user, password string, log *logger.Logger) (*ssh.Client, error) {
	log.Info(fmt.Sprintf("Establishing SSH connection to %s@%s:%s", user, serverIP, port))

	// Parse port to int
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use proper host key verification
		Timeout:         10 * time.Second,
	}

	// Connect to the remote server
	hostPort := fmt.Sprintf("%s:%d", serverIP, portInt)
	client, err := ssh.Dial("tcp", hostPort, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	log.Success(fmt.Sprintf("SSH connection established to %s", hostPort))
	return client, nil
}

// ExecuteRemoteCommand runs a command on the remote server and returns the output
func ExecuteRemoteCommand(client *ssh.Client, command string, log *logger.Logger) (string, error) {
	log.Info(fmt.Sprintf("Executing remote command: %s", command))

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Run the command and capture output
	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}
