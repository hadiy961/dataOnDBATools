package utils

import (
	"bytes"
	"context"
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

// checkMariaDB checks for MariaDB service with improved reliability
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

	// Check if MariaDB/MySQL is installed by checking for the binary
	checkInstallCmd := `command -v mysqld >/dev/null 2>&1 || command -v mariadbd >/dev/null 2>&1; echo $?`
	installOutput, err := sc.executeCommand(checkInstallCmd)
	if err != nil || strings.TrimSpace(installOutput) != "0" {
		return nil, fmt.Errorf("MariaDB not installed")
	}

	// Robust service status check using multiple methods
	serviceStatusCmd := `
if command -v systemctl >/dev/null 2>&1; then
    # Using systemctl (systemd systems)
    ACTIVE=$(systemctl is-active mariadb 2>/dev/null || systemctl is-active mysql 2>/dev/null)
    ENABLED=$(systemctl is-enabled mariadb 2>/dev/null || systemctl is-enabled mysql 2>/dev/null 2>/dev/null || echo "unknown")
    echo "${ACTIVE} ${ENABLED}"
else
    # Using service command (init systems)
    if command -v service >/dev/null 2>&1; then
        if service mariadb status >/dev/null 2>&1 || service mysql status >/dev/null 2>&1; then
            echo "active unknown"
        else
            echo "inactive unknown"
        fi
    else
        # Last resort: process check
        pgrep -f 'mysqld|mariadbd' >/dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo "active unknown"
        else
            echo "inactive unknown"
        fi
    fi
fi`

	statusOutput, _ := sc.executeCommand(serviceStatusCmd)
	statusFields := strings.Fields(strings.TrimSpace(statusOutput))

	// Parse status output
	if len(statusFields) >= 1 {
		if statusFields[0] == "active" {
			service.Status = "running"
		} else {
			service.Status = "stopped"
		}

		// Parse enabled status if available
		if len(statusFields) >= 2 && statusFields[1] == "enabled" {
			service.AutoStart = true
		}
	} else {
		// Additional fallback - direct process check
		processCheckCmd := `pgrep -f 'mysqld|mariadbd' >/dev/null 2>&1 && echo "running" || echo "stopped"`
		processOutput, _ := sc.executeCommand(processCheckCmd)
		if strings.TrimSpace(processOutput) == "running" {
			service.Status = "running"
		} else {
			service.Status = "stopped"
		}
	}

	// Get listening port, config file, and binary path
	portConfigCmd := `
# Find listening port
PORT=$(ss -tulpn 2>/dev/null | grep -E 'mysql|mariadb' | grep LISTEN | awk '{print $5}' | awk -F: '{print $NF}' | head -1)
if [ -z "$PORT" ]; then
    # Fallback - check common port from config files
    PORT=$(grep -E 'port\s*=\s*[0-9]+' /etc/my.cnf /etc/mysql/my.cnf /etc/mysql/mariadb.cnf 2>/dev/null | head -1 | grep -o '[0-9]\+')
    if [ -z "$PORT" ]; then
        PORT="3306" # Default port
    fi
fi

# Find config file - prioritize files that actually exist
CONFIG=""
for CONFIG_FILE in /etc/my.cnf /etc/mysql/my.cnf /etc/mysql/mariadb.cnf /etc/my.cnf.d/server.cnf /etc/mysql/mariadb/50-server.cnf /etc/mysql/mariadb.conf.d/50-server.cnf /etc/mysql/conf.d/mysql.cnf /etc/opt/rh/rh-mariadb*/my.cnf; do
    if [ -f "$CONFIG_FILE" ]; then
        CONFIG=$CONFIG_FILE
        break
    fi
done

# If no specific config file found, check directories for includes
if [ -z "$CONFIG" ]; then
    for CONFIG_DIR in /etc/my.cnf.d/ /etc/mysql/conf.d/ /etc/mysql/mariadb.conf.d/; do
        if [ -d "$CONFIG_DIR" ] && [ "$(ls -A $CONFIG_DIR 2>/dev/null)" ]; then
            CONFIG="$CONFIG_DIR"
            break
        fi
    done
fi

# Find binary path and data directory (for installation path)
BINARY=$(command -v mysqld 2>/dev/null || command -v mariadbd 2>/dev/null)
INSTALL_PATH=""

# If binary found, try to get real installation path 
if [ -n "$BINARY" ]; then
    # Follow symlinks to get real binary location
    REAL_BINARY=$(readlink -f "$BINARY" 2>/dev/null)
    if [ -n "$REAL_BINARY" ]; then
        # Get parent directory of bin (typical installation structure)
        INSTALL_PATH=$(dirname "$(dirname "$REAL_BINARY")" 2>/dev/null)
    fi
    
    # If that didn't work, try package manager
    if [ -z "$INSTALL_PATH" ] || [ "$INSTALL_PATH" = "/" ]; then
        if command -v dpkg >/dev/null 2>&1; then
            # Debian/Ubuntu
            INSTALL_PATH=$(dpkg -L mariadb-server 2>/dev/null | grep -m1 "^/usr/share/mariadb" || dpkg -L mysql-server 2>/dev/null | grep -m1 "^/usr/share/mysql")
        elif command -v rpm >/dev/null 2>&1; then
            # Red Hat/CentOS
            INSTALL_PATH=$(rpm -ql MariaDB-server 2>/dev/null | grep -m1 "^/usr/share/mariadb" || rpm -ql mysql-server 2>/dev/null | grep -m1 "^/usr/share/mysql")
        fi
    fi
    
    # If still no install path, use binary directory as fallback
    if [ -z "$INSTALL_PATH" ] || [ "$INSTALL_PATH" = "/" ]; then
        INSTALL_PATH=$(dirname "$BINARY")
    fi
fi

echo "$PORT:$CONFIG:$INSTALL_PATH"
`
	infoOutput, _ := sc.executeCommand(portConfigCmd)

	if infoOutput != "" {
		parts := strings.Split(infoOutput, ":")
		if len(parts) >= 1 && parts[0] != "" {
			if port, err := strconv.Atoi(parts[0]); err == nil && port > 0 {
				service.Port = port
			}
		}

		// Handle config file path
		if len(parts) >= 2 && parts[1] != "" {
			service.ConfigFile = parts[1]

			// Log the config file path for debugging
			if sc.isSSH {
				configVerifyCmd := fmt.Sprintf("ls -la %s 2>/dev/null || echo 'Not found'", service.ConfigFile)
				if output, _ := sc.executeCommand(configVerifyCmd); !strings.Contains(output, "Not found") {
					// Config file exists, keeping it
				} else if strings.HasSuffix(service.ConfigFile, "/") {
					// It's a directory, try to list files
					dirListCmd := fmt.Sprintf("ls -la %s 2>/dev/null | head -5", service.ConfigFile)
					if dirContents, _ := sc.executeCommand(dirListCmd); dirContents != "" {
						// Keep directory path as it contains files
					} else {
						service.ConfigFile = "Not found"
					}
				} else {
					service.ConfigFile = "Not found"
				}
			}
		}

		// Handle install path
		if len(parts) >= 3 && parts[2] != "" {
			service.InstallPath = parts[2]

			// Verify the installation path exists
			if sc.isSSH {
				installVerifyCmd := fmt.Sprintf("ls -la %s 2>/dev/null || echo 'Not found'", service.InstallPath)
				if output, _ := sc.executeCommand(installVerifyCmd); strings.Contains(output, "Not found") {
					// Try data directory as fallback
					dataDirCmd := "mysqld --verbose --help 2>/dev/null | grep 'datadir' | awk '{print $2}' | head -1"
					if dataDir, _ := sc.executeCommand(dataDirCmd); dataDir != "" {
						service.InstallPath = strings.TrimSpace(dataDir)
					}
				}
			}
		}
	}

	// Get version with multiple fallback methods
	versionCmd := `
# Method 1: Query running server (most accurate)
if VERSION=$(mysql --connect-timeout=2 -N -B -e "SELECT VERSION()" 2>/dev/null); then
    echo "$VERSION"
    exit 0
fi

# Method 2: Try with sudo if applicable
if VERSION=$(sudo mysql --connect-timeout=2 -N -B -e "SELECT VERSION()" 2>/dev/null); then
    echo "$VERSION"
    exit 0
fi

# Method 3: Binary version from mysqld
if VERSION=$(mysqld --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); then
    echo "$VERSION"
    exit 0
fi

# Method 4: Binary version from mariadbd
if VERSION=$(mariadbd --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); then
    echo "$VERSION"
    exit 0
fi

# Method 5: Check package
if command -v dpkg >/dev/null 2>&1; then
    if VERSION=$(dpkg -l | grep -E 'mariadb-server|mysql-server' | awk '{print $3}' | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); then
        echo "$VERSION"
        exit 0
    fi
elif command -v rpm >/dev/null 2>&1; then
    if VERSION=$(rpm -qa | grep -E 'MariaDB-server|mysql-server' | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); then
        echo "$VERSION"
        exit 0
    fi
fi

echo "unknown"
`
	versionOutput, _ := sc.executeCommand(versionCmd)
	if versionOutput != "" && versionOutput != "unknown" {
		service.Version = strings.TrimSpace(versionOutput)
	}

	// Final check: if we have a binary path but no version, try to extract version from binary
	if service.Version == "" && service.InstallPath != "" {
		binaryVersionCmd := service.InstallPath + " --version | grep -oE '[0-9]+\\.[0-9]+\\.[0-9]+' | head -1"
		binaryVersionOutput, _ := sc.executeCommand(binaryVersionCmd)
		if binaryVersionOutput != "" {
			service.Version = strings.TrimSpace(binaryVersionOutput)
		}
	}

	return service, nil
}

