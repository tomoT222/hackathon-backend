package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
    // Load Env (Minimal approach, assuming env vars are set in shell)
    user := os.Getenv("MYSQL_USER")
    pwd := os.Getenv("MYSQL_PWD")
    host := os.Getenv("MYSQL_HOST")
    dbName := os.Getenv("MYSQL_DATABASE")

    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pwd, host, dbName)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 1. Add Column
    fmt.Println("Adding initial_price column...")
    _, err = db.Exec("ALTER TABLE items ADD COLUMN initial_price INT DEFAULT 0;")
    if err != nil {
        // Ignore if already exists (safe check)
        fmt.Printf("Warning (might be ok if exists): %v\n", err)
    }

    // 2. Update existing rows
    fmt.Println("Updating existing rows...")
    res, err := db.Exec("UPDATE items SET initial_price = price WHERE initial_price = 0")
    if err != nil {
        log.Fatal(err)
    }
    rows, _ := res.RowsAffected()
    fmt.Printf("Updated %d rows.\n", rows)
    
    fmt.Println("Migration Done.")
}
