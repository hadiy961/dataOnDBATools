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

func GatherServerByID(log *logger.Logger, dbConn *sql.DB, id int) error {
	// Use the provided database connection instead of trying to get a new one
	if dbConn == nil {
		return fmt.Errorf("invalid database connection")
	}
	// Get servers data using the provided connection
	servers, err := query.GetServerHeadByID(dbConn, id)
	if err != nil {
		log.Error("Failed to retrieve server data", err)
		return err
	}
	time.Sleep(101 * time.Millisecond)
	cred, err := query.GetServerCredID(dbConn, servers.ID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to retrieve credentials for server %s", servers.Code), err)
	}

	if cred == nil {
		log.Warning("No credentials found for server " + servers.Code)
	}

	// Check SSH connection
	connected, err := ssh.CheckSSHConnection(cred.IPAddress, cred.Port, cred.User, cred.Pass)
	if err != nil {
		log.Error(fmt.Sprintf("SSH connection failed for %s (%s)", servers.Name, cred.IPAddress), err)
	} else if connected {
		//Get systeminfo
		sysInfo, infoErr := utils.GetSystemInfoNew(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
		if infoErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve system info for %s (%s)", servers.Name, cred.IPAddress), infoErr)
		} else {
			// Save the system information to the database
			if err := query.InsertServerDetail(dbConn, servers.ID, sysInfo); err != nil {
				log.Error(fmt.Sprintf("Failed to save system info for %s (%s)", servers.Name, cred.IPAddress), err)
			} else {
				log.Info(fmt.Sprintf("System information for %s (%s) saved to database", servers.Name, cred.IPAddress))
			}
		}

		// After successfully getting system info
		storageInfoList, storageErr := utils.GetServerStorage(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
		if storageErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve storage info for %s (%s)", servers.Name, cred.IPAddress), storageErr)
		} else {
			// Save storage information to database
			if err := query.InsertServerStorage(dbConn, servers.ID, storageInfoList); err != nil {
				log.Error(fmt.Sprintf("Failed to save storage info for %s", servers.Name), err)
			} else {
				log.Info(fmt.Sprintf("Storage information for %s (%s) saved to database", servers.Name, cred.IPAddress))
			}
		}

		serviceChecker := utils.NewServiceChecker(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
		services, serviceErr := serviceChecker.CheckServices(servers.ID)
		if serviceErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve service info for %s (%s)", servers.Name, cred.IPAddress), serviceErr)
		} else {
			// Save service information to database
			if err := query.InsertServiceInfo(dbConn, services); err != nil {
				log.Error(fmt.Sprintf("Failed to save service info for %s", servers.Name), err)
			} else {
				log.Info(fmt.Sprintf("Service information for %s (%s) saved to database", servers.Name, cred.IPAddress))
			}
		}
	}

	return nil
}
