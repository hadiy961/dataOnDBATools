package utils

import (
	"os"
	"os/exec"
	"runtime"

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

func ClearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "linux", "darwin", "freebsd", "openbsd", "netbsd":
		// Untuk Unix-like systems (Linux, macOS, BSD)
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		// Fallback menggunakan ANSI escape codes
		// \033[H memindahkan kursor ke posisi home (0,0)
		// \033[2J membersihkan layar
		// \033[3J membersihkan scrollback buffer (opsional)
		print("\033[H\033[2J\033[3J")
	}
}
