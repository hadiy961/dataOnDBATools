package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func MoveDirectory(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func RemoveDirectory(path string, recursive bool) error {
	if recursive {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

func CopyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func GetDiskUsage(path string) (*syscall.Statfs_t, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	return &stat, err
}

// NewDirectoryConfig creates a new DirectoryConfig with default values
func NewDirectoryConfig(basePath string) *DirectoryConfig {
	return &DirectoryConfig{
		BasePath:      basePath,
		Mode:          0755,
		CreateParents: true,
		SkipIfExists:  true,
	}
}

func FindInDirectory(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}

func LockDirectory(path string) (*os.File, error) {
	lockFile := filepath.Join(path, ".lock")
	return os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0600)
}

func CreateDirectory(paths []string, config *DirectoryConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if len(paths) == 0 {
		return fmt.Errorf("no paths provided for directory creation")
	}

	for _, path := range paths {
		if path == "" {
			continue
		}

		// Sanitize and construct the full path
		cleanPath := filepath.Clean(path)
		fullPath := filepath.Join(config.BasePath, cleanPath)

		// Validate the path
		if err := validatePath(fullPath); err != nil {
			return fmt.Errorf("invalid path %s: %w", fullPath, err)
		}

		// Handle existing directory
		exists, err := directoryExists(fullPath)
		if err != nil {
			return fmt.Errorf("error checking directory %s: %w", fullPath, err)
		}

		if exists {
			if config.SkipIfExists {
				continue
			}
			return fmt.Errorf("directory already exists: %s", fullPath)
		}

		// Create the directory
		if err := createDir(fullPath, config); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	return nil
}

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path provided")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains invalid sequences")
	}

	// Additional path validation can be added here
	return nil
}

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func createDir(path string, config *DirectoryConfig) error {
	if config.CreateParents {
		return os.MkdirAll(path, config.Mode)
	}
	return os.Mkdir(path, config.Mode)
}

// CreateNestedDirectories creates nested directory structure from a string path
func CreateNestedDirectories(path string, config *DirectoryConfig) error {
	// Split path into segments
	segments := strings.Split(filepath.Clean(path), string(os.PathSeparator))
	return CreateDirectory(segments, config)
}

func GetDirectoryStats(path string) (*DirectoryStats, error) {
	stats := &DirectoryStats{}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			stats.TotalDirs++
		} else {
			stats.TotalFiles++
			stats.TotalSize += info.Size()
		}

		if info.ModTime().After(stats.LastModified) {
			stats.LastModified = info.ModTime()
		}

		depth := len(strings.Split(filePath, string(os.PathSeparator)))
		if depth > stats.MaxDepth {
			stats.MaxDepth = depth
		}

		return nil
	})

	return stats, err
}

// CheckDirectory memeriksa keberadaan dan status direktori
func CheckDirectory(path string) (*DirectoryInfo, error) {
	info := &DirectoryInfo{
		Path: path,
	}

	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			info.Exists = false
			return info, nil
		}
		return nil, fmt.Errorf("error checking directory %s: %w", path, err)
	}

	info.Exists = true
	info.IsDir = fileInfo.IsDir()
	info.Mode = fileInfo.Mode()
	info.Size = fileInfo.Size()
	info.LastMod = fileInfo.ModTime()

	// Jika path bukan direktori, return early
	if !info.IsDir {
		return info, nil
	}

	// Baca konten direktori
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading directory contents %s: %w", path, err)
	}

	info.IsEmpty = len(entries) == 0
	info.Children = make([]string, 0, len(entries))
	for _, entry := range entries {
		info.Children = append(info.Children, entry.Name())
	}

	return info, nil
}

// IsDirectoryEmpty memeriksa apakah direktori kosong
func IsDirectoryEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("error checking if directory is empty %s: %w", path, err)
	}
	return len(entries) == 0, nil
}

// HasPermission memeriksa apakah ada akses ke direktori
func HasPermission(path string, mode os.FileMode) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error checking directory permissions %s: %w", path, err)
	}
	return info.Mode().Perm()&mode == mode, nil
}

// GetSubDirectories mendapatkan daftar subdirektori
func GetSubDirectories(path string) ([]string, error) {
	var dirs []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading subdirectories %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(path, entry.Name()))
		}
	}
	return dirs, nil
}

// DirectoryExists mengecek apakah sebuah direktori ada dan benar-benar direktori
func DirectoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("error checking directory existence %s: %w", path, err)
	}
	return info.IsDir(), nil
}

// IsValidDirectory mengecek apakah path adalah direktori yang valid dan dapat diakses
func IsValidDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("error accessing directory %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Coba buka direktori untuk memastikan dapat diakses
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("directory is not accessible %s: %w", path, err)
	}
	defer dir.Close()

	return nil
}

// GetDirectorySize menghitung ukuran total direktori
func GetDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("error calculating directory size %s: %w", path, err)
	}
	return size, nil
}
