// internal/app/config/Status/list.go
package status

import (
	"database/sql"
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	StatusService "dbaTools/internal/package/manage/status/service"
	tables "dbaTools/internal/ui/component/table"
	"dbaTools/internal/utils"
	"fmt"
)

type Status struct {
	ID          int
	Name        string
	Description string
}

func getStatuss(db *sql.DB) ([]Status, error) {
	rows, err := db.Query("SELECT id, name, description FROM status ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Statuss []Status
	for rows.Next() {
		var z Status
		if err := rows.Scan(&z.ID, &z.Name, &z.Description); err != nil {
			return nil, err
		}
		Statuss = append(Statuss, z)
	}
	return Statuss, nil
}

func DisplayStatusList(log *logger.Logger, dbConn *sql.DB) error {
	utils.ClearScreen()
	// Get DB connection using singleton pattern
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get Statuss data
	Statuss, err := getStatuss(cfg.DB)
	if err != nil {
		return err
	}

	// Convert to table rows if data exists
	if len(Statuss) > 0 {

		// Definisi kolom
		columns := []tables.TableColumn{
			{Title: "ID", Width: 5},
			{Title: "Nama", Width: 10},
			{Title: "Email", Width: 30},
		}

		// Inisialisasi tabel
		table := tables.NewTable(
			columns,
			tables.WithTitle("Status List"),
			tables.WithUpdateFunction(func(id string) {
				UpdateStatus(log, dbConn, id)
				DisplayStatusList(log, dbConn)
			}),
			tables.WithDeleteFunction(func(id string) {
				DeleteStatus(log, dbConn, id)
				DisplayStatusList(log, dbConn)
			}),
			tables.WithAddFunction(func() {
				AddStatus(log, dbConn)
				DisplayStatusList(log, dbConn)
			}),
			tables.WithDetailFunction(func(id string) {
				DisplayStatusDetail(log, dbConn, id)
			}),
		)

		for _, z := range Statuss {
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
		utils.ClearScreen()
		log.Info("Tidak ada data")
	}

	return nil
}

func DisplayStatusDetail(log *logger.Logger, dbConn *sql.DB, id string) error {
	utils.ClearScreen()

	// Get DB connection using singleton pattern
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get Status data
	service := StatusService.NewStatusService(cfg.DB, log)
	StatusData, err := service.GetStatusById(id)
	if err != nil {
		return fmt.Errorf("failed to get Status data: %w", err)
	}

	// Create detail table
	detailTable := tables.NewDetailTable(
		tables.WithDetailTitle("Status Detail"),
		tables.WithDetailBackFunction(func() {
			DisplayStatusList(log, dbConn)
		}),
		tables.WithDetailUpdateFunction(func() {
			UpdateStatus(log, dbConn, id)
			DisplayStatusList(log, dbConn)
		}),
		tables.WithDetailDeleteFunction(func() {
			DeleteStatus(log, dbConn, id)
			DisplayStatusList(log, dbConn)
		}),
	)

	// Add rows
	detailTable.AddDetailRow("Status ID", StatusData.Id)
	detailTable.AddDetailRow("Name", StatusData.Name)
	detailTable.AddDetailRow("Description", StatusData.Description)
	detailTable.AddDetailRow("Created At", StatusData.CreatedAt.String())
	detailTable.AddDetailRow("Created By", StatusData.CreatedBy)
	detailTable.AddDetailRow("Updated At", StatusData.UpdatedAt.String())
	detailTable.AddDetailRow("Updated By", StatusData.UpdatedBy)

	// Run table
	return tables.RunDetailTable(detailTable)
}
