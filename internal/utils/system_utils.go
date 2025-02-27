package utils

import (
	"bufio"
	"bytes"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

type OSInfo struct {
	Name         string
	Version      string
	Distribution string
}

// CheckSystemRequirements verifies if the system meets the required specifications
func CheckSystemRequirements() error {
	// Check if running on Linux
	if runtime.GOOS != "linux" {
		return fmt.Errorf("this application only runs on Linux systems, current OS: %s", runtime.GOOS)
	}

	// Get OS information
	osInfo, err := GetOSInfo()
	if err != nil {
		return fmt.Errorf("failed to get OS information: %w", err)
	}

	// Check for supported distributions
	switch strings.ToLower(osInfo.Distribution) {
	case "ubuntu":
		// Ubuntu is supported
		return nil
	case "centos", "red hat enterprise linux", "rocky linux":
		// RHEL-based systems are supported
		return nil
	default:
		return fmt.Errorf("unsupported Linux distribution: %s. Only Ubuntu and RHEL-based systems (CentOS, RHEL, Rocky) are supported", osInfo.Distribution)
	}
}

// GetOSInfo retrieves detailed information about the operating system
func GetOSInfo() (*OSInfo, error) {
	// Read /etc/os-release file
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("failed to read OS information: %w", err)
	}
	defer file.Close()

	info := &OSInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"")

		switch key {
		case "NAME":
			info.Name = value
		case "VERSION_ID":
			info.Version = value
		case "ID":
			info.Distribution = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading OS information: %w", err)
	}

	// Additional check for RHEL-based systems using /etc/redhat-release
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		content, err := os.ReadFile("/etc/redhat-release")
		if err == nil {
			rhel := string(content)
			if strings.Contains(strings.ToLower(rhel), "red hat") {
				info.Distribution = "red hat enterprise linux"
			} else if strings.Contains(strings.ToLower(rhel), "rocky") {
				info.Distribution = "rocky linux"
			}
		}
	}

	return info, nil
}

// GetSystemInfo returns detailed system information
func GetSystemInfo() (*OSInfo, error) {
	info, err := GetOSInfo()
	if err != nil {
		return nil, err
	}

	// Additional system information can be added here if needed
	return info, nil
}

// GetSystemInfoNew retrieves system information either locally or via SSH
func GetSystemInfoNew(isSSH bool, host string, port int, user, password string) (*SystemInfo, error) {
	// Create spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ System Info ] "
	if isSSH {
		spin.Suffix = fmt.Sprintf(" Retrieving system information from %s@%s:%d...", user, host, port)
	} else {
		spin.Suffix = " Retrieving local system information..."
	}
	spin.Color("cyan")
	spin.Start()
	defer spin.Stop()

	// Commands to retrieve system information
	commands := []string{
		"hostname", // Hostname
		"grep MemTotal /proc/meminfo | awk '{print $2}'", // Total RAM in KB
		"nproc", // CPU cores
		"grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs", // CPU Model
		"uname -s", // OS Type
		"cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"'", // OS Version
	}

	var results []string

	if isSSH {
		// Get SSH client
		client, err := ssh.GetSSHClient(host, port, user, password)
		if err != nil {
			spin.FinalMSG = fmt.Sprintf("Failed to connect to %s@%s:%d\n", user, host, port)
			return nil, err
		}
		defer client.Close()

		// Execute all commands through the single SSH connection
		for _, cmd := range commands {
			session, err := client.NewSession()
			if err != nil {
				return nil, fmt.Errorf("failed to create SSH session: %w", err)
			}

			var out bytes.Buffer
			session.Stdout = &out
			err = session.Run(cmd)
			session.Close()

			if err != nil {
				results = append(results, "N/A")
			} else {
				results = append(results, strings.TrimSpace(out.String()))
			}
		}
	} else {
		// Execute commands locally
		for _, cmd := range commands {
			// For local execution, we need to use shell interpretation
			cmdParts := []string{"/bin/sh", "-c", cmd}
			output, err := exec.Command(cmdParts[0], cmdParts[1:]...).Output()

			if err != nil {
				results = append(results, "N/A")
			} else {
				results = append(results, strings.TrimSpace(string(output)))
			}
		}
	}

	// Format RAM to include units
	ramKB := results[1]
	if ramKB != "N/A" {
		if kb, err := strconv.ParseInt(ramKB, 10, 64); err == nil {
			mb := kb / 1024
			gb := float64(mb) / 1024
			if gb >= 1 {
				results[1] = fmt.Sprintf("%.2f GB", gb)
			} else {
				results[1] = fmt.Sprintf("%d MB", mb)
			}
		}
	}

	// Create SystemInfo struct
	sysInfo := &SystemInfo{
		Hostname:  results[0],
		TotalRAM:  results[1],
		CPUCore:   results[2],
		CPUModel:  results[3],
		OSType:    results[4],
		OSVersion: results[5],
	}

	time.Sleep(200 * time.Millisecond)

	return sysInfo, nil
}
