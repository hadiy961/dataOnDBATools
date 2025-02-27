package utils

import (
	"database/sql"
	"os"
	"time"
)

// ServiceInfo represents information about a service
type ServiceInfo struct {
	ID               int       `json:"id"`
	ServerID         int       `json:"server_id"`
	ServiceName      string    `json:"service_name"`
	ServiceType      string    `json:"service_type"`
	Status           string    `json:"status"`
	Port             int       `json:"port"`
	Version          string    `json:"version"`
	InstallPath      string    `json:"install_path"`
	ConfigFile       string    `json:"config_file"`
	AutoStart        bool      `json:"auto_start"`
	LastChecked      time.Time `json:"last_checked"`
	LastStatusChange time.Time `json:"last_status_change"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Notes            string    `json:"notes"`
}

// ServiceChecker provides methods to check service status
type ServiceChecker struct {
	isSSH    bool
	host     string
	port     int
	user     string
	password string
}

// StorageInfo menyimpan informasi tentang penyimpanan
type StorageInfo struct {
	DeviceName     string
	FilesystemType string
	TotalSpace     string
	UsedSpace      string
	FreeSpace      string
	UsePercent     int
	MountPoint     string
}

// SystemInfo represents system information retrieved from a server
type SystemInfo struct {
	Hostname  string
	TotalRAM  string
	CPUCore   string
	CPUModel  string
	OSType    string
	OSVersion string
}

// FileInfo menyimpan informasi detail tentang file
type FileInfo struct {
	Exists      bool
	Size        int64
	Permissions string
	IsDir       bool
	ModTime     time.Time
	Owner       string
	Group       string
	Path        string
	Extension   string
	MimeType    string
	MD5Hash     string
	SHA256Hash  string
}

// DirectoryInfo menyimpan informasi tentang direktori
type DirectoryInfo struct {
	Path     string
	Exists   bool
	IsDir    bool
	Mode     os.FileMode
	Size     int64
	LastMod  time.Time
	IsEmpty  bool
	Children []string
}

type DirectoryStats struct {
	TotalFiles   int64
	TotalDirs    int64
	TotalSize    int64
	LastModified time.Time
	MaxDepth     int
}

// DirectoryConfig holds configuration for directory creation
type DirectoryConfig struct {
	BasePath      string
	Mode          os.FileMode
	CreateParents bool
	SkipIfExists  bool
}

type Config struct {
	Database struct {
		Host string `yaml:"host"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		Port int    `yaml:"port"`
		Name string `yaml:"name"`
	} `yaml:"database"`
}

type BackupConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Destination string
}

type MySQLDumpConfig struct {
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	User        string   `yaml:"user"`
	Password    string   `yaml:"password"`
	Databases   []string `yaml:"databases"`    // For multiple databases
	ExcludeDBs  []string `yaml:"exclude_dbs"`  // For excluding databases
	Tables      []string `yaml:"tables"`       // For specific tables
	ExcludeTabs []string `yaml:"exclude_tabs"` // For excluding tables
	OutputDir   string   `yaml:"output_dir"`
	StructOnly  bool     `yaml:"struct_only"` // For structure-only dumps
	WithUsers   bool     `yaml:"with_users"`  // Include user grants
}

type MySQLDumper struct {
	config *MySQLDumpConfig
}

// YAMLConfig represents the structure of dbaTools.yaml
type YAMLConfig struct {
	Database struct {
		Host string `yaml:"db_host"`
		User string `yaml:"db_user"`
		Pass string `yaml:"db_pass"`
		Port int    `yaml:"db_port"`
		Name string `yaml:"db_name"`
	} `yaml:"database"`
}

type LogLevel int

type LogOutput uint8

type LoggerConfig struct {
	AppName     string
	AppVersion  string
	Environment string
	DebugMode   bool
	LogPath     string
	DB          *sql.DB
	MinLevel    LogLevel
	Output      LogOutput
}

type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	ProcessID   string                 `json:"process_id"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Module      string                 `json:"module,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Environment string                 `json:"environment"`
	AppName     string                 `json:"app_name"`
	AppVersion  string                 `json:"app_version"`
	Context     map[string]interface{} `json:"context,omitempty"`
}
