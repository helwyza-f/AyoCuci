package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	
	"BackendFramework/internal/config"
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/route"
)

func init() {
	
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment")
	}


	log.Println("=== Environment Variables ===")
	log.Println("ENVIRONMENT:", os.Getenv("ENVIRONMENT"))
	log.Println("INFOBIP_SENDER:", os.Getenv("INFOBIP_SENDER"))
	log.Println("INFOBIP_API_KEY:", os.Getenv("INFOBIP_API_KEY")[:20]+"...")
	log.Println("=============================")

	config.InitEnvronment()
	config.InitInfobipVars() 
	config.InitDatabaseVars()
	config.InitEncryptionVars()
	config.InitBucketVars()
	config.InitEmailVars()

	middleware.InitLogger()
	middleware.InitValidator()

	database.OpenAkademik()
	database.OpenAuth()
}

func main() {
	router := route.SetupRouter()
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}