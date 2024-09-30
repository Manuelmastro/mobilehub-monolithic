package main

import (
	"MOBILEHUB/config"
	"MOBILEHUB/route"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	config.Initialize()
	config.AutoMigrate()
	//godotenv.Load("env")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	router := gin.Default()
	route.UrlRouteSet(router)
	router.LoadHTMLGlob("templates/*")
	router.Run("0.0.0.0:8080")
}
