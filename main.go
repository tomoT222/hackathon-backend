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
	// 1. DB Connection
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
	fmt.Println("Connected to Database!")

	// 2. Dependency Injection
	itemRepo := dao.NewItemRepository(db)
	itemUsecase := usecase.NewItemUsecase(itemRepo)
	itemController := controller.NewItemController(itemUsecase)

	userRepo := dao.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userController := controller.NewUserController(userUsecase)

	// 3. Routing
	http.HandleFunc("/items", itemController.HandleItems)
	http.HandleFunc("/items/", itemController.HandleItemDetail)
	http.HandleFunc("/register", userController.Register)

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
