package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// DB Connection
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "user"
	}
	pwd := os.Getenv("MYSQL_PWD")
	if pwd == "" {
		pwd = "password"
	}
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "tcp(127.0.0.1:3306)"
	}
	dbName := os.Getenv("MYSQL_DATABASE")
	if dbName == "" {
		dbName = "hackathon_db"
	}

	dsn := fmt.Sprintf("%s:%s@%s/%s?parseTime=true&loc=Local", user, pwd, host, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	fmt.Println("Connected to Database for Migration!")

	// Read Migration File
	content, err := ioutil.ReadFile("db/migration_phase3_5.sql")
	if err != nil {
		log.Fatal(err)
	}

	queries := strings.Split(string(content), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		// Skip comments
        if strings.HasPrefix(query, "--") {
            continue
        }

		log.Printf("Executing: %s", query)
		_, err := db.Exec(query)
		if err != nil {
			// MySQL 8.0 doesn't support IF NOT EXISTS for columns easily in one line without procedure,
			// So we ignore "Duplicate column name" error (Code 1060)
			if strings.Contains(err.Error(), "Duplicate column name") {
				log.Printf("Skipping duplicate column error: %v", err)
			} else {
				log.Printf("Error executing query: %v", err)
			}
		}
	}

	fmt.Println("Migration Phase 3.5 Completed Successfully!")
}
