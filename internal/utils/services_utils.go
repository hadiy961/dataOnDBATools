package utils

import (
	"context"
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
// func (sc *ServiceChecker) executeCommand(command string) (string, error) {
// 	if sc.isSSH {
// 		// For SSH execution, use SSH client
// 		client, err := ssh.GetSSHClient(sc.host, sc.port, sc.user, sc.password)
// 		if err != nil {
// 			return "", fmt.Errorf("SSH connection error: %w", err)
// 		}
// 		defer client.Close()

// 		session, err := client.NewSession()
// 		if err != nil {
// 			return "", fmt.Errorf("failed to create SSH session: %w", err)
// 		}
// 		defer session.Close()

// 		var out bytes.Buffer
// 		session.Stdout = &out
// 		err = session.Run(command)
// 		if err != nil {
// 			return "", fmt.Errorf("command execution failed: %w", err)
// 		}

// 		return strings.TrimSpace(out.String()), nil
// 	} else {
// 		// For local execution
// 		cmdParts := []string{"/bin/sh", "-c", command}
// 		output, err := exec.Command(cmdParts[0], cmdParts[1:]...).Output()
// 		if err != nil {
// 			return "", fmt.Errorf("command execution failed: %w", err)
// 		}

// 		return strings.TrimSpace(string(output)), nil
// 	}
// }

// checkMariaDB checks for MariaDB service
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

	// Gunakan context dengan timeout untuk semua command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Periksa status dan autostart dalam satu command
	serviceStatusCmd := "systemctl show -p ActiveState,UnitFileState --value mariadb 2>/dev/null || systemctl show -p ActiveState,UnitFileState --value mysql 2>/dev/null"
	statusOutput, err := sc.executeCommandContext(ctx, serviceStatusCmd)
	if err == nil && statusOutput != "" {
		lines := strings.Split(strings.TrimSpace(statusOutput), "\n")
		if len(lines) >= 1 && lines[0] == "active" {
			service.Status = "running"
		} else {
			service.Status = "stopped"
		}

		if len(lines) >= 2 && lines[1] == "enabled" {
			service.AutoStart = true
		}
	} else {
		service.Status = "stopped"
	}

	// Periksa port dan path dalam satu command yang lebih efisien
	infoCmd := `
	PORT=$(ss -tulpn 2>/dev/null | awk '/mysql|mariadb/ && /LISTEN/ {print $5}' | awk -F: '{print $NF}' | head -1)
	CONFIG=$(find /etc/my.cnf /etc/mysql/my.cnf /etc/mysql/mariadb.cnf /etc/my.cnf.d/server.cnf /etc/mysql/mariadb/50-server.cnf /etc/mysql/mariadb.conf.d/50-server.cnf /etc/mysql/conf.d/mysql.cnf /etc/opt/rh/rh-mariadb*/my.cnf -type f 2>/dev/null | grep -v '.dpkg\|.rpm' | head -1)
	BINARY=$(which mysqld 2>/dev/null || which mariadbd 2>/dev/null)
	echo "$PORT:$CONFIG:$BINARY"
	`
	infoOutput, _ := sc.executeCommandContext(ctx, infoCmd)

	if infoOutput != "" {
		parts := strings.Split(infoOutput, ":")
		if len(parts) >= 1 && parts[0] != "" {
			if port, err := strconv.Atoi(parts[0]); err == nil {
				service.Port = port
			}
		}
		if len(parts) >= 2 && parts[1] != "" {
			service.ConfigFile = parts[1]
		}
		if len(parts) >= 3 && parts[2] != "" {
			service.InstallPath = parts[2]
		}
	}

	// Periksa versi dengan satu query jika service berjalan
	if service.Status == "running" {
		// Coba cara paling sederhana dulu
		versionCmd := `mysql --connect-timeout=2 -N -B -e "SELECT VERSION()" 2>/dev/null || sudo mysql --connect-timeout=2 -N -B -e "SELECT VERSION()" 2>/dev/null`
		versionOutput, _ := sc.executeCommandContext(ctx, versionCmd)

		if versionOutput != "" {
			service.Version = strings.TrimSpace(versionOutput)
		} else {
			// Fallback ke versi binary jika gagal
			versionBinCmd := `(mysqld --version 2>/dev/null || mariadbd --version 2>/dev/null) | grep -o "[0-9]\+\.[0-9]\+\.[0-9]\+"`
			versionOutput, _ := sc.executeCommandContext(ctx, versionBinCmd)
			if versionOutput != "" {
				service.Version = strings.TrimSpace(versionOutput)
			}
		}
	} else {
		// Service tidak berjalan, cek versi dari binary
		versionBinCmd := `(mysqld --version 2>/dev/null || mariadbd --version 2>/dev/null) | grep -o "[0-9]\+\.[0-9]\+\.[0-9]\+"`
		versionOutput, _ := sc.executeCommandContext(ctx, versionBinCmd)
		if versionOutput != "" {
			service.Version = strings.TrimSpace(versionOutput)
		}
	}

	// Jika tidak ada versi yang ditemukan, mungkin MariaDB tidak terinstall
	if service.Version == "" && service.Status == "stopped" && service.InstallPath == "" {
		return nil, fmt.Errorf("MariaDB not installed")
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

	// Gunakan context dengan timeout untuk semua command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Periksa apakah MaxScale terinstall sekaligus dapatkan versinya
	versionCmd := "maxscale --version 2>/dev/null | grep -o '[0-9]\\+\\.[0-9]\\+\\.[0-9]\\+' | head -1"
	versionOutput, err := sc.executeCommandContext(ctx, versionCmd)
	if err != nil || versionOutput == "" {
		return nil, fmt.Errorf("MaxScale not installed")
	}
	service.Version = versionOutput

	// Periksa status, autostart, path, dan port dalam satu command
	infoCmd := `
	STATUS=$(systemctl show -p ActiveState,UnitFileState --value maxscale 2>/dev/null)
	PORT=$(ss -tulpn 2>/dev/null | grep maxscale | grep LISTEN | awk '{print $5}' | awk -F: '{print $NF}' | head -1)
	CONFIG=$(find /etc -name maxscale.cnf 2>/dev/null | head -1)
	BINARY=$(which maxscale 2>/dev/null)
	echo "$STATUS:$PORT:$CONFIG:$BINARY"
	`
	infoOutput, _ := sc.executeCommandContext(ctx, infoCmd)

	parts := strings.Split(infoOutput, ":")

	// Parse status dan autostart
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

// checkNFSUtils checks for NFS utils
// func (sc *ServiceChecker) checkNFSUtils(serverID int, now time.Time) (*ServiceInfo, error) {
// 	service := &ServiceInfo{
// 		ServerID:    serverID,
// 		ServiceName: "NFS Utils",
// 		ServiceType: "filesystem",
// 		LastChecked: now,
// 		CreatedAt:   now,
// 		UpdatedAt:   now,
// 	}

// 	// Check if package is installed
// 	checkCmd := "rpm -q nfs-utils 2>/dev/null || dpkg -l | grep nfs-common 2>/dev/null"
// 	checkOutput, err := sc.executeCommand(checkCmd)
// 	if err != nil || checkOutput == "" {
// 		return nil, fmt.Errorf("NFS Utils not installed")
// 	}

// 	// Extract version
// 	versionParts := strings.Fields(checkOutput)
// 	if len(versionParts) > 0 {
// 		for _, part := range versionParts {
// 			if strings.Contains(part, "-") {
// 				parts := strings.Split(part, "-")
// 				if len(parts) >= 2 {
// 					service.Version = parts[1]
// 					break
// 				}
// 			}
// 		}
// 	}

// 	// Check status of NFS service
// 	statusCmd := "systemctl is-active nfs-server 2>/dev/null || systemctl is-active nfs-kernel-server 2>/dev/null"
// 	statusOutput, err := sc.executeCommand(statusCmd)
// 	if err == nil && statusOutput == "active" {
// 		service.Status = "running"
// 	} else {
// 		service.Status = "installed"
// 	}

// 	// Check config file
// 	configCmd := "echo /etc/exports"
// 	configOutput, err := sc.executeCommand(configCmd)
// 	if err == nil && configOutput != "" {
// 		service.ConfigFile = configOutput
// 	}

// 	// Check install path
// 	installPathCmd := "which showmount 2>/dev/null || echo /usr/sbin/showmount"
// 	installPathOutput, err := sc.executeCommand(installPathCmd)
// 	if err == nil && installPathOutput != "" {
// 		service.InstallPath = installPathOutput
// 	}

// 	// Check auto start
// 	autoStartCmd := "systemctl is-enabled nfs-server 2>/dev/null || systemctl is-enabled nfs-kernel-server 2>/dev/null"
// 	autoStartOutput, err := sc.executeCommand(autoStartCmd)
// 	if err == nil && autoStartOutput == "enabled" {
// 		service.AutoStart = true
// 	} else {
// 		service.AutoStart = false
// 	}

// 	return service, nil
// }

// checkSSHPass checks for sshpass
// func (sc *ServiceChecker) checkSSHPass(serverID int, now time.Time) (*ServiceInfo, error) {
// 	service := &ServiceInfo{
// 		ServerID:    serverID,
// 		ServiceName: "SSHPass",
// 		ServiceType: "utility",
// 		LastChecked: now,
// 		CreatedAt:   now,
// 		UpdatedAt:   now,
// 	}

// 	// Check if command is available
// 	versionCmd := "sshpass -V 2>&1 | head -1"
// 	versionOutput, err := sc.executeCommand(versionCmd)
// 	if err != nil || versionOutput == "" {
// 		return nil, fmt.Errorf("SSHPass not installed")
// 	}

// 	// Extract version
// 	versionParts := strings.Fields(versionOutput)
// 	if len(versionParts) > 0 {
// 		for _, part := range versionParts {
// 			if strings.Contains(part, ".") {
// 				service.Version = part
// 				break
// 			}
// 		}
// 	}

// 	// No service status for sshpass as it's a utility, not a daemon
// 	service.Status = "installed"

// 	// No port for sshpass
// 	service.Port = 0

// 	// No config file for sshpass
// 	service.ConfigFile = "N/A"

// 	// Check install path
// 	installPathCmd := "which sshpass 2>/dev/null"
// 	installPathOutput, err := sc.executeCommand(installPathCmd)
// 	if err == nil && installPathOutput != "" {
// 		service.InstallPath = installPathOutput
// 	}

// 	// No auto start for sshpass
// 	service.AutoStart = false

// 	return service, nil
// }

// checkNetTools checks for net-tools
// func (sc *ServiceChecker) checkNetTools(serverID int, now time.Time) (*ServiceInfo, error) {
// 	service := &ServiceInfo{
// 		ServerID:    serverID,
// 		ServiceName: "Net Tools",
// 		ServiceType: "utility",
// 		LastChecked: now,
// 		CreatedAt:   now,
// 		UpdatedAt:   now,
// 	}

// 	// Check if package is installed
// 	checkCmd := "rpm -q net-tools 2>/dev/null || dpkg -l | grep net-tools 2>/dev/null"
// 	checkOutput, err := sc.executeCommand(checkCmd)
// 	if err != nil || checkOutput == "" {
// 		return nil, fmt.Errorf("Net Tools not installed")
// 	}

// 	// Extract version
// 	versionParts := strings.Fields(checkOutput)
// 	if len(versionParts) > 0 {
// 		for _, part := range versionParts {
// 			if strings.Contains(part, "-") || strings.Contains(part, ".") {
// 				service.Version = part
// 				break
// 			}
// 		}
// 	}

// 	// No service status for net-tools as it's a utility, not a daemon
// 	service.Status = "installed"

// 	// No port for net-tools
// 	service.Port = 0

// 	// No config file for net-tools
// 	service.ConfigFile = "N/A"

// 	// Check install path for a common nettools command (netstat)
// 	installPathCmd := "which netstat 2>/dev/null"
// 	installPathOutput, err := sc.executeCommand(installPathCmd)
// 	if err == nil && installPathOutput != "" {
// 		service.InstallPath = installPathOutput
// 	}

// 	// No auto start for net-tools
// 	service.AutoStart = false

// 	return service, nil
// }

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
