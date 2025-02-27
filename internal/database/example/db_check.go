package example

import (
	"dbaTools/internal/database"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func ExampleUsageDBCheck() {
	// Initialize config
	config := database.DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "test_db",
	}

	// Create connection
	db := database.NewDBConnection(config)

	// Test basic connectivity
	if err := db.Connect(); err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer db.Close()

	// Check credentials
	if err := db.CheckCredentials(); err != nil {
		log.Fatal("Invalid credentials:", err)
	}

	// Check if database exists
	exists, err := db.CheckDatabaseExists()
	if err != nil {
		log.Fatal("Error checking database:", err)
	}

	if !exists {
		fmt.Println("Database does not exist")
		return
	}

	// Check active connection
	if err := db.CheckConnection(); err != nil {
		log.Fatal("Connection check failed:", err)
	}

	fmt.Println("Successfully connected to database")
}
