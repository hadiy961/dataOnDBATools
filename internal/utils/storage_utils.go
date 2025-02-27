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

// GetServerStorage retrieves storage information either locally or via SSH
func GetServerStorage(isSSH bool, host string, port int, user, password string) ([]StorageInfo, error) {
	// Create spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Storage Info ] "
	if isSSH {
		spin.Suffix = fmt.Sprintf(" Retrieving storage information from %s@%s:%d...", user, host, port)
	} else {
		spin.Suffix = " Retrieving local storage information..."
	}
	spin.Color("cyan")
	spin.Start()
	defer spin.Stop()

	// Command untuk mendapatkan informasi storage (hanya df)
	command := "df -hT | grep -v tmpfs | grep -v devtmpfs | grep -v squashfs | grep -v overlay | tail -n +2"

	var result string

	if isSSH {
		// Get SSH client
		client, err := ssh.GetSSHClient(host, port, user, password)
		if err != nil {
			spin.FinalMSG = fmt.Sprintf("Failed to connect to %s@%s:%d\n", user, host, port)
			return nil, err
		}
		defer client.Close()

		// Execute command through SSH
		session, err := client.NewSession()
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH session: %w", err)
		}

		var out bytes.Buffer
		session.Stdout = &out
		err = session.Run(command)
		session.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to execute command on remote host: %w", err)
		}

		result = strings.TrimSpace(out.String())
	} else {
		// Execute command locally
		cmdParts := []string{"/bin/sh", "-c", command}
		output, err := exec.Command(cmdParts[0], cmdParts[1:]...).Output()

		if err != nil {
			return nil, fmt.Errorf("failed to execute command locally: %w", err)
		}

		result = strings.TrimSpace(string(output))
	}

	// Parse df output to get filesystem information
	var storageInfoList []StorageInfo

	dfLines := strings.Split(result, "\n")
	for _, line := range dfLines {
		fields := strings.Fields(line)
		if len(fields) >= 7 {
			// Standard df -hT output format: Filesystem Type Size Used Avail Use% Mounted
			usePercent := 0
			// Remove % character from the percentage field
			percentStr := strings.TrimSuffix(fields[5], "%")
			usePercent, _ = strconv.Atoi(percentStr)

			storageInfo := StorageInfo{
				DeviceName:     fields[0],
				FilesystemType: fields[1],
				TotalSpace:     fields[2],
				UsedSpace:      fields[3],
				FreeSpace:      fields[4],
				UsePercent:     usePercent,
				MountPoint:     fields[6],
			}

			storageInfoList = append(storageInfoList, storageInfo)
		}
	}

	time.Sleep(101 * time.Millisecond)

	return storageInfoList, nil
}
