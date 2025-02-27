package lokal

import (
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"strings"
)

// PromptPackageLocation prompts for and validates package location based on system distribution
func PromptPackageLocation(defaultPath string) (string, error) {
	// Get OS info to determine package type
	osInfo, err := utils.GetOSInfo()
	if err != nil {
		return "", fmt.Errorf("gagal mendeteksi sistem operasi: %w", err)
	}

	// Determine expected package extension
	var pkgExt string
	switch osInfo.Distribution {
	case "ubuntu":
		pkgExt = ".deb"
	case "centos", "red hat enterprise linux", "rocky linux":
		pkgExt = ".rpm"
	default:
		return "", fmt.Errorf("distribusi tidak didukung: %s", osInfo.Distribution)
	}

	for {
		// Prompt for directory path
		path := utils.PromptString(fmt.Sprintf("Masukkan lokasi direktori package %s", strings.ToUpper(pkgExt[1:])), defaultPath)

		// Validate directory exists
		exists, err := utils.DirectoryExists(path)
		if err != nil {
			return "", fmt.Errorf("gagal memeriksa direktori: %w", err)
		}

		if !exists {
			fmt.Println("Direktori tidak ditemukan. Silakan masukkan path yang valid.")
			continue
		}

		// Validate directory is readable
		if err := utils.IsValidDirectory(path); err != nil {
			fmt.Println("Direktori tidak dapat diakses:", err)
			continue
		}

		// Check for package files
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("gagal membaca direktori: %w", err)
		}

		var pkgFiles []string
		var totalSize int64

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filename := strings.ToLower(entry.Name())
			if strings.HasSuffix(filename, pkgExt) {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				pkgFiles = append(pkgFiles, filename)
				totalSize += info.Size()
			}
		}

		// Display package information
		if len(pkgFiles) > 0 {
			fmt.Printf("\nDitemukan %d file%s:\n", len(pkgFiles), pkgExt)
			for _, file := range pkgFiles {
				fmt.Printf("- %s\n", file)
			}
			fmt.Printf("Total ukuran: %.2f MB\n\n", float64(totalSize)/(1024*1024))
		} else {
			fmt.Printf("\nPeringatan: Tidak ada file%s yang ditemukan di direktori ini.\n", pkgExt)
			if !utils.PromptYesNo("Tetap gunakan direktori ini?", false) {
				continue
			}
		}

		// Validate MariaDB packages
		hasMariaDB := false
		for _, file := range pkgFiles {
			if strings.Contains(strings.ToLower(file), "mariadb") {
				hasMariaDB = true
				break
			}
		}

		if !hasMariaDB && len(pkgFiles) > 0 {
			fmt.Println("\nPeringatan: Tidak ditemukan package MariaDB.")
			if !utils.PromptYesNo("Lanjutkan dengan direktori ini?", false) {
				continue
			}
		}

		// Confirm selection
		fmt.Printf("\nLokasi package: %s\n", path)
		if utils.PromptYesNo("Konfirmasi lokasi ini?", true) {
			return path, nil
		}
	}
}
