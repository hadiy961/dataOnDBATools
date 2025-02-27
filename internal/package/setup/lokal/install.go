package lokal

import (
	"bytes"
	"dbaTools/internal/logger"
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// InstallPackages handles package installation from a specified directory
func InstallPackages(path string, log *logger.Logger) error {
	// Validate inputs
	if path == "" {
		return &InstallError{
			Stage:   "Validation",
			Message: "Path tidak boleh kosong",
			Err:     nil,
		}
	}

	if log == nil {
		return &InstallError{
			Stage:   "Validation",
			Message: "Logger tidak boleh nil",
			Err:     nil,
		}
	}

	// Get OS info for package type
	osInfo, err := utils.GetOSInfo()
	if err != nil {
		return &InstallError{
			Stage:   "OS Detection",
			Message: "Gagal mendeteksi sistem operasi",
			Err:     err,
		}
	}

	// Validate root/sudo access
	if os.Geteuid() != 0 {
		return &InstallError{
			Stage:   "Permission",
			Message: "Instalasi membutuhkan akses root/sudo",
			Err:     nil,
		}
	}

	// Validate directory exists and is accessible
	if exists, err := utils.DirectoryExists(path); !exists || err != nil {
		return &InstallError{
			Stage:   "Directory Check",
			Message: "Direktori tidak ditemukan atau tidak dapat diakses",
			Err:     err,
		}
	}

	// Read directory contents with error handling
	entries, err := os.ReadDir(path)
	if err != nil {
		return &InstallError{
			Stage:   "Directory Read",
			Message: "Gagal membaca direktori",
			Err:     err,
		}
	}

	var pkgFiles []string
	pkgExt := ".rpm"
	var installCmd string
	var installArgs []string

	// Set installation parameters based on OS
	switch osInfo.Distribution {
	case "ubuntu":
		pkgExt = ".deb"
		// Check if dpkg is available
		if _, err := exec.LookPath("dpkg"); err != nil {
			return &InstallError{
				Stage:   "Dependency Check",
				Message: "dpkg tidak ditemukan",
				Err:     err,
			}
		}
		installCmd = "dpkg"
		installArgs = []string{"-i"}
	case "centos", "red hat enterprise linux", "rocky linux":
		pkgExt = ".rpm"
		// Check if rpm is available
		if _, err := exec.LookPath("rpm"); err != nil {
			return &InstallError{
				Stage:   "Dependency Check",
				Message: "rpm tidak ditemukan",
				Err:     err,
			}
		}
		installCmd = "rpm"
		installArgs = []string{"-Uvh", "--nodeps"}
	default:
		return &InstallError{
			Stage:   "OS Support",
			Message: fmt.Sprintf("Distribusi tidak didukung: %s", osInfo.Distribution),
			Err:     nil,
		}
	}

	// Collect and validate package files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(strings.ToLower(filename), pkgExt) {
			// Validate file readability
			if !utils.IsReadable(filepath.Join(path, filename)) {
				log.Warning(fmt.Sprintf("File tidak dapat dibaca: %s", filename))
				continue
			}
			pkgFiles = append(pkgFiles, filename)
		}
	}

	if len(pkgFiles) == 0 {
		return &InstallError{
			Stage:   "Package Check",
			Message: fmt.Sprintf("Tidak ada package valid yang ditemukan di %s", path),
			Err:     nil,
		}
	}

	// Sort packages with dependency handling
	sort.Slice(pkgFiles, func(i, j int) bool {
		iName := strings.ToLower(pkgFiles[i])
		jName := strings.ToLower(pkgFiles[j])

		// Priority order with error tolerance
		if strings.Contains(iName, "common") || strings.Contains(iName, "shared") {
			return true
		}
		if strings.Contains(jName, "common") || strings.Contains(jName, "shared") {
			return false
		}
		if strings.Contains(iName, "server") {
			return false
		}
		if strings.Contains(jName, "server") {
			return true
		}
		return iName < jName
	})

	// Display installation plan
	fmt.Printf("\nRencana instalasi package:\n")
	for i, pkg := range pkgFiles {
		fmt.Printf("%d. %s\n", i+1, pkg)
	}

	if !utils.PromptYesNo("\nLanjutkan instalasi?", true) {
		return &InstallError{
			Stage:   "User Confirmation",
			Message: "Instalasi dibatalkan oleh pengguna",
			Err:     nil,
		}
	}

	// Track installation progress
	type InstallResult struct {
		Package string
		Success bool
		Error   error
	}

	var results []InstallResult
	var failedCount int

	// Install packages with detailed error handling
	for _, pkg := range pkgFiles {
		pkgPath := filepath.Join(path, pkg)
		log.Info(fmt.Sprintf("Menginstall: %s", pkg))

		cmd := exec.Command(installCmd, append(installArgs, pkgPath)...)
		var stderr bytes.Buffer
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		result := InstallResult{
			Package: pkg,
			Success: err == nil,
			Error:   err,
		}
		results = append(results, result)

		if err != nil {
			failedCount++
			log.Error(fmt.Sprintf("Gagal menginstall %s", pkg), fmt.Errorf("%v: %s", err, stderr.String()))

			if failedCount >= 3 {
				return &InstallError{
					Stage:   "Installation",
					Message: "Terlalu banyak kegagalan instalasi",
					Err:     fmt.Errorf("gagal menginstall 3 atau lebih package"),
				}
			}

			if !utils.PromptYesNo("Lanjutkan instalasi package berikutnya?", true) {
				return &InstallError{
					Stage:   "Installation",
					Message: "Instalasi dibatalkan setelah error",
					Err:     err,
				}
			}
			continue
		}

		log.Success(fmt.Sprintf("Berhasil menginstall %s", pkg))
	}

	// Generate installation report
	fmt.Printf("\nLaporan Instalasi:\n")
	fmt.Printf("Total package: %d\n", len(results))
	fmt.Printf("Berhasil: %d\n", len(results)-failedCount)
	fmt.Printf("Gagal: %d\n", failedCount)

	if failedCount > 0 {
		fmt.Printf("\nPackage yang gagal:\n")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("- %s: %v\n", result.Package, result.Error)
			}
		}
	}

	// Configure MariaDB with error handling
	if err := configureMariaDB(log); err != nil {
		return &InstallError{
			Stage:   "Configuration",
			Message: "Gagal mengkonfigurasi MariaDB",
			Err:     err,
		}
	}

	return nil
}
