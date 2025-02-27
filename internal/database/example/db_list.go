package example

import (
	"dbaTools/internal/database"
	"fmt"
	"log"
)

func ExampleUsageDBList() {
	// Initialize database connection
	config := database.DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "test_db",
	}

	db := database.NewDBConnection(config)
	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// List all databases with their sizes
	fmt.Println("=== Databases and Sizes ===")
	databases, err := db.GetDatabases()
	if err != nil {
		log.Fatal(err)
	}
	for dbName, size := range databases {
		fmt.Printf("Database: %s, Size: %d bytes (%.2f MB)\n",
			dbName, size, float64(size)/1024/1024)
	}

	// List all tables in current database
	fmt.Println("\n=== Tables Information ===")
	tables, err := db.GetTablesList()
	if err != nil {
		log.Fatal(err)
	}
	for _, table := range tables {
		fmt.Printf("Table: %s\n", table.Name)
		fmt.Printf("  Engine: %s\n", table.Engine)
		fmt.Printf("  Rows: %d\n", table.Rows)
		fmt.Printf("  Size: %d bytes (%.2f MB)\n\n",
			table.Size, float64(table.Size)/1024/1024)
	}

	// List all views
	fmt.Println("=== Views ===")
	views, err := db.GetViewsList()
	if err != nil {
		log.Fatal(err)
	}
	for name, definition := range views {
		fmt.Printf("View: %s\n", name)
		fmt.Printf("Definition:\n%s\n\n", definition)
	}

	// List all stored procedures
	fmt.Println("=== Stored Procedures ===")
	procs, err := db.GetProceduresList()
	if err != nil {
		log.Fatal(err)
	}
	for name, definition := range procs {
		fmt.Printf("Procedure: %s\n", name)
		fmt.Printf("Definition:\n%s\n\n", definition)
	}

	// List all functions
	fmt.Println("=== Functions ===")
	funcs, err := db.GetFunctionsList()
	if err != nil {
		log.Fatal(err)
	}
	for name, definition := range funcs {
		fmt.Printf("Function: %s\n", name)
		fmt.Printf("Definition:\n%s\n\n", definition)
	}

	// List all events
	fmt.Println("=== Events ===")
	events, err := db.GetEventsList()
	if err != nil {
		log.Fatal(err)
	}
	for name, definition := range events {
		fmt.Printf("Event: %s\n", name)
		fmt.Printf("Definition:\n%s\n\n", definition)
	}

	// List all triggers
	fmt.Println("=== Triggers ===")
	triggers, err := db.GetTriggersList()
	if err != nil {
		log.Fatal(err)
	}
	for name, definition := range triggers {
		fmt.Printf("Trigger: %s\n", name)
		fmt.Printf("Definition:\n%s\n\n", definition)
	}
}
