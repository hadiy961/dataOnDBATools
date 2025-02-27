package database

import (
	"database/sql"
	"time"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type DBConnection struct {
	db     *sql.DB
	config DBConfig
}

// QueryResult represents a generic query result
type QueryResult struct {
	LastInsertId int64
	RowsAffected int64
	Error        error
}

// DBStats represents database statistics
type DBStats struct {
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

// TableInfo represents table information
type TableInfo struct {
	Name    string
	Columns []ColumnInfo
	Engine  string
	Rows    int64
	Size    int64
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Name          string
	Type          string
	Nullable      bool
	Key           string
	DefaultValue  interface{}
	AutoIncrement bool
}

// DatabaseInfo represents database information
type DatabaseInfo struct {
	Name         string
	Charset      string
	Collation    string
	Tables       []TableInfo
	Size         int64
	CreatedAt    time.Time
	TableCount   int
	ViewCount    int
	ProcCount    int
	FuncCount    int
	EventCount   int
	TriggerCount int
}
