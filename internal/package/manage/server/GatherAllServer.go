package server

import (
	"database/sql"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server/query"
	"dbaTools/internal/utils"
	"fmt"
	"sync"
	"time"
)

func GatherAllServer(log *logger.Logger, dbConn *sql.DB) error {
	// Validate database connection
	if dbConn == nil {
		return fmt.Errorf("invalid database connection")
	}

	// Get servers data
	servers, err := query.GetAllServerHead(dbConn)
	if err != nil {
		log.Error("Failed to retrieve server data", err)
		return err
	}

	time.Sleep(200 * time.Millisecond)

	// Check if servers exist
	if len(servers) == 0 {
		log.Info("No server data found")
		fmt.Println("No server data found in the database.")
		return nil
	}

	// Mutex for synchronized logging and operations that create spinners
	spinnerMutex := &sync.Mutex{}

	// Create a synchronized logger function
	syncLog := func(level, message string, err error) {
		spinnerMutex.Lock()
		defer spinnerMutex.Unlock()

		switch level {
		case "info":
			log.Info(message)
		case "warning":
			log.Warning(message)
		case "error":
			log.Error(message, err)
		case "success":
			log.Success(message)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Wait group for all servers
	var serverWg sync.WaitGroup

	// Process each server
	for _, server := range servers {
		serverWg.Add(1)

		// Process each server in a separate goroutine
		go func(server query.ServerHead) {
			defer serverWg.Done()

			// Get server credentials
			cred, err := query.GetServerCredID(dbConn, server.ID)
			if err != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve credentials for server %s", server.Code), err)
				return
			}

			if cred == nil {
				syncLog("warning", "No credentials found for server "+server.Code, nil)
				return
			}

			// syncLog("info", fmt.Sprintf("Processing server %s (%s)", server.Name, cred.IPAddress), nil)

			// Process system information
			syncLog("info", fmt.Sprintf("Gathering system information for %s", server.Name), nil)
			spinnerMutex.Lock()
			sysInfo, sysErr := utils.GetSystemInfoNew(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			spinnerMutex.Unlock()

			if sysErr != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve system info for %s", server.Name), sysErr)
			} else {
				// Save system information
				spinnerMutex.Lock()
				if err := query.InsertServerDetail(dbConn, server.ID, sysInfo); err != nil {
					log.Error(fmt.Sprintf("Failed to save system info for %s", server.Name), err)
				} else {
					log.Success(fmt.Sprintf("System information for %s saved to database", server.Name))
				}
				spinnerMutex.Unlock()
			}

			// Process storage information
			syncLog("info", fmt.Sprintf("Gathering storage information for %s", server.Name), nil)
			spinnerMutex.Lock()
			storageInfoList, storageErr := utils.GetServerStorage(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			spinnerMutex.Unlock()

			if storageErr != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve storage info for %s", server.Name), storageErr)
			} else {
				// Save storage information
				spinnerMutex.Lock()
				if err := query.InsertServerStorage(dbConn, server.ID, storageInfoList); err != nil {
					log.Error(fmt.Sprintf("Failed to save storage info for %s", server.Name), err)
				} else {
					log.Success(fmt.Sprintf("Storage information for %s saved to database", server.Name))
				}
				spinnerMutex.Unlock()
			}

			// Process service information
			syncLog("info", fmt.Sprintf("Gathering service information for %s", server.Name), nil)
			spinnerMutex.Lock()
			serviceChecker := utils.NewServiceChecker(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			services, serviceErr := serviceChecker.CheckServices(server.ID)
			spinnerMutex.Unlock()

			if serviceErr != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve service info for %s", server.Name), serviceErr)
			} else {
				// Save service information
				spinnerMutex.Lock()
				if err := query.InsertServiceInfo(dbConn, services); err != nil {
					log.Error(fmt.Sprintf("Failed to save service info for %s", server.Name), err)
				} else {
					log.Success(fmt.Sprintf("Service information for %s saved to database", server.Name))
				}
				spinnerMutex.Unlock()
			}

			syncLog("success", fmt.Sprintf("Completed gathering information for %s", server.Name), nil)
		}(server)

		// Small delay between launching server goroutines
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for all server goroutines to complete
	serverWg.Wait()

	syncLog("success", fmt.Sprintf("Completed gathering information for all %d servers", len(servers)), nil)
	return nil
}