// executeCommandContext runs command with context support for timeout
func (sc *ServiceChecker) executeCommandContext(ctx context.Context, cmdStr string) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// checkMaxScale checks for MaxScale service
func (sc *ServiceChecker) checkMaxScale(serverID int, now time.Time) (*ServiceInfo, error) {
	service := &ServiceInfo{
		ServerID:    serverID,
		ServiceName: "MaxScale",
		ServiceType: "proxy",
		LastChecked: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Port:        4006, // Set default port
	}

	// Check if MaxScale is installed and get version
	versionCmd := "maxscale --version 2>/dev/null | grep -o '[0-9]\\+\\.[0-9]\\+\\.[0-9]\\+' | head -1"
	versionOutput, err := sc.executeCommand(versionCmd)
	if err != nil || versionOutput == "" {
		return nil, fmt.Errorf("MaxScale not installed")
	}
	service.Version = versionOutput

	// Check status, autostart, path, and port in a single command
	infoCmd := `
    STATUS=$(systemctl show -p ActiveState,UnitFileState --value maxscale 2>/dev/null)
    PORT=$(ss -tulpn 2>/dev/null | grep maxscale | grep LISTEN | awk '{print $5}' | awk -F: '{print $NF}' | head -1)
    CONFIG=$(find /etc -name maxscale.cnf 2>/dev/null | head -1)
    BINARY=$(which maxscale 2>/dev/null)
    echo "$STATUS:$PORT:$CONFIG:$BINARY"
    `
	infoOutput, _ := sc.executeCommand(infoCmd)

	parts := strings.Split(infoOutput, ":")

	// Parse status and autostart
	if len(parts) > 0 {
		statusParts := strings.Split(parts[0], "\n")
		if len(statusParts) >= 1 && statusParts[0] == "active" {
			service.Status = "running"
		} else {
			service.Status = "stopped"
		}

		if len(statusParts) >= 2 && statusParts[1] == "enabled" {
			service.AutoStart = true
		}
	}

	// Parse port
	if len(parts) > 1 && parts[1] != "" {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			service.Port = port
		}
	}

	// Parse config file
	if len(parts) > 2 && parts[2] != "" {
		service.ConfigFile = parts[2]
	}

	// Parse install path
	if len(parts) > 3 && parts[3] != "" {
		service.InstallPath = parts[3]
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
