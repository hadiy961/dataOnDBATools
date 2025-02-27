package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// GetDB returns the underlying *sql.DB connection
func (d *DBConnection) GetDB() *sql.DB {
	return d.db
}

func (d *DBConnection) GetConfig() DBConfig {
	return d.config
}

// NewDBConnection creates a new database connection instance
func NewDBConnection(config DBConfig) *DBConnection {
	return &DBConnection{
		config: config,
	}
}

// Connect establishes connection to database
func (d *DBConnection) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		d.config.User,
		d.config.Password,
		d.config.Host,
		d.config.Port,
		d.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("%w", err)
	}

	d.db = db
	return nil
}

// Close closes the database connection
func (d *DBConnection) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// CheckConnection verifies database connection
func (d *DBConnection) CheckConnection() error {
	if d.db == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return d.db.Ping()
}

// CheckCredentials verifies user credentials
func (d *DBConnection) CheckCredentials() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		d.config.User,
		d.config.Password,
		d.config.Host,
		d.config.Port,
	)

	tempDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}
	defer tempDB.Close()

	return tempDB.Ping()
}

// CheckDatabaseExists verifies if database exists
func (d *DBConnection) CheckDatabaseExists() (bool, error) {
	query := `SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?`

	var dbName string
	err := d.db.QueryRow(query, d.config.Database).Scan(&dbName)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking database existence: %w", err)
	}

	return true, nil
}
