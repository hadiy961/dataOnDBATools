// File: internal/package/setup/lokal_rpm/transfer.go
package lokal_rpm

import (
	"dbaTools/internal/logger"
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"path/filepath"
)

type TransferConfig struct {
	User      string
	IPAddress string
	Port      string
	DestDir   string
}

// PromptForTransfer menangani seluruh proses transfer
func PromptForTransfer(log *logger.Logger) error {
	// Prompt for transfer
	transferWanted := utils.PromptYesNo("Apakah Anda ingin mentransfer package ke server lain?", false)
	if !transferWanted {
		log.Success("Program selesai. Package tersimpan di direktori mariadb_packages")
		return nil
	}

	// Get transfer configuration
	config := promptTransferConfig()

	// Display summary
	displayTransferSummary(config)

	// Final confirmation
	confirm := utils.PromptYesNo("Apakah Anda ingin melanjutkan transfer?", false)
	if !confirm {
		log.Warning("Transfer dibatalkan. Program selesai.")

		return nil
	}

	log.Info("Memulai transfer ke " + config.User + "@" + config.IPAddress)

	// Process the transfer
	if err := ProcessTransfer(config); err != nil {
		log.Error("Error during transfer", err)

		// Ask if user wants to retry
		if utils.PromptYesNo("Apakah Anda ingin mencoba transfer ulang?", true) {
			log.Info("Mencoba Transfer ulang....")
			return PromptForTransfer(log)
		}
	}
	log.Success("Transfer Selesai !!")

	return nil
}

// displayTransferSummary menampilkan ringkasan transfer
func displayTransferSummary(config TransferConfig) {
	fmt.Println("\nRingkasan Transfer:")
	fmt.Println("==================")
	fmt.Printf("Server Tujuan    : %s@%s\n", config.User, config.IPAddress)
	fmt.Printf("Port SSH         : %s\n", config.Port)
	fmt.Printf("Direktori Tujuan : %s\n", config.DestDir)

	// Get package directory size
	size, err := getDirSize("mariadb_packages")
	if err == nil {
		fmt.Printf("Ukuran Package   : %.2f MB\n", float64(size)/(1024*1024))
	}
	fmt.Println("==================")
}

// getDirSize menghitung ukuran direktori
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
