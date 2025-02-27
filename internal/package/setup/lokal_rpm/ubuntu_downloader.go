package lokal_rpm

import (
	"dbaTools/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// UbuntuDownloader implementasi untuk Ubuntu
type UbuntuDownloader struct {
	*BaseDownloader
	logger *logger.Logger
}

// NewUbuntuDownloader membuat instance baru UbuntuDownloader
func NewUbuntuDownloader(baseDownloader *BaseDownloader, log *logger.Logger) *UbuntuDownloader {
	return &UbuntuDownloader{
		BaseDownloader: baseDownloader,
		logger:         log,
	}
}

// ensureAptPermissions ensures the directory exists and has correct permissions for _apt user
func (u *UbuntuDownloader) ensureAptPermissions() error {
	// First ensure the directory exists
	if err := os.MkdirAll(u.packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(u.packagesDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	u.packagesDir = absPath

	aptUser, err := user.Lookup("_apt")
	if err != nil {
		return fmt.Errorf("failed to lookup _apt user: %w", err)
	}

	uid, err := strconv.Atoi(aptUser.Uid)
	if err != nil {
		return fmt.Errorf("invalid uid for _apt user: %w", err)
	}

	gid, err := strconv.Atoi(aptUser.Gid)
	if err != nil {
		return fmt.Errorf("invalid gid for _apt user: %w", err)
	}

	// Set directory permissions for _apt user
	if err := os.Chown(u.packagesDir, uid, gid); err != nil {
		return fmt.Errorf("failed to change directory ownership: %w", err)
	}

	return nil
}

func (u *UbuntuDownloader) PrepareDependencies() error {
	if err := u.ensureAptPermissions(); err != nil {
		u.logger.Error("Gagal menyiapkan direktori", err)
		return fmt.Errorf("failed to prepare directory: %w", err)
	}

	// Add MariaDB 10.6 repository
	u.logger.Info("Menambahkan repository MariaDB 10.6...")

	repoContent := `
# MariaDB 10.6 repository
deb [arch=amd64] http://mirror.rackspace.com/mariadb/repo/10.6/ubuntu jammy main
deb-src http://mirror.rackspace.com/mariadb/repo/10.6/ubuntu jammy main`

	if err := os.WriteFile("/etc/apt/sources.list.d/mariadb.list", []byte(repoContent), 0644); err != nil {
		u.logger.Error("Gagal membuat file repository", err)
		return fmt.Errorf("failed to create repository file: %w", err)
	}

	// Import MariaDB GPG key
	keyCmd := exec.Command("sh", "-c", "curl -LsSO https://supplychain.mariadb.com/MariaDB-Server-GPG-KEY && sudo apt-key add MariaDB-Server-GPG-KEY")
	keyCmd.Stdout = os.Stdout
	keyCmd.Stderr = os.Stderr

	if err := keyCmd.Run(); err != nil {
		u.logger.Error("Gagal mengimpor GPG key", err)
		return fmt.Errorf("failed to import GPG key: %w", err)
	}

	// Update apt cache using sudo
	u.logger.Info("Memperbarui cache package...")
	updateCmd := exec.Command("sudo", "apt-get", "update")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		u.logger.Error("Gagal memperbarui cache package", err)
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	return nil
}

func (u *UbuntuDownloader) ConfigureRepository() error {
	return nil
}

func (u *UbuntuDownloader) runAptGet(args ...string) error {
	aptUser, err := user.Lookup("_apt")
	if err != nil {
		return fmt.Errorf("failed to lookup _apt user: %w", err)
	}

	uid, _ := strconv.Atoi(aptUser.Uid)
	gid, _ := strconv.Atoi(aptUser.Gid)

	cmd := exec.Command("apt-get", args...)
	cmd.Dir = u.packagesDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	return cmd.Run()
}

// getDependencies gets all dependencies for a package
func (u *UbuntuDownloader) getDependencies(pkg string) ([]string, error) {
	cmd := exec.Command("apt-cache", "depends", "--recurse", "--no-recommends", "--no-suggests", "--no-conflicts", "--no-breaks", "--no-replaces", "--no-enhances", pkg)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var deps []string
	seen := make(map[string]bool)

	// Mapping untuk package yang perlu diganti namanya
	packageMapping := map[string]string{
		"mysql-common": "mariadb-common",
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Depends:") {
			dep := strings.TrimPrefix(line, "Depends:")
			dep = strings.TrimSpace(dep)

			// Periksa apakah perlu mengganti nama package
			if mappedPkg, exists := packageMapping[dep]; exists {
				dep = mappedPkg
			}

			if !seen[dep] {
				// Cek ketersediaan package
				checkCmd := exec.Command("apt-cache", "show", dep)
				if err := checkCmd.Run(); err == nil {
					deps = append(deps, dep)
					seen[dep] = true
				} else {
					u.logger.Warning(fmt.Sprintf("Package %s tidak tersedia, mencoba alternatif", dep))
					// Jika package tidak ditemukan, coba cari alternatif
					if alt, exists := packageMapping[dep]; exists {
						deps = append(deps, alt)
						seen[alt] = true
					}
				}
			}
		}
	}

	return deps, nil
}

func (u *UbuntuDownloader) downloadPackageWithDeps(pkg string, downloaded map[string]bool) error {
	if downloaded[pkg] {
		return nil
	}

	u.logger.Info(fmt.Sprintf("Menganalisis dependensi untuk %s...", pkg))
	deps, err := u.getDependencies(pkg)
	if err != nil {
		u.logger.Warning(fmt.Sprintf("Gagal mendapatkan dependensi untuk %s: %v", pkg, err))
	}

	// Download dependencies first
	for _, dep := range deps {
		if !downloaded[dep] {
			u.logger.Info(fmt.Sprintf("Mengunduh dependensi: %s", dep))
			if err := u.runAptGet("download", dep); err != nil {
				u.logger.Warning(fmt.Sprintf("Gagal mengunduh dependensi %s: %v", dep, err))
				continue
			}
			downloaded[dep] = true
			u.logger.Success(fmt.Sprintf("Berhasil mengunduh dependensi: %s", dep))
		}
	}

	// Download the package itself
	u.logger.Info(fmt.Sprintf("Mengunduh package: %s", pkg))
	if err := u.runAptGet("download", pkg); err != nil {
		u.logger.Error(fmt.Sprintf("Gagal mengunduh %s", pkg), err)
		return fmt.Errorf("failed to download %s: %w", pkg, err)
	}
	downloaded[pkg] = true
	u.logger.Success(fmt.Sprintf("Berhasil mengunduh package: %s", pkg))

	return nil
}

func (u *UbuntuDownloader) fixFilePermissions() error {
	originalUid := os.Getuid()
	originalGid := os.Getgid()

	entries, err := os.ReadDir(u.packagesDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(u.packagesDir, entry.Name())
		if err := os.Chown(path, originalUid, originalGid); err != nil {
			u.logger.Warning(fmt.Sprintf("Gagal mengubah kepemilikan file %s: %v", entry.Name(), err))
		}
	}

	if err := os.Chown(u.packagesDir, originalUid, originalGid); err != nil {
		u.logger.Warning(fmt.Sprintf("Gagal mengembalikan kepemilikan direktori: %v", err))
	}

	return nil
}

func (u *UbuntuDownloader) DownloadPackages() error {
	downloaded := make(map[string]bool)

	// Define package groups in correct installation order
	packageGroups := [][]string{
		// 1. Common packages
		{"mariadb-common"},
		// 2. Core packages
		{"mariadb-client-core-10.6"},
		// 3. Client packages
		{"mariadb-client-10.6", "mariadb-client"},
		// 4. Server packages
		{"mariadb-server-10.6", "mariadb-server"},
		// 5. Backup packages
		{"mariadb-backup"},
		// 6. Libraries and dependencies
		{
			"libmariadb3",
			"libaio1",
			"libncurses6",
			"libtinfo6",
			"libssl3",
			"libstdc++6",
			"zlib1g",
		},
	}

	// Download each group in order
	groupNames := []string{
		"Common Packages",
		"Core Packages",
		"Client Packages",
		"Server Packages",
		"Backup Packages",
		"Libraries & Dependencies",
	}

	for groupIndex, group := range packageGroups {
		u.logger.Info(fmt.Sprintf("Memulai unduhan grup %s (%d/%d)",
			groupNames[groupIndex],
			groupIndex+1,
			len(packageGroups)))

		for _, pkg := range group {
			if err := u.downloadPackageWithDeps(pkg, downloaded); err != nil {
				u.logger.Warning(fmt.Sprintf("Gagal mengunduh %s dan dependensinya: %v", pkg, err))
				continue
			}
		}
		u.logger.Success(fmt.Sprintf("Selesai mengunduh grup %s (%d/%d)",
			groupNames[groupIndex],
			groupIndex+1,
			len(packageGroups)))
	}

	// Fix permissions after all downloads are complete
	u.logger.Info("Memperbaiki permission file...")
	if err := u.fixFilePermissions(); err != nil {
		u.logger.Warning(fmt.Sprintf("Gagal memperbaiki permission file: %v", err))
	}

	return nil
}
