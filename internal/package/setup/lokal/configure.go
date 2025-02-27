package lokal

import (
	"bytes"
	"context"
	"dbaTools/internal/logger"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// configureMariaDB handles post-installation configuration with enhanced error handling
func configureMariaDB(log *logger.Logger) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Function to execute commands with timeout
	execWithTimeout := func(name string, args ...string) error {
		cmd := exec.CommandContext(ctx, name, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%v: %s", err, stderr.String())
		}
		return nil
	}

	// Enable MariaDB service
	log.Info("Mengaktifkan service MariaDB")
	if err := execWithTimeout("systemctl", "enable", "mariadb"); err != nil {
		return &InstallError{
			Stage:   "Service Enable",
			Message: "Gagal mengaktifkan service MariaDB",
			Err:     err,
		}
	}

	// Start MariaDB service
	if err := execWithTimeout("systemctl", "start", "mariadb"); err != nil {
		return &InstallError{
			Stage:   "Service Start",
			Message: "Gagal menjalankan service MariaDB",
			Err:     err,
		}
	}

	// Wait for service to be fully started
	for i := 0; i < 30; i++ {
		cmd := exec.CommandContext(ctx, "systemctl", "is-active", "mariadb")
		output, err := cmd.Output()

		if err == nil && strings.TrimSpace(string(output)) == "active" {
			log.Success("MariaDB service berhasil diaktifkan")
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return &InstallError{
		Stage:   "Service Verification",
		Message: "Timeout menunggu MariaDB service aktif",
		Err:     nil,
	}
}
