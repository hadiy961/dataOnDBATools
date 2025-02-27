package utils

import "github.com/dustin/go-humanize"

// FormatBytes formats bytes into a human-readable string
func FormatBytes(bytes uint64) string {
	return humanize.Bytes(bytes)
}
