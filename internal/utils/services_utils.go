package utils

import (
	"bytes"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// NewServiceChecker creates a new service checker instance
func NewServiceChecker(isSSH bool, host string, port int, user, password string) *ServiceChecker {
	return &ServiceChecker{
		isSSH:    isSSH,
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

// CheckServices checks for all supported services
func (sc *ServiceChecker) CheckServices(serverID int) ([]ServiceInfo, error) {
	// Create spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Service Check ] "
	if sc.isSSH {
		spin.Suffix = fmt.Sprintf(" Scanning for services on %s@%s:%d...", sc.user, sc.host, sc.port)
	} else {
		spin.Suffix = " Scanning for local services..."
	}
	spin.Color("cyan")
	spin.Start()
	defer spin.Stop()

	now := time.Now()
	services := []ServiceInfo{}

	// Check MariaDB
	spin.Suffix = " Checking MariaDB service..."
	mariaDB, err := sc.checkMariaDB(serverID, now)
	if err == nil {
		services = append(services, *mariaDB)
	}

	// Check MaxScale
	spin.Suffix = " Checking MaxScale proxy service..."
	maxScale, err := sc.checkMaxScale(serverID, now)
	if err == nil {
		services = append(services, *maxScale)
	}

	return services, nil
}

// executeCommand executes a command either locally or via SSH
func (sc *ServiceChecker) executeCommand(command string) (string, error) {
	if sc.isSSH {
		// For SSH execution, use SSH client
		client, err := ssh.GetSSHClient(sc.host, sc.port, sc.user, sc.password)
		if err != nil {
			return "", fmt.Errorf("SSH connection error: %w", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			return "", fmt.Errorf("failed to create SSH session: %w", err)
		}
		defer session.Close()

		var out bytes.Buffer
		session.Stdout = &out
		err = session.Run(command)
		if err != nil {
			return "", fmt.Errorf("command execution failed: %w", err)
		}

		return strings.TrimSpace(out.String()), nil
	} else {
		// For local execution
		cmdParts := []string{"/bin/sh", "-c", command}
		output, err := exec.Command(cmdParts[0], cmdParts[1:]...).Output()
		if err != nil {
			return "", fmt.Errorf("command execution failed: %w", err)
		}

		return strings.TrimSpace(string(output)), nil
	}
}

// checkMariaDB checks for MariaDB service with optimized performance
func (sc *ServiceChecker) checkMariaDB(serverID int, now time.Time) (*ServiceInfo, error) {
	service := &ServiceInfo{
		ServerID:    serverID,
		ServiceName: "MariaDB",
		ServiceType: "database",
		LastChecked: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Port:        3306, // Set default port
	}

	// Single command to check installation, status, and auto-start
	// This reduces the number of SSH calls from 3 to 1
	checkCmd := `
INSTALLED=$(command -v mysqld >/dev/null 2>&1 || command -v mariadbd >/dev/null 2>&1; echo $?)
if [ "$INSTALLED" != "0" ]; then echo "not-installed"; exit 1; fi
STATUS=$(systemctl is-active mariadb 2>/dev/null || systemctl is-active mysql 2>/dev/null)
ENABLED=$(systemctl is-enabled mariadb 2>/dev/null || systemctl is-enabled mysql 2>/dev/null 2>/dev/null || echo "unknown")
echo "$STATUS $ENABLED"`

	checkOutput, err := sc.executeCommand(checkCmd)
	if err != nil || strings.Contains(checkOutput, "not-installed") {
		return nil, fmt.Errorf("MariaDB not installed")
	}

	// Parse status and enabled state
	parts := strings.Fields(strings.TrimSpace(checkOutput))
	if len(parts) > 0 && parts[0] == "active" {
		service.Status = "Running"
	} else {
		service.Status = "Stopped"
	}

	if len(parts) > 1 && parts[1] == "enabled" {
		service.AutoStart = true
	}

	// Single command to get version, install path and config file
	// This reduces the number of SSH calls from 3 to 1
	infoCmd := `
VERSION=$(mariadb --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -z "$VERSION" ]; then VERSION=$(mysql --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); fi
INSTALL_PATH="/usr"
CONFIG_PATH="/etc/my.cnf"
if [ ! -f "$CONFIG_PATH" ]; then 
    for f in /etc/mysql/my.cnf /etc/mysql/mariadb.cnf; do
        if [ -f "$f" ]; then CONFIG_PATH="$f"; break; fi
    done
fi
echo "$VERSION $INSTALL_PATH $CONFIG_PATH"
`
	infoOutput, _ := sc.executeCommand(infoCmd)
	infoParts := strings.SplitN(strings.TrimSpace(infoOutput), " ", 3)

	if len(infoParts) > 0 && infoParts[0] != "" {
		service.Version = infoParts[0]
	}

	if len(infoParts) > 1 {
		service.InstallPath = infoParts[1]
	}

	if len(infoParts) > 2 {
		service.ConfigFile = infoParts[2]
	}

	return service, nil
}

func (sc ServiceChecker) checkMaxScale(serverID int, now time.Time) (*ServiceInfo, error) {
	// Initialize with defaults
	service := &ServiceInfo{
		ServerID:    serverID,
		ServiceName: "MaxScale",
		ServiceType: "proxy",
		LastChecked: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Port:        4006, // Default port
		Status:      "stopped",
	}

	// Check if MaxScale is installed
	if output, err := sc.executeCommand("which maxscale 2>/dev/null"); err != nil || output == "" {
		return nil, fmt.Errorf("MaxScale not installed")
	}

	// Combined command to get multiple pieces of information at once
	infoCmd := `
    VERSION=$(maxscale --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
    STATUS=$(systemctl is-active maxscale 2>/dev/null)
    ENABLED=$(systemctl is-enabled maxscale 2>/dev/null)
    PORT=$(ss -tulpn 2>/dev/null | grep maxscale | grep LISTEN | awk '{print $5}' | awk -F: '{print $NF}' | head -1)
    CONFIG=$(find /etc -name maxscale.cnf 2>/dev/null | head -1)
    BIN=$(which maxscale 2>/dev/null)
    echo "$VERSION|$STATUS|$ENABLED|$PORT|$CONFIG|$BIN"
    `

	output, err := sc.executeCommand(infoCmd)
	if err != nil {
		// Still return basic service info even if detailed check fails
		return service, fmt.Errorf("failed to get MaxScale details: %w", err)
	}

	parts := strings.Split(output, "|")
	if len(parts) >= 6 {
		// Parse version
		if parts[0] != "" {
			service.Version = strings.TrimSpace(parts[0])
		}

		// Parse status
		if strings.TrimSpace(parts[1]) == "active" {
			service.Status = "Running"
		}

		// Parse auto-start
		service.AutoStart = strings.TrimSpace(parts[2]) == "enabled"

		// Parse port
		if parts[3] != "" {
			if port, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil && port > 0 {
				service.Port = port
			}
		}

		// Parse config file
		if parts[4] != "" {
			service.ConfigFile = strings.TrimSpace(parts[4])
		}

		// Parse binary path
		if parts[5] != "" {
			service.InstallPath = strings.TrimSpace(parts[5])
		}
	}

	return service, nil
}

// Colorize status text based on service status
func ColorizeStatus(status string) string {
	switch status {
	case "running":
		return "\033[32mRunning\033[0m" // Green
	case "stopped":
		return "\033[31mStopped\033[0m" // Red
	case "installed":
		return "\033[33mInstalled\033[0m" // Yellow
	default:
		return status
	}
}

// FormatServicePort formats the port display
func FormatServicePort(port int) string {
	if port == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", port)
}

// FormatAutoStart formats the autostart display
func FormatAutoStart(autoStart bool) string {
	if autoStart {
		return "\033[32mYes\033[0m" // Green
	}
	return "\033[31mNo\033[0m" // Red
}
