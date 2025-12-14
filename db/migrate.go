package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	user := "user"
	pwd := "password"
	host := "tcp(127.0.0.1:3306)"
	dbName := "hackathon_db"

	if os.Getenv("MYSQL_USER") != "" { user = os.Getenv("MYSQL_USER") }
	if os.Getenv("MYSQL_PWD") != "" { pwd = os.Getenv("MYSQL_PWD") }
	if os.Getenv("MYSQL_HOST") != "" { host = os.Getenv("MYSQL_HOST") }
	if os.Getenv("MYSQL_DATABASE") != "" { dbName = os.Getenv("MYSQL_DATABASE") }

	dsn := fmt.Sprintf("%s:%s@%s/%s?parseTime=true&loc=Local", user, pwd, host, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	queries := []string{
		"ALTER TABLE items ADD COLUMN views_count INT DEFAULT 0",
		"ALTER TABLE items ADD COLUMN ai_negotiation_enabled BOOLEAN DEFAULT FALSE",
		"ALTER TABLE items ADD COLUMN min_price INT",
		`CREATE TABLE IF NOT EXISTS messages (
			id CHAR(26) PRIMARY KEY COMMENT 'ULID',
			item_id CHAR(26) NOT NULL,
			sender_id CHAR(26) NOT NULL,
			content TEXT NOT NULL,
			is_ai_response BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS negotiation_logs (
			id CHAR(26) PRIMARY KEY COMMENT 'ULID',
			item_id CHAR(26) NOT NULL,
			user_id CHAR(26) NOT NULL COMMENT 'Buyer ID',
			proposed_price INT NOT NULL,
			ai_decision VARCHAR(50) NOT NULL COMMENT 'ACCEPT, REJECT, COUNTER',
			ai_reasoning TEXT,
			log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Printf("Error executing query: %s\nError: %v\n", q, err)
            // Continue even if error (e.g. column exists if syntax is slightly different)
		} else {
			fmt.Println("Executed successfully:", q[:20], "...")
		}
	}
    fmt.Println("Migration completed.")
}
