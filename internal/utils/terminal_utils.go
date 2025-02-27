package utils

import (
	"os"

	"golang.org/x/term"
)

type TerminalSize struct {
	Width  int
	Height int
}

// GetTerminalSize returns the current terminal dimensions
func GetTerminalSize() (*TerminalSize, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, err
	}
	return &TerminalSize{Width: width, Height: height}, nil
}
