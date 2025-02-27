package lokal

import "fmt"

type SetupType int

const (
	LocalSetup SetupType = iota + 1
	RemoteSetup
)

// InstallError represents custom installation error types
type InstallError struct {
	Stage   string
	Message string
	Err     error
}

func (e *InstallError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Stage, e.Message, e.Err)
}
