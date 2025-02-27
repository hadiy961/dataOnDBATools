// File: internal/package/manage/server/query/server.go
package query

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// GetAllServerHead retrieves all records from the server_head table
func GetAllServerHead(db *sql.DB) ([]ServerHead, error) {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Server Check ] "
	spin.Color("green")
	spin.Suffix = " Retrieving All Server data..."
	spin.Start()
	time.Sleep(500 * time.Millisecond)
	defer spin.Stop()

	query := `
		SELECT 
			id, code, name, ipaddress, port, description, zone_id,
			created_at, created_by, updated_at, updated_by, category_id, type
		FROM server_head where type = 'SSH'
		ORDER BY id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var servers []ServerHead
	for rows.Next() {
		var s ServerHead
		var createdAt, updatedAt []byte

		err := rows.Scan(
			&s.ID, &s.Code, &s.Name, &s.IPAddress, &s.Port, &s.Description, &s.ZoneID,
			&createdAt, &s.CreatedBy, &updatedAt, &s.UpdatedBy, &s.CategoryID, &s.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Parse timestamp strings to time.Time
		if len(createdAt) > 0 {
			t, err := time.Parse("2006-01-02 15:04:05", string(createdAt))
			if err == nil {
				s.CreatedAt.Time = t
				s.CreatedAt.Valid = true
			}
		}

		if len(updatedAt) > 0 {
			t, err := time.Parse("2006-01-02 15:04:05", string(updatedAt))
			if err == nil {
				s.UpdatedAt.Time = t
				s.UpdatedAt.Valid = true
			}
		}

		servers = append(servers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Add a small pause to ensure spinner is visible
	time.Sleep(200 * time.Millisecond)

	return servers, nil
}

// GetServerHeadByID retrieves a single server_head record by ID
func GetServerHeadByID(db *sql.DB, id int) (*ServerHead, error) {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Server Check ] "
	spin.Color("green")
	spin.Suffix = fmt.Sprintf(" Retrieving Server data with id %d...", id)
	spin.Start()
	time.Sleep(200 * time.Millisecond)
	defer spin.Stop()

	query := `
		SELECT 
			id, code, name, ipaddress, port, description, zone_id,
			created_at, created_by, updated_at, updated_by, category_id, type
		FROM server_head
		WHERE id = ?
	`

	var s ServerHead
	var createdAt, updatedAt []byte

	err := db.QueryRow(query, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.IPAddress, &s.Port, &s.Description, &s.ZoneID,
		&createdAt, &s.CreatedBy, &updatedAt, &s.UpdatedBy, &s.CategoryID, &s.Type,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No server found with the given ID
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// Parse timestamp strings to time.Time
	if len(createdAt) > 0 {
		t, err := time.Parse("2006-01-02 15:04:05", string(createdAt))
		if err == nil {
			s.CreatedAt.Time = t
			s.CreatedAt.Valid = true
		}
	}

	if len(updatedAt) > 0 {
		t, err := time.Parse("2006-01-02 15:04:05", string(updatedAt))
		if err == nil {
			s.UpdatedAt.Time = t
			s.UpdatedAt.Valid = true
		}
	}

	// Add a small pause to ensure spinner is visible
	time.Sleep(200 * time.Millisecond)

	return &s, nil
}
