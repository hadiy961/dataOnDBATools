package database

import (
	"fmt"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
)

type DatabaseSummary struct {
	TotalDatabases int
	TotalSize      uint64
}

func (d *DBConnection) GetDatabases() (map[string]int64, error) {
	tempConn := &DBConnection{db: d.db, config: d.config}

	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "Scanning databases "
	s.Start()
	defer s.Stop()

	// Get database names
	dbNames := make([]string, 0)
	rows, err := d.db.Query(`
        SELECT schema_name 
        FROM information_schema.schemata 
        WHERE schema_name NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
    `)
	if err != nil {
		return nil, fmt.Errorf("error getting databases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		dbNames = append(dbNames, name)
	}

	type dbSizeResult struct {
		name string
		size int64
		err  error
	}

	var (
		mu             sync.Mutex
		processedCount int32
		totalSize      uint64
		databases      = make(map[string]int64)
		prevDB         = ""
		currentDB      = ""
		nextDB         = ""
	)

	updateDisplay := func(prev string, prevSize int64, current, next string, count int32, total uint64) {
		fmt.Printf("\033[4A\033[2K\rPrevious    : %s %s\n", prev, humanize.Bytes(uint64(prevSize)))
		fmt.Printf("\033[2K\rCalculating : %s\n", current)
		fmt.Printf("\033[2K\rNext        : %s\n", next)
		fmt.Printf("\033[2K\r=================================================\n")
		fmt.Printf("\033[2K\rTotal DB Count: %d/%d | Total Size: %s\n",
			count, len(dbNames), humanize.Bytes(total))
	}

	numWorkers := runtime.NumCPU() * 35
	results := make(chan dbSizeResult, len(dbNames))
	dbChan := make(chan string, len(dbNames))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dbName := range dbChan {
				query := `
                    SELECT COALESCE(SUM(data_length + index_length), 0) as size
                    FROM information_schema.tables 
                    WHERE table_schema = ?
                `
				var dbSize int64
				err := tempConn.db.QueryRow(query, dbName).Scan(&dbSize)

				mu.Lock()
				if err == nil && dbSize > 0 {
					databases[dbName] = dbSize
					atomic.AddUint64(&totalSize, uint64(dbSize))
				}

				prevDB = currentDB
				currentDB = dbName
				if idx := slices.Index(dbNames, dbName); idx < len(dbNames)-1 {
					nextDB = dbNames[idx+1]
				}

				count := atomic.AddInt32(&processedCount, 1)
				updateDisplay(prevDB, dbSize, currentDB, nextDB, count, totalSize)
				mu.Unlock()

				results <- dbSizeResult{name: dbName, size: dbSize, err: err}
			}
		}()
	}

	for _, dbName := range dbNames {
		dbChan <- dbName
	}
	close(dbChan)

	wg.Wait()
	close(results)

	return databases, nil
}

// GetTablesList returns detailed information about all tables
func (d *DBConnection) GetTablesList() ([]TableInfo, error) {
	query := `
        SELECT 
            TABLE_NAME,
            ENGINE,
            TABLE_ROWS,
            DATA_LENGTH + INDEX_LENGTH as SIZE
        FROM information_schema.TABLES 
        WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'`

	rows, err := d.db.Query(query, d.config.Database)
	if err != nil {
		return nil, fmt.Errorf("error getting tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Engine, &table.Rows, &table.Size); err != nil {
			return nil, fmt.Errorf("error scanning table info: %w", err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// GetViewsList returns all views and their definitions
func (d *DBConnection) GetViewsList() (map[string]string, error) {
	query := `
        SELECT 
            TABLE_NAME,
            VIEW_DEFINITION
        FROM information_schema.VIEWS 
        WHERE TABLE_SCHEMA = ?`

	rows, err := d.db.Query(query, d.config.Database)
	if err != nil {
		return nil, fmt.Errorf("error getting views: %w", err)
	}
	defer rows.Close()

	views := make(map[string]string)
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			return nil, fmt.Errorf("error scanning view info: %w", err)
		}
		views[name] = definition
	}
	return views, nil
}

// GetProceduresList returns all stored procedures and their definitions
func (d *DBConnection) GetProceduresList() (map[string]string, error) {
	query := `
        SELECT 
            ROUTINE_NAME,
            ROUTINE_DEFINITION
        FROM information_schema.ROUTINES 
        WHERE ROUTINE_SCHEMA = ? AND ROUTINE_TYPE = 'PROCEDURE'`

	rows, err := d.db.Query(query, d.config.Database)
	if err != nil {
		return nil, fmt.Errorf("error getting procedures: %w", err)
	}
	defer rows.Close()

	procs := make(map[string]string)
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			return nil, fmt.Errorf("error scanning procedure info: %w", err)
		}
		procs[name] = definition
	}
	return procs, nil
}

// GetFunctionsList returns all functions and their definitions
func (d *DBConnection) GetFunctionsList() (map[string]string, error) {
	query := `
        SELECT 
            ROUTINE_NAME,
            ROUTINE_DEFINITION
        FROM information_schema.ROUTINES 
        WHERE ROUTINE_SCHEMA = ? AND ROUTINE_TYPE = 'FUNCTION'`

	rows, err := d.db.Query(query, d.config.Database)
	if err != nil {
		return nil, fmt.Errorf("error getting functions: %w", err)
	}
	defer rows.Close()

	funcs := make(map[string]string)
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			return nil, fmt.Errorf("error scanning function info: %w", err)
		}
		funcs[name] = definition
	}
	return funcs, nil
}

// GetEventsList returns all events and their definitions
func (d *DBConnection) GetEventsList() (map[string]string, error) {
	query := `
        SELECT 
            EVENT_NAME,
            EVENT_DEFINITION
        FROM information_schema.EVENTS 
        WHERE EVENT_SCHEMA = ?`

	rows, err := d.db.Query(query, d.config.Database)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %w", err)
	}
	defer rows.Close()

	events := make(map[string]string)
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			return nil, fmt.Errorf("error scanning event info: %w", err)
		}
		events[name] = definition
	}
	return events, nil
}

// GetTriggersList returns all triggers and their definitions
func (d *DBConnection) GetTriggersList() (map[string]string, error) {
	query := `
        SELECT 
            TRIGGER_NAME,
            ACTION_STATEMENT
        FROM information_schema.TRIGGERS 
        WHERE TRIGGER_SCHEMA = ?`

	rows, err := d.db.Query(query, d.config.Database)
	if err != nil {
		return nil, fmt.Errorf("error getting triggers: %w", err)
	}
	defer rows.Close()

	triggers := make(map[string]string)
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			return nil, fmt.Errorf("error scanning trigger info: %w", err)
		}
		triggers[name] = definition
	}
	return triggers, nil
}
