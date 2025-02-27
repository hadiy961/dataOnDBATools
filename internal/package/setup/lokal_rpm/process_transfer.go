package lokal_rpm

import (
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProcessTransfer handles the file transfer using sshpass and scp
func ProcessTransfer(config TransferConfig) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if sshpass is installed
	if err := checkSSHPass(); err != nil {
		return err
	}

	// Get source directory path
	sourceDir := filepath.Join(currentDir, "mariadb_packages")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory not found: %s", sourceDir)
	}

	// Get password securely
	password := utils.PromptPassword("Masukkan password SSH")
	if password == "" {
		return fmt.Errorf("password tidak boleh kosong")
	}

	// Create remote directory if it doesn't exist
	if err := createRemoteDirectory(config, password); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Start transfer
	fmt.Printf("\nMemulai transfer file ke %s@%s:%s\n", config.User, config.IPAddress, config.DestDir)

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	totalFiles := len(entries)
	for i, entry := range entries {
		srcPath := filepath.Join(sourceDir, entry.Name())
		fmt.Printf("Transferring [%d/%d]: %s\n", i+1, totalFiles, entry.Name())

		if err := transferFile(srcPath, config, password); err != nil {
			return fmt.Errorf("failed to transfer %s: %w", entry.Name(), err)
		}
	}

	fmt.Printf("\nTransfer selesai! %d file berhasil ditransfer ke %s@%s:%s\n",
		totalFiles, config.User, config.IPAddress, config.DestDir)

	return nil
}

// checkSSHPass verifies if sshpass is installed
func checkSSHPass() error {
	cmd := exec.Command("which", "sshpass")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sshpass tidak ditemukan. Silakan install dengan menjalankan: sudo apt-get install sshpass")
	}
	return nil
}

// createRemoteDirectory ensures the target directory exists on remote server
func createRemoteDirectory(config TransferConfig, password string) error {
	sshCmd := fmt.Sprintf("mkdir -p %s", config.DestDir)
	cmd := exec.Command("sshpass", "-p", password, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-p", config.Port,
		fmt.Sprintf("%s@%s", config.User, config.IPAddress),
		sshCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}
	return nil
}

// transferFile transfers a single file using scp
func transferFile(srcPath string, config TransferConfig, password string) error {
	cmd := exec.Command("sshpass", "-p", password, "scp",
		"-o", "StrictHostKeyChecking=no",
		"-P", config.Port,
		srcPath,
		fmt.Sprintf("%s@%s:%s", config.User, config.IPAddress, config.DestDir))

	// Capture error output
	errOutput := &strings.Builder{}
	cmd.Stderr = errOutput

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("transfer failed: %s", errOutput.String())
	}
	return nil
}
