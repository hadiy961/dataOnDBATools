package query

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// GetServerCredID retrieves a single server credential record by server ID
func GetServerCredID(db *sql.DB, id int) (*ServerCredential, error) {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Server Check ] "
	spin.Color("green")
	spin.Suffix = fmt.Sprintf(" Retrieving Server credential data with server_id %d...", id)
	spin.Start()
	time.Sleep(200 * time.Millisecond)
	defer spin.Stop()

	query := `
		SELECT 
			id, server_id, auth_id, code, name, ipaddress, port, user, pass, description
		FROM v_server_cred
		WHERE server_id = ? AND type = 'SSH'
	`

	var cred ServerCredential
	err := db.QueryRow(query, id).Scan(
		&cred.ID, &cred.ServerID, &cred.AuthID, &cred.Code, &cred.Name,
		&cred.IPAddress, &cred.Port, &cred.User, &cred.Pass, &cred.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No credential found with the given server ID
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// Add a small pause to ensure spinner is visible
	time.Sleep(200 * time.Millisecond)

	return &cred, nil
}
