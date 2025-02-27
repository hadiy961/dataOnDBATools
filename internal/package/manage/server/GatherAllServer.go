package server

import (
	"database/sql"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server/query"
	"dbaTools/internal/utils"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"sync"
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

	// Process servers sequentially to avoid spinner conflicts
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

			// Log SSH connection attempt (outside the mutex because CheckSSHConnection has its own spinner)
			syncLog("info", fmt.Sprintf("Connecting to %s@%s:%d...", cred.User, cred.IPAddress, cred.Port), nil)

			// Use mutex to prevent spinner conflicts in CheckSSHConnection
			spinnerMutex.Lock()
			connected, err := ssh.CheckSSHConnection(cred.IPAddress, cred.Port, cred.User, cred.Pass)
			spinnerMutex.Unlock()

			if err != nil {
				syncLog("error", fmt.Sprintf("SSH connection failed for %s (%s)", server.Name, cred.IPAddress), err)
				return
			}

			if !connected {
				syncLog("warning", fmt.Sprintf("Could not establish SSH connection to %s (%s)", server.Name, cred.IPAddress), nil)
				return
			}

			syncLog("info", fmt.Sprintf("Processing server %s (%s)", server.Name, cred.IPAddress), nil)

			// Create channels to collect results from concurrent operations
			sysInfoCh := make(chan *utils.SystemInfo, 1)
			storageInfoCh := make(chan []utils.StorageInfo, 1)
			serviceInfoCh := make(chan []utils.ServiceInfo, 1)

			// Gather data one by one to avoid spinner conflicts

			// System info
			syncLog("info", fmt.Sprintf("Gathering system information for %s (%s)", server.Name, cred.IPAddress), nil)
			spinnerMutex.Lock()
			sysInfo, sysErr := utils.GetSystemInfoNew(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			spinnerMutex.Unlock()

			if sysErr != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve system info for %s (%s)", server.Name, cred.IPAddress), sysErr)
				sysInfoCh <- nil
			} else {
				sysInfoCh <- sysInfo
			}

			// Storage info
			syncLog("info", fmt.Sprintf("Gathering storage information for %s (%s)", server.Name, cred.IPAddress), nil)
			spinnerMutex.Lock()
			storageInfoList, storageErr := utils.GetServerStorage(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			spinnerMutex.Unlock()

			if storageErr != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve storage info for %s (%s)", server.Name, cred.IPAddress), storageErr)
				storageInfoCh <- nil
			} else {
				storageInfoCh <- storageInfoList
			}

			// Service info
			syncLog("info", fmt.Sprintf("Gathering service information for %s (%s)", server.Name, cred.IPAddress), nil)
			spinnerMutex.Lock()
			serviceChecker := utils.NewServiceChecker(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			services, serviceErr := serviceChecker.CheckServices(server.ID)
			spinnerMutex.Unlock()

			if serviceErr != nil {
				syncLog("error", fmt.Sprintf("Failed to retrieve service info for %s (%s)", server.Name, cred.IPAddress), serviceErr)
				serviceInfoCh <- nil
			} else {
				serviceInfoCh <- services
			}

			// Process system info
			sysInfo = <-sysInfoCh
			if sysInfo != nil {
				spinnerMutex.Lock()
				if err := query.InsertServerDetail(dbConn, server.ID, sysInfo); err != nil {
					log.Error(fmt.Sprintf("Failed to save system info for %s (%s)", server.Name, cred.IPAddress), err)
				} else {
					log.Info(fmt.Sprintf("System information for %s (%s) saved to database", server.Name, cred.IPAddress))
				}
				spinnerMutex.Unlock()
			}

			// Process storage info
			storageInfoList = <-storageInfoCh
			if storageInfoList != nil {
				spinnerMutex.Lock()
				if err := query.InsertServerStorage(dbConn, server.ID, storageInfoList); err != nil {
					log.Error(fmt.Sprintf("Failed to save storage info for %s", server.Name), err)
				} else {
					log.Info(fmt.Sprintf("Storage information for %s (%s) saved to database", server.Name, cred.IPAddress))
				}
				spinnerMutex.Unlock()
			}

			// Process service info
			services = <-serviceInfoCh
			if services != nil {
				spinnerMutex.Lock()
				if err := query.InsertServiceInfo(dbConn, services); err != nil {
					log.Error(fmt.Sprintf("Failed to save service info for %s", server.Name), err)
				} else {
					log.Info(fmt.Sprintf("Service information for %s (%s) saved to database", server.Name, cred.IPAddress))
				}
				spinnerMutex.Unlock()
			}

			syncLog("success", fmt.Sprintf("Completed gathering information for %s (%s)", server.Name, cred.IPAddress), nil)
		}(server)

		// Small delay between launching server goroutines
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for all server goroutines to complete
	serverWg.Wait()

	syncLog("success", fmt.Sprintf("Completed gathering information for all %d servers", len(servers)), nil)
	return nil
}
