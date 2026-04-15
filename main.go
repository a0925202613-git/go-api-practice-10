package main

import (
	"log"

	"go-api-practice-10/config"
	"go-api-practice-10/database"
	"go-api-practice-10/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Println("no .env found, using defaults")
	}
	if err := database.Connect(); err != nil {
		log.Fatal("database connect failed:", err)
	}

	r := gin.Default()
	routes.Setup(r)

	port := config.Port()
	log.Printf("Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
