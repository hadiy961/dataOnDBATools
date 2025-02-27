package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CompressData compresses byte data using gzip
func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressData decompresses gzipped byte data
func DecompressData(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gr); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return buf.Bytes(), nil
}

// CompressFile compresses a file using gzip and saves it with .gz extension
func CompressFile(src string) error {
	dest := src + ".gz"

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	gw := gzip.NewWriter(destFile)
	defer gw.Close()

	if _, err := io.Copy(gw, srcFile); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	return nil
}

// DecompressFile decompresses a .gz file, removing the .gz extension
func DecompressFile(src string) error {
	if filepath.Ext(src) != ".gz" {
		return fmt.Errorf("file must have .gz extension")
	}

	dest := src[:len(src)-3] // Remove .gz extension

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, gr); err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	return nil
}

// CompressDirectory compresses all files in a directory (non-recursive)
func CompressDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		if filepath.Ext(filePath) != ".gz" {
			if err := CompressFile(filePath); err != nil {
				return fmt.Errorf("failed to compress %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// CompressDirectoryRecursive compresses all files in a directory and its subdirectories
func CompressDirectoryRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) != ".gz" {
			if err := CompressFile(path); err != nil {
				return fmt.Errorf("failed to compress %s: %w", path, err)
			}
		}

		return nil
	})
}

// IsGzipped checks if a file is gzip compressed
func IsGzipped(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first two bytes - gzip magic number is 0x1f 0x8b
	magic := make([]byte, 2)
	if _, err := file.Read(magic); err != nil {
		return false, fmt.Errorf("failed to read file header: %w", err)
	}

	return magic[0] == 0x1f && magic[1] == 0x8b, nil
}
