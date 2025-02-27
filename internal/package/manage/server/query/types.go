package query

import "database/sql"

type ServerCredential struct {
	ID          int    `json:"id"`
	ServerID    int    `json:"server_id"`
	AuthID      int    `json:"auth_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	IPAddress   string `json:"ipaddress"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Description string `json:"description"`
}

// ServerHead represents a record from the server_head table
type ServerHead struct {
	ID          int
	Code        string
	Name        string
	IPAddress   string
	Port        int
	Description string
	ZoneID      int
	CreatedAt   sql.NullTime // Using sql.NullTime to handle potential NULL values
	CreatedBy   string
	UpdatedAt   sql.NullTime // Using sql.NullTime to handle potential NULL values
	UpdatedBy   string
	CategoryID  int
	Type        string
}
