package saas_lokal

import (
	"dbaTools/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// ServiceStatus represents the current state of a service
type ServiceStatus struct {
	IsInstalled bool
	IsRunning   bool
	Version     string
	Error       error
}

// MariaDBStatus checks MariaDB service status in detail
func checkMariaDBStatus(log *logger.Logger, s *spinner.Spinner) (*ServiceStatus, error) {
	status := &ServiceStatus{}

	// Update spinner text
	s.Suffix = " Checking MariaDB installation... \n"
	time.Sleep(101 * time.Millisecond) // Give users a chance to see the message

	// Check if MariaDB is installed by looking for the binary
	mariaPath, err := exec.LookPath("mariadb")
	if err != nil {
		log.Warning("MariaDB binary not found in PATH")
		status.Error = fmt.Errorf("MariaDB is not installed")
		return status, nil
	}
	log.Info(fmt.Sprintf("Found MariaDB binary at: %s", mariaPath))
	status.IsInstalled = true

	// Update spinner for version check
	s.Suffix = " Detecting MariaDB version... \n"
	time.Sleep(101 * time.Millisecond)

	// Get MariaDB version
	versionCmd := exec.Command("mariadb", "--version")
	versionOutput, err := versionCmd.Output()
	if err == nil {
		version := string(versionOutput)
		status.Version = strings.TrimSpace(strings.Split(version, "\n")[0])
		log.Info(fmt.Sprintf("Detected MariaDB version: %s", status.Version))
	} else {
		log.Warning("Failed to get MariaDB version")
	}

	// Update spinner for service check
	s.Suffix = " Checking MariaDB service status... \n"
	time.Sleep(101 * time.Millisecond)

	// Check if MariaDB service is running
	serviceCmd := exec.Command("systemctl", "is-active", "mariadb")
	if output, err := serviceCmd.Output(); err == nil {
		status.IsRunning = strings.TrimSpace(string(output)) == "active"
		if status.IsRunning {
			log.Info("MariaDB service is running")
		} else {
			log.Warning("MariaDB service is not running")
		}
	} else {
		log.Warning("Failed to check MariaDB service status")
	}

	return status, nil
}

// SetupSAAS initializes and validates the SAAS environment
func Setup(log *logger.Logger) error {
	log.Info("Starting SAAS setup process")

	// Initialize spinner
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "[ SAAS Setup ]"
	s.Color("green")
	s.Start()
	defer s.Stop()

	// Check if running as root
	s.Suffix = " Checking privileges..."
	time.Sleep(101 * time.Millisecond)

	if os.Geteuid() != 0 {
		s.Stop()
		log.Error("Setup must be run as root", fmt.Errorf("insufficient privileges"))
		return fmt.Errorf("this setup must be run as root")
	}

	s.Suffix = " Initializing MariaDB service check..."
	time.Sleep(101 * time.Millisecond)

	status, err := checkMariaDBStatus(log, s)
	if err != nil {
		s.Stop()
		log.Error("Failed to check MariaDB status", err)
		return fmt.Errorf("failed to check MariaDB status: %w", err)
	}

	// Stop spinner temporarily for status report
	s.Stop()

	// Report MariaDB status
	log.Info(fmt.Sprintf("Installation Status: %v", status.IsInstalled))
	if status.IsInstalled {
		log.Info(fmt.Sprintf("Service Running: %v", status.IsRunning))
		log.Info(fmt.Sprintf("Version: %s", status.Version))
	}

	// Restart spinner for remaining tasks
	s.Start()

	// Handle different scenarios
	if !status.IsInstalled {
		s.Stop()
		err := fmt.Errorf("MariaDB is not installed. Please install MariaDB before continuing")
		log.Error("Setup failed", err)
		return err
	}

	if !status.IsRunning {
		s.Suffix = " Starting MariaDB service... \n"
		log.Info("Attempting to start MariaDB service")

		startCmd := exec.Command("systemctl", "start", "mariadb")
		if err := startCmd.Run(); err != nil {
			s.Stop()
			log.Error("Failed to start MariaDB service", err)
			return fmt.Errorf("failed to start MariaDB service: %w", err)
		}

		// Verify service started successfully
		s.Suffix = " Verifying service status... \n"
		time.Sleep(2 * time.Second) // Give service time to start

		status, _ = checkMariaDBStatus(log, s)
		if !status.IsRunning {
			s.Stop()
			err := fmt.Errorf("failed to start MariaDB service")
			log.Error("Service verification failed", err)
			return err
		}
		log.Success("MariaDB service started successfully")
	}

	// Enable MariaDB service to start on boot if not already enabled
	s.Suffix = " Configuring startup settings... \n"
	log.Info("Ensuring MariaDB starts on boot")

	enableCmd := exec.Command("systemctl", "enable", "mariadb")
	if err := enableCmd.Run(); err != nil {
		log.Warning("Failed to enable MariaDB service on startup")
	} else {
		log.Success("MariaDB service enabled for startup")
	}

	s.Stop()
	log.Success("MariaDB service check completed successfully")
	return nil
}
