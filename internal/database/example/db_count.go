package example

import (
	"dbaTools/internal/database"
	"fmt"
	"log"
)

func ExamapleDBCount() {
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

	// Get basic database info
	charset, collation, err := db.GetDatabaseCharset()
	if err != nil {
		log.Fatal("Error getting charset:", err)
	}

	size, err := db.GetDatabaseSize()
	if err != nil {
		log.Fatal("Error getting size:", err)
	}

	creationTime, err := db.GetDatabaseCreationTime()
	if err != nil {
		log.Fatal("Error getting creation time:", err)
	}

	// Get object counts
	tableCount, _ := db.GetTableCount()
	viewCount, _ := db.GetViewCount()
	procCount, _ := db.GetStoredProcedureCount()
	funcCount, _ := db.GetFunctionCount()
	eventCount, _ := db.GetEventCount()
	triggerCount, _ := db.GetTriggerCount()

	// Print database summary
	fmt.Printf("Database: %s\n", config.Database)
	fmt.Printf("Created: %s\n", creationTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Size: %.2f MB\n", float64(size)/1024/1024)
	fmt.Printf("Charset: %s\n", charset)
	fmt.Printf("Collation: %s\n\n", collation)

	fmt.Println("Object Counts:")
	fmt.Printf("Tables: %d\n", tableCount)
	fmt.Printf("Views: %d\n", viewCount)
	fmt.Printf("Stored Procedures: %d\n", procCount)
	fmt.Printf("Functions: %d\n", funcCount)
	fmt.Printf("Events: %d\n", eventCount)
	fmt.Printf("Triggers: %d\n", triggerCount)
}
