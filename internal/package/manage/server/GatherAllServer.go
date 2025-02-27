package server

import (
	"database/sql"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server/query"
	"dbaTools/internal/utils"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"time"
)

func GatherAllServer(log *logger.Logger, dbConn *sql.DB) error {
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
	time.Sleep(200 * time.Millisecond)
	if len(servers) > 0 {
		for _, s := range servers {
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
				//Get systeminfo
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

				// After successfully getting system info
				storageInfoList, storageErr := utils.GetServerStorage(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
				if storageErr != nil {
					log.Error(fmt.Sprintf("Failed to retrieve storage info for %s (%s)", s.Name, cred.IPAddress), storageErr)
				} else {
					// Save storage information to database
					if err := query.InsertServerStorage(dbConn, s.ID, storageInfoList); err != nil {
						log.Error(fmt.Sprintf("Failed to save storage info for %s", s.Name), err)
					} else {
						log.Info(fmt.Sprintf("Storage information for %s (%s) saved to database", s.Name, cred.IPAddress))
					}
				}

				serviceChecker := utils.NewServiceChecker(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
				services, serviceErr := serviceChecker.CheckServices(s.ID)
				if serviceErr != nil {
					log.Error(fmt.Sprintf("Failed to retrieve service info for %s (%s)", s.Name, cred.IPAddress), serviceErr)
				} else {
					// Save service information to database
					if err := query.InsertServiceInfo(dbConn, services); err != nil {
						log.Error(fmt.Sprintf("Failed to save service info for %s", s.Name), err)
					} else {
						log.Info(fmt.Sprintf("Service information for %s (%s) saved to database", s.Name, cred.IPAddress))
					}
				}
			}

		}
	} else {
		log.Info("No server data found")
		fmt.Println("No server data found in the database.")
	}

	return nil
}
