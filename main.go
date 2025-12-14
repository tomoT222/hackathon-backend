package main

import (
	"context"
	"database/sql"
	"fmt"
	"hackathon-backend/controller"
	"hackathon-backend/dao"
	"hackathon-backend/pkg/gemini"
	"hackathon-backend/usecase"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
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

	// 2. Gemini Client (API Key)
    apiKey := os.Getenv("GEMINI_API_KEY")
    
	var geminiClient *gemini.Client
	if apiKey != "" {
		ctx := context.Background()
		client, err := gemini.NewClient(ctx, apiKey)
		if err != nil {
			log.Printf("Failed to init Gemini Client: %v", err)
		} else {
			geminiClient = client
			fmt.Println("Gemini Client Initialized!")
		}
	} else {
		fmt.Println("GEMINI_API_KEY not set. Smart-Nego will be disabled.")
	}

	// 3. Dependency Injection
	itemRepo := dao.NewItemRepository(db)
	msgRepo := dao.NewMessageRepository(db)
	
	itemUsecase := usecase.NewItemUsecase(itemRepo, msgRepo, geminiClient)
	itemController := controller.NewItemController(itemUsecase)

	userRepo := dao.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userController := controller.NewUserController(userUsecase)

	// 4. Routing
	http.HandleFunc("/items", itemController.HandleItems)
	http.HandleFunc("/items/", itemController.HandleItemDetail) // Handles /buy and /messages
    http.HandleFunc("/messages/", itemController.HandleMessages) // Handles /messages/{id}/approve
	http.HandleFunc("/register", userController.Register)

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
