package unused

import (
	"database/sql"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server/query"
	"dbaTools/internal/utils"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"time"
)

func AllServerCheck(log *logger.Logger, dbConn *sql.DB) error {
	// Use the provided database connection instead of trying to get a new one
	if dbConn == nil {
		return fmt.Errorf("invalid database connection")
	}
	// Get servers data using the provided connection
	servers, err := query.GetAllServerHead(dbConn)
	if err != nil {
		log.Error("Failed to retrieve server data", err)
		return err
	}

	// Allow time for spinner to finish
	time.Sleep(200 * time.Millisecond)
	// Define a counter for successful connections
	successCount := 0

	// Convert to table rows if data exists
	if len(servers) > 0 {
		// Define columns

		// Add rows to the table
		for _, s := range servers {
			// Logic untuk check server
			cred, err := query.GetServerCredID(dbConn, s.ID)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to retrieve credentials for server %s", s.Code), err)
				continue
			}

			if cred == nil {
				log.Warning("No credentials found for server " + s.Code)
				continue
			}

			// Check SSH connection
			connected, err := ssh.CheckSSHConnection(cred.IPAddress, cred.Port, cred.User, cred.Pass)
			if err != nil {
				log.Error(fmt.Sprintf("SSH connection failed for %s (%s)", s.Name, cred.IPAddress), err)
			} else if connected {
				successCount++
				log.Info(fmt.Sprintf("SSH connection successful for %s (%s)", s.Name, cred.IPAddress))
				// Get system information since the SSH connection was successful
				sysInfo, infoErr := utils.GetSystemInfoNew(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
				if infoErr != nil {
					log.Error(fmt.Sprintf("Failed to retrieve system info for %s (%s)", s.Name, cred.IPAddress), infoErr)
				} else {
					// Save the system information to the database
					if err := query.InsertServerDetail(dbConn, s.ID, sysInfo); err != nil {
						log.Error(fmt.Sprintf("Failed to save system info for %s (%s)", s.Name, cred.IPAddress), err)
					} else {
						log.Info(fmt.Sprintf("System information for %s (%s) saved to database", s.Name, cred.IPAddress))
					}
				}
			}
		}

		// Show the total count
		fmt.Printf("Total servers: %d\n\n", len(servers))

	} else {
		log.Info("No server data found")
		fmt.Println("No server data found in the database.")
	}

	return nil
}
