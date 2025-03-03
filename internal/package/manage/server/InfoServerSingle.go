package server

import (
	"database/sql"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server/query"
	"dbaTools/internal/utils"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

func GetServerInfoByID(log *logger.Logger, dbConn *sql.DB, id int) error {
	// Use the provided database connection instead of trying to get a new one
	// Check if database connection is nil
	if dbConn == nil {
		log.Error("Database connection is nil", fmt.Errorf("invalid database connection"))
		return fmt.Errorf("invalid database connection")
	}

	// Get servers data using the provided connection
	servers, err := query.GetServerHeadByID(dbConn, id)
	if err != nil {
		log.Error("Failed to retrieve server data", err)
		return err
	}
	// Check if server data is nil
	if servers == nil {
		log.Error("No server found with ID", fmt.Errorf("server with ID %d not found", id))
		return fmt.Errorf("server with ID %d not found", id)
	}

	time.Sleep(200 * time.Millisecond)

	cred, err := query.GetServerCredID(dbConn, servers.ID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to retrieve credentials for server %s", servers.Code), err)
	}

	if cred == nil {
		log.Warning("No credentials found for server " + servers.Code)
		return fmt.Errorf("no credentials found for server %s", servers.Code)
	}

	// Print server information header
	fmt.Println("\n==========================================================")
	fmt.Printf("Server Information for: %s (%s)\n", servers.Name, cred.IPAddress)
	fmt.Println("==========================================================")

	// Check SSH connection
	connected, err := ssh.CheckSSHConnection(cred.IPAddress, cred.Port, cred.User, cred.Pass)
	if err != nil {
		log.Error(fmt.Sprintf("SSH connection failed for %s (%s)", servers.Name, cred.IPAddress), err)
		fmt.Printf("SSH Connection: \033[31mFAILED\033[0m (%s)\n", err.Error())
		return err
	} else if connected {
		fmt.Printf("SSH Connection: \033[32mSUCCESSFUL\033[0m\n\n")

		// Define channels to collect results from concurrent operations
		sysInfoCh := make(chan *utils.SystemInfo, 1)
		sysInfoErrCh := make(chan error, 1)
		storageInfoCh := make(chan []utils.StorageInfo, 1)
		storageInfoErrCh := make(chan error, 1)
		serviceInfoCh := make(chan []utils.ServiceInfo, 1)
		serviceInfoErrCh := make(chan error, 1)

		// WaitGroup to ensure all goroutines complete
		var wg sync.WaitGroup
		wg.Add(3) // Adding 3 for the three concurrent operations

		// Create a mutex for synchronized console output
		var outputMutex sync.Mutex

		// Goroutine for system info
		go func() {
			defer wg.Done()
			sysInfo, infoErr := utils.GetSystemInfoNew(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			if infoErr != nil {
				sysInfoErrCh <- infoErr
				sysInfoCh <- nil
				return
			}
			sysInfoCh <- sysInfo
			sysInfoErrCh <- nil
		}()

		// Goroutine for storage info
		go func() {
			defer wg.Done()
			storageInfo, storageErr := utils.GetServerStorage(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			if storageErr != nil {
				storageInfoErrCh <- storageErr
				storageInfoCh <- nil
				return
			}
			storageInfoCh <- storageInfo
			storageInfoErrCh <- nil
		}()

		// Goroutine for service info
		go func() {
			defer wg.Done()
			serviceChecker := utils.NewServiceChecker(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
			services, serviceErr := serviceChecker.CheckServices(servers.ID)
			if serviceErr != nil {
				serviceInfoErrCh <- serviceErr
				serviceInfoCh <- nil
				return
			}
			serviceInfoCh <- services
			serviceInfoErrCh <- nil
		}()

		// Wait for all goroutines to complete
		wg.Wait()

		// Close all channels
		close(sysInfoCh)
		close(sysInfoErrCh)
		close(storageInfoCh)
		close(storageInfoErrCh)
		close(serviceInfoCh)
		close(serviceInfoErrCh)

		// Process system information result
		sysInfo := <-sysInfoCh
		sysInfoErr := <-sysInfoErrCh

		outputMutex.Lock()
		if sysInfoErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve system info for %s (%s)", servers.Name, cred.IPAddress), sysInfoErr)
			fmt.Printf("System Information: \033[31mFAILED\033[0m (%s)\n", sysInfoErr.Error())
		} else if sysInfo != nil {
			// Display system information using tablewriter
			fmt.Println("System Information:")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Property", "Value"})
			table.SetBorder(false)
			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
			table.SetBorder(true)
			table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor})

			table.Append([]string{"Hostname", sysInfo.Hostname})
			table.Append([]string{"OS Type", sysInfo.OSType})
			table.Append([]string{"OS Version", sysInfo.OSVersion})
			table.Append([]string{"CPU Model", sysInfo.CPUModel})
			table.Append([]string{"CPU Cores", sysInfo.CPUCore})
			table.Append([]string{"Total RAM", sysInfo.TotalRAM})

			table.Render()

			// Save the system information to the database
			if err := query.InsertServerDetail(dbConn, servers.ID, sysInfo); err != nil {
				log.Error(fmt.Sprintf("Failed to save system info for %s (%s)", servers.Name, cred.IPAddress), err)
			} else {
				log.Info(fmt.Sprintf("System information for %s (%s) saved to database", servers.Name, cred.IPAddress))
			}
			fmt.Println()
		}
		outputMutex.Unlock()

		// Process storage information result
		storageInfoList := <-storageInfoCh
		storageInfoErr := <-storageInfoErrCh

		outputMutex.Lock()
		if storageInfoErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve storage info for %s (%s)", servers.Name, cred.IPAddress), storageInfoErr)
			fmt.Printf("Storage Information: \033[31mFAILED\033[0m (%s)\n", storageInfoErr.Error())
		} else if storageInfoList != nil {
			// Display storage information using tablewriter
			fmt.Println("Storage Information:")

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Device", "Mount Point", "Total", "Used", "Free", "Use %"})
			table.SetBorder(true)
			table.SetColumnAlignment([]int{
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_RIGHT,
				tablewriter.ALIGN_RIGHT,
				tablewriter.ALIGN_RIGHT,
				tablewriter.ALIGN_RIGHT,
			})

			for _, storage := range storageInfoList {
				usePercent := fmt.Sprintf("%d%%", storage.UsePercent)
				table.Append([]string{
					storage.DeviceName,
					storage.MountPoint,
					storage.TotalSpace,
					storage.UsedSpace,
					storage.FreeSpace,
					usePercent,
				})
			}

			table.Render()

			// Save storage information to database
			if err := query.InsertServerStorage(dbConn, servers.ID, storageInfoList); err != nil {
				log.Error(fmt.Sprintf("Failed to save storage info for %s", servers.Name), err)
			} else {
				log.Info(fmt.Sprintf("Storage information for %s (%s) saved to database", servers.Name, cred.IPAddress))
			}
			fmt.Println()
		}
		outputMutex.Unlock()

		// Process service information result
		services := <-serviceInfoCh
		serviceInfoErr := <-serviceInfoErrCh

		outputMutex.Lock()
		fmt.Println("Service Information:")
		if serviceInfoErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve service info for %s (%s)", servers.Name, cred.IPAddress), serviceInfoErr)
			fmt.Printf("Service Information: \033[31mFAILED\033[0m (%s)\n", serviceInfoErr.Error())
		} else if services != nil {
			// Display service information using tablewriter
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service", "Type", "Status", "Version", "Port", "Auto Start", "Install Path", "Config Path"})
			table.SetBorder(true)
			table.SetColumnAlignment([]int{
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_CENTER,
				tablewriter.ALIGN_CENTER,
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_LEFT,
			})

			for _, service := range services {
				table.Append([]string{
					service.ServiceName,
					service.ServiceType,
					utils.ColorizeStatus(service.Status),
					service.Version,
					utils.FormatServicePort(service.Port),
					utils.FormatAutoStart(service.AutoStart),
					service.InstallPath,
					service.ConfigFile,
				})
			}

			table.Render()

			// Save service information to database
			if err := query.InsertServiceInfo(dbConn, services); err != nil {
				log.Error(fmt.Sprintf("Failed to save service info for %s", servers.Name), err)
			} else {
				log.Info(fmt.Sprintf("Service information for %s (%s) saved to database", servers.Name, cred.IPAddress))
			}
		}
		outputMutex.Unlock()
	}

	return nil
}
