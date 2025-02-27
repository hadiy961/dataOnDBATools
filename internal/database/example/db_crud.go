package example

import (
	"dbaTools/internal/database"
	"fmt"
	"log"
)

func ExmapleDBCrud() {
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

	// Basic CRUD operations
	// Select example
	rows, err := db.Select("SELECT id, name FROM users WHERE age > ?", 18)
	if err != nil {
		log.Fatal("Select error:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: %d - %s\n", id, name)
	}

	// Insert example
	id, err := db.Insert("INSERT INTO users (name, age) VALUES (?, ?)", "John Doe", 25)
	if err != nil {
		log.Fatal("Insert error:", err)
	}
	fmt.Printf("Inserted user with ID: %d\n", id)

	// Update example
	affected, err := db.Update("UPDATE users SET age = ? WHERE name = ?", 26, "John Doe")
	if err != nil {
		log.Fatal("Update error:", err)
	}
	fmt.Printf("Updated %d rows\n", affected)

	// Delete example
	affected, err = db.Delete("DELETE FROM users WHERE name = ?", "John Doe")
	if err != nil {
		log.Fatal("Delete error:", err)
	}
	fmt.Printf("Deleted %d rows\n", affected)

	// Transaction example
	queries := []string{
		"INSERT INTO users (name, age) VALUES ('Alice', 30)",
		"INSERT INTO users (name, age) VALUES ('Bob', 35)",
		"UPDATE users SET age = 31 WHERE name = 'Alice'",
	}

	if err := db.ExecuteInTransaction(queries); err != nil {
		log.Fatal("Transaction error:", err)
	}
	fmt.Println("Transaction completed successfully")

	// Database creation/deletion example
	if err := db.CreateDatabase(); err != nil {
		log.Fatal("Create database error:", err)
	}
	fmt.Println("Database created")

	if err := db.DropDatabase(); err != nil {
		log.Fatal("Drop database error:", err)
	}
	fmt.Println("Database dropped")
}
