package lokal_rpm

import (
	"dbaTools/internal/logger"
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"path/filepath"
)

func DownloadRPM(log *logger.Logger) error {
	// Step 1: Check OS
	log.Info("Checking system requirements...")
	osInfo, err := utils.GetOSInfo()
	if err != nil {
		log.Error("failed to get OS information", err)
	}

	// Display OS Summary
	fmt.Printf("\nSystem Information Summary:\n")
	fmt.Printf("========================\n")
	fmt.Printf("OS Name: %s\n", osInfo.Name)
	fmt.Printf("Version: %s\n", osInfo.Version)
	fmt.Printf("Distribution: %s\n", osInfo.Distribution)
	fmt.Printf("========================\n\n")

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("failed to get current directory", err)
	}

	// Create absolute path for packages directory
	packagesDir := filepath.Join(currentDir, "mariadb_packages")

	// Check if directory exists and is not empty
	if _, err := os.Stat(packagesDir); err == nil {
		entries, err := os.ReadDir(packagesDir)
		if err != nil {
			log.Error("failed to read directory", err)
		}

		if len(entries) > 0 {
			fmt.Printf("\nDitemukan direktori mariadb_packages yang tidak kosong.\n")
			fmt.Printf("Isi direktori:\n")
			fmt.Printf("==============\n")

			// Display directory contents with sizes
			var totalSize int64
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					fmt.Printf("- %s (error getting size)\n", entry.Name())
					continue
				}
				totalSize += info.Size()
				fmt.Printf("- %-60s %s\n", entry.Name(), formatSize(info.Size()))
			}
			fmt.Printf("\nTotal Size: %s\n", formatSize(totalSize))
			fmt.Printf("==============\n\n")

			// Prompt user for deletion
			if !utils.PromptYesNo("Apakah Anda ingin menghapus direktori ini dan melanjutkan?", true) {
				return PromptForTransfer(log)
			}
			log.Info("Membersihkan direktori " + packagesDir)

			if err := os.RemoveAll(packagesDir); err != nil {
				log.Error("failed to clean directory", err)

			}
		}
	}

	// Get appropriate downloader with absolute path
	downloader, err := getDownloader(packagesDir, osInfo, log)
	if err != nil {
		return err
	}

	// Execute download steps
	if err := downloader.PrepareDependencies(); err != nil {
		log.Error("failed to prepare dependencies", err)
	}

	if err := downloader.ConfigureRepository(); err != nil {
		log.Error("failed to configure repository", err)
	}

	if err := downloader.DownloadPackages(); err != nil {
		log.Error("failed to download packages", err)
	}

	// Generate and save package list
	if err := savePackageList(packagesDir); err != nil {
		log.Error("Failed to save package list", err)
	}
	log.Success("Packages have been downloaded to " + packagesDir)
	return nil
}
