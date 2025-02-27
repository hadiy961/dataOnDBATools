// internal/app/config/zone/list.go
package zone

import (
	"database/sql"
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	zoneService "dbaTools/internal/package/manage/zone/service"
	tables "dbaTools/internal/ui/component/table"
	TUIutils "dbaTools/internal/ui/component/utils"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

type Zone struct {
	ID          int
	Name        string
	Description string
}

func getZones(db *sql.DB) ([]Zone, error) {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Zone List ] "
	spin.Color("green")
	spin.Suffix = " Retrieving zone data... \n"
	spin.Start()
	time.Sleep(200 * time.Millisecond)

	defer spin.Stop()

	rows, err := db.Query("SELECT id, name, description FROM zone ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	for rows.Next() {
		var z Zone
		if err := rows.Scan(&z.ID, &z.Name, &z.Description); err != nil {
			return nil, err
		}
		zones = append(zones, z)
	}
	return zones, nil
}

func DisplayZoneList(log *logger.Logger, dbConn *sql.DB) error {
	TUIutils.ClearScreen()
	// Get DB connection using singleton pattern
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get zones data
	zones, err := getZones(cfg.DB)
	if err != nil {
		return err
	}

	// Convert to table rows if data exists
	if len(zones) > 0 {

		// Definisi kolom
		columns := []tables.TableColumn{
			{Title: "ID", Width: 5},
			{Title: "Nama", Width: 10},
			{Title: "Email", Width: 30},
		}

		// Inisialisasi tabel
		table := tables.NewTable(
			columns,
			tables.WithTitle("Zones List"),
			tables.WithUpdateFunction(func(id string) {
				UpdateZone(log, dbConn, id)
				DisplayZoneList(log, dbConn)
			}),
			tables.WithDeleteFunction(func(id string) {
				DeleteZone(log, dbConn, id)
				DisplayZoneList(log, dbConn)
			}),
			tables.WithAddFunction(func() {
				AddZone(log, dbConn)
				DisplayZoneList(log, dbConn)
			}),
			tables.WithDetailFunction(func(id string) {
				DisplayZoneDetail(log, dbConn, id)
			}),
		)

		for _, z := range zones {
			idStr := fmt.Sprintf("%d", z.ID)
			table.AddRow(idStr, []string{
				idStr,
				z.Name,
				z.Description,
			})
		}

		// Tambah data

		// Jalankan tabel
		tables.RunTable(table)
	} else {
		TUIutils.ClearScreen()
		log.Info("Tidak ada data")
	}

	return nil
}

func DisplayZoneDetail(log *logger.Logger, dbConn *sql.DB, id string) error {
	TUIutils.ClearScreen()

	// Get DB connection using singleton pattern
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get zone data
	service := zoneService.NewZoneService(cfg.DB, log)
	zoneData, err := service.GetZoneById(id)
	if err != nil {
		return fmt.Errorf("failed to get zone data: %w", err)
	}

	// Create detail table
	detailTable := tables.NewDetailTable(
		tables.WithDetailTitle("Zone Detail"),
		tables.WithDetailBackFunction(func() {
			DisplayZoneList(log, dbConn)
		}),
		tables.WithDetailUpdateFunction(func() {
			UpdateZone(log, dbConn, id)
			DisplayZoneList(log, dbConn)
		}),
		tables.WithDetailDeleteFunction(func() {
			DeleteZone(log, dbConn, id)
			DisplayZoneList(log, dbConn)
		}),
	)

	// Add rows
	detailTable.AddDetailRow("Zone ID", zoneData.Id)
	detailTable.AddDetailRow("Name", zoneData.Name)
	detailTable.AddDetailRow("Description", zoneData.Description)
	detailTable.AddDetailRow("Created At", zoneData.CreatedAt.String())
	detailTable.AddDetailRow("Created By", zoneData.CreatedBy)
	detailTable.AddDetailRow("Updated At", zoneData.UpdatedAt.String())
	detailTable.AddDetailRow("Updated By", zoneData.UpdatedBy)

	// Run table
	return tables.RunDetailTable(detailTable)
}
