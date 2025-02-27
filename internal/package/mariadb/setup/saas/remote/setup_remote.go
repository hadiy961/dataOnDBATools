package saas_remote

import (
	"dbaTools/internal/logger"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// SetupSAASRemote initializes and validates the SAAS environment on a remote server
func Setup(log *logger.Logger) error {
	log.Info("Starting remote SAAS setup process")

	// Initialize and run the remote server form
	formModel := NewRemoteServerFormModel()
	p := tea.NewProgram(formModel)

	model, err := p.Run()
	if err != nil {
		log.Error("Form error", err)
		return err
	}

	// Get form data after submission
	m, ok := model.(RemoteServerFormModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if !m.submitted {
		return fmt.Errorf("setup cancelled")
	}

	// Extract server details from form values
	serverIP := m.values["ip_address"]
	sshPort := m.values["ssh_port"]
	sshUser := m.values["ssh_user"]
	sshPass := m.values["ssh_pass"]

	log.Info(fmt.Sprintf("Connecting to remote server: %s:%s:%s", serverIP, sshPort, sshPass))

	// Initialize spinner for visual feedback
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "[ Remote SAAS Setup ]"
	s.Color("green")
	s.Start()
	defer s.Stop()

	// Connect to the remote server
	s.Suffix = " Establishing SSH connection...\n"
	client, err := SSHConnection(serverIP, sshPort, sshUser, sshPass, log)
	if err != nil {
		s.Stop()
		log.Error("SSH connection failed", err)
		return fmt.Errorf("failed to connect to remote server: %w", err)
	}
	defer client.Close()

	// Check if MariaDB is installed
	s.Suffix = " Checking if MariaDB is installed..."
	output, err := ExecuteRemoteCommand(client, "mariadb --version || mysql --version", log)
	if err != nil {
		s.Stop()
		errorMsg := "MariaDB is not installed on the remote server. Please install MariaDB before continuing."
		log.Error(errorMsg, err)
		return fmt.Errorf(errorMsg)
	}

	// MariaDB is installed
	log.Info(fmt.Sprintf("MariaDB is installed: %s", strings.TrimSpace(output)))

	// Check if MariaDB service is running
	s.Suffix = " Checking if MariaDB service is running..."
	serviceOutput, err := ExecuteRemoteCommand(client, "systemctl is-active mariadb || systemctl is-active mysql", log)
	isRunning := err == nil && strings.TrimSpace(serviceOutput) == "active"

	if !isRunning {
		s.Suffix = " Starting MariaDB service..."
		log.Info("MariaDB service is not running, attempting to start")

		_, err := ExecuteRemoteCommand(client, "sudo systemctl start mariadb || sudo systemctl start mysql", log)
		if err != nil {
			s.Stop()
			log.Error("Failed to start MariaDB service", err)
			return fmt.Errorf("failed to start MariaDB service: %w", err)
		}

		// Verify service started
		_, err = ExecuteRemoteCommand(client, "systemctl is-active mariadb || systemctl is-active mysql", log)
		if err != nil {
			s.Stop()
			log.Error("MariaDB service failed to start", err)
			return fmt.Errorf("MariaDB service failed to start")
		}

		log.Success("MariaDB service started successfully")
	} else {
		log.Info("MariaDB service is already running")
	}

	// Configure MariaDB to start at boot
	s.Suffix = " Configuring MariaDB to start at boot..."
	_, err = ExecuteRemoteCommand(client, "sudo systemctl enable mariadb || sudo systemctl enable mysql", log)
	if err != nil {
		log.Warning("Failed to enable MariaDB service at boot")
	} else {
		log.Success("MariaDB service enabled to start at boot")
	}

	// Setup complete
	s.Stop()
	log.Success(fmt.Sprintf("Remote MariaDB setup completed on %s", serverIP))

	return nil
}
