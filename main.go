package main

import (
	"database/sql"
	"fmt"
	"hackathon-backend/controller"
	"hackathon-backend/dao"
	"hackathon-backend/usecase"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. DB Connection
	// user:password@tcp(127.0.0.1:3306)/hackathon_db
	dsn := "user:password@tcp(127.0.0.1:3306)/hackathon_db?parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	fmt.Println("Connected to Database!")

	// 2. Dependency Injection
	itemRepo := dao.NewItemRepository(db)
	itemUsecase := usecase.NewItemUsecase(itemRepo)
	itemController := controller.NewItemController(itemUsecase)

	// 3. Routing
	http.HandleFunc("/items", itemController.HandleItems)
	http.HandleFunc("/items/", itemController.HandleItemDetail)

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
