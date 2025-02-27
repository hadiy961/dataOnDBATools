package ssh

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// CheckSSHConnection checks if SSH connection is possible
func CheckSSHConnection(host string, port int, user, password string) (bool, error) {
	// Create spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ SSH Check ] "
	spin.Suffix = fmt.Sprintf(" Connecting to %s@%s:%d...", user, host, port)
	spin.Color("blue")
	spin.Start()
	defer spin.Stop()

	// Get SSH client
	client, err := GetSSHClient(host, port, user, password)
	if err != nil {
		spin.FinalMSG = fmt.Sprintf("Connection to %s@%s:%d failed.\n", user, host, port)
		return false, err
	}
	defer client.Close()

	// Add a slight delay so the spinner completion message is visible
	time.Sleep(200 * time.Millisecond)

	return true, nil
}
