package server

import (
	"database/sql"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server/query"
	"dbaTools/internal/utils"
	"dbaTools/internal/utils/ssh"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
)

func GetServerInfoByID(log *logger.Logger, dbConn *sql.DB, id int) error {
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

		// Get system info
		sysInfo, infoErr := utils.GetSystemInfoNew(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
		if infoErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve system info for %s (%s)", servers.Name, cred.IPAddress), infoErr)
			fmt.Printf("System Information: \033[31mFAILED\033[0m (%s)\n", infoErr.Error())
		} else {
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

		// After successfully getting system info, get storage info
		storageInfoList, storageErr := utils.GetServerStorage(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
		if storageErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve storage info for %s (%s)", servers.Name, cred.IPAddress), storageErr)
			fmt.Printf("Storage Information: \033[31mFAILED\033[0m (%s)\n", storageErr.Error())
		} else {
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
			fmt.Println()
		}

		// Check services
		fmt.Println("Service Information:")
		serviceChecker := utils.NewServiceChecker(true, cred.IPAddress, cred.Port, cred.User, cred.Pass)
		services, serviceErr := serviceChecker.CheckServices(servers.ID)

		if serviceErr != nil {
			log.Error(fmt.Sprintf("Failed to retrieve service info for %s (%s)", servers.Name, cred.IPAddress), serviceErr)
		} else {
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
	}

	return nil
}
