package database

import (
	"database/sql"
	"fmt"
)

// Select executes a SELECT query and returns rows
func (d *DBConnection) Select(query string, args ...interface{}) (*sql.Rows, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing select query: %w", err)
	}

	return rows, nil
}

// Insert executes an INSERT query
func (d *DBConnection) Insert(query string, args ...interface{}) (int64, error) {
	if d.db == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	result, err := d.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("error executing insert query: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID: %w", err)
	}

	return id, nil
}

// Update executes an UPDATE query
func (d *DBConnection) Update(query string, args ...interface{}) (int64, error) {
	if d.db == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	result, err := d.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("error executing update query: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting affected rows: %w", err)
	}

	return affected, nil
}

// Delete executes a DELETE query
func (d *DBConnection) Delete(query string, args ...interface{}) (int64, error) {
	if d.db == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	result, err := d.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("error executing delete query: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting affected rows: %w", err)
	}

	return affected, nil
}

// Transaction starts a new transaction
func (d *DBConnection) Transaction() (*sql.Tx, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	return d.db.Begin()
}

// ExecuteInTransaction executes multiple queries in a transaction
func (d *DBConnection) ExecuteInTransaction(queries []string) error {
	tx, err := d.Transaction()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	for _, query := range queries {
		_, err := tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing query in transaction: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// CreateDatabase creates a new database
func (d *DBConnection) CreateDatabase() error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", d.config.Database)
	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}
	return nil
}

// DropDatabase drops an existing database
func (d *DBConnection) DropDatabase() error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", d.config.Database)
	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error dropping database: %w", err)
	}
	return nil
}
